package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
	"github.com/maab88/repomemory/apps/worker/internal/services"
)

type fakeGenerateDigestStore struct {
	attempts            int
	runningCalled       bool
	succeededCalled     bool
	failedCalled        bool
	retryableFailCalled bool
}

func (f *fakeGenerateDigestStore) GetJobAttempts(context.Context, uuid.UUID) (int, error) {
	return f.attempts, nil
}
func (f *fakeGenerateDigestStore) MarkJobRunning(context.Context, uuid.UUID, int) error {
	f.runningCalled = true
	return nil
}
func (f *fakeGenerateDigestStore) MarkJobSucceeded(context.Context, uuid.UUID, int) error {
	f.succeededCalled = true
	return nil
}
func (f *fakeGenerateDigestStore) MarkJobRetryableFailure(context.Context, uuid.UUID, int, string) error {
	f.retryableFailCalled = true
	return nil
}
func (f *fakeGenerateDigestStore) MarkJobFailed(context.Context, uuid.UUID, int, string) error {
	f.failedCalled = true
	return nil
}

type fakeGenerateDigestService struct {
	err error
}

func (f *fakeGenerateDigestService) GenerateAndPersistForRepository(context.Context, jobs.RepoGenerateDigestPayload) (services.DigestGenerationResult, error) {
	if f.err != nil {
		return services.DigestGenerationResult{}, f.err
	}
	return services.DigestGenerationResult{}, nil
}

func makeGenerateDigestTask(t *testing.T) *asynq.Task {
	t.Helper()
	payload, err := json.Marshal(jobs.TaskEnvelope[jobs.RepoGenerateDigestPayload]{
		JobID: uuid.New(),
		Payload: jobs.RepoGenerateDigestPayload{
			RepositoryID:      uuid.New(),
			OrganizationID:    uuid.New(),
			TriggeredByUserID: uuid.New(),
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return asynq.NewTask(jobs.TaskRepoGenerateDigest, payload)
}

func TestGenerateDigestHandlerSuccessLifecycle(t *testing.T) {
	store := &fakeGenerateDigestStore{}
	handler := NewGenerateDigestHandler(store, &fakeGenerateDigestService{})

	if err := handler.Handle(context.Background(), makeGenerateDigestTask(t)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !store.runningCalled || !store.succeededCalled {
		t.Fatalf("expected running and succeeded transitions, got running=%v succeeded=%v", store.runningCalled, store.succeededCalled)
	}
}

func TestGenerateDigestHandlerFailureLifecycle(t *testing.T) {
	store := &fakeGenerateDigestStore{}
	handler := NewGenerateDigestHandler(store, &fakeGenerateDigestService{err: errors.New("digest generation failed")})

	if err := handler.Handle(context.Background(), makeGenerateDigestTask(t)); err == nil {
		t.Fatal("expected handler error")
	}
	if !store.retryableFailCalled && !store.failedCalled {
		t.Fatal("expected failed transition to be persisted")
	}
}
