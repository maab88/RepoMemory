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

type fakeGenerateMemoryStore struct {
	attempts            int
	runningCalled       bool
	succeededCalled     bool
	failedCalled        bool
	retryableFailCalled bool
}

func (f *fakeGenerateMemoryStore) GetJobAttempts(context.Context, uuid.UUID) (int, error) {
	return f.attempts, nil
}
func (f *fakeGenerateMemoryStore) MarkJobRunning(context.Context, uuid.UUID, int) error {
	f.runningCalled = true
	return nil
}
func (f *fakeGenerateMemoryStore) MarkJobSucceeded(context.Context, uuid.UUID, int) error {
	f.succeededCalled = true
	return nil
}
func (f *fakeGenerateMemoryStore) MarkJobRetryableFailure(context.Context, uuid.UUID, int, string) error {
	f.retryableFailCalled = true
	return nil
}
func (f *fakeGenerateMemoryStore) MarkJobFailed(context.Context, uuid.UUID, int, string) error {
	f.failedCalled = true
	return nil
}

type fakeGenerateMemoryService struct {
	err error
}

func (f *fakeGenerateMemoryService) GenerateAndPersistForRepository(context.Context, jobs.RepoGenerateMemoryPayload) (services.MemoryGenerationResult, error) {
	if f.err != nil {
		return services.MemoryGenerationResult{}, f.err
	}
	return services.MemoryGenerationResult{}, nil
}

func makeGenerateMemoryTask(t *testing.T) *asynq.Task {
	t.Helper()
	payload, err := json.Marshal(jobs.TaskEnvelope[jobs.RepoGenerateMemoryPayload]{
		JobID: uuid.New(),
		Payload: jobs.RepoGenerateMemoryPayload{
			RepositoryID:      uuid.New(),
			OrganizationID:    uuid.New(),
			TriggeredByUserID: uuid.New(),
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return asynq.NewTask(jobs.TaskRepoGenerateMemory, payload)
}

func TestGenerateMemoryHandlerSuccessLifecycle(t *testing.T) {
	store := &fakeGenerateMemoryStore{}
	handler := NewGenerateMemoryHandler(store, &fakeGenerateMemoryService{})

	if err := handler.Handle(context.Background(), makeGenerateMemoryTask(t)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !store.runningCalled || !store.succeededCalled {
		t.Fatalf("expected running and succeeded transitions, got running=%v succeeded=%v", store.runningCalled, store.succeededCalled)
	}
}

func TestGenerateMemoryHandlerFailureLifecycle(t *testing.T) {
	store := &fakeGenerateMemoryStore{}
	handler := NewGenerateMemoryHandler(store, &fakeGenerateMemoryService{err: errors.New("memory generation failed")})

	if err := handler.Handle(context.Background(), makeGenerateMemoryTask(t)); err == nil {
		t.Fatal("expected handler error")
	}
	if !store.retryableFailCalled && !store.failedCalled {
		t.Fatal("expected failed transition to be persisted")
	}
}
