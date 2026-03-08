package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type fakeStore struct {
	attempts            int
	runningCalled       bool
	succeededCalled     bool
	failedCalled        bool
	retryableFailCalled bool
	lastSyncStatus      string
}

func (f *fakeStore) GetJobAttempts(context.Context, uuid.UUID) (int, error) {
	return f.attempts, nil
}
func (f *fakeStore) MarkJobRunning(context.Context, uuid.UUID, int) error {
	f.runningCalled = true
	return nil
}
func (f *fakeStore) MarkJobSucceeded(context.Context, uuid.UUID, int) error {
	f.succeededCalled = true
	return nil
}
func (f *fakeStore) MarkJobRetryableFailure(context.Context, uuid.UUID, int, string) error {
	f.retryableFailCalled = true
	return nil
}
func (f *fakeStore) MarkJobFailed(context.Context, uuid.UUID, int, string) error {
	f.failedCalled = true
	return nil
}
func (f *fakeStore) UpdateRepositorySyncStatus(_ context.Context, _ uuid.UUID, status string, _ *string, _ *time.Time) error {
	f.lastSyncStatus = status
	return nil
}

type fakeProcessor struct {
	err error
}

func (f *fakeProcessor) Run(context.Context, jobs.RepoInitialSyncPayload) error {
	return f.err
}

func makeTask(t *testing.T) *asynq.Task {
	t.Helper()
	payload, err := json.Marshal(jobs.TaskEnvelope[jobs.RepoInitialSyncPayload]{
		JobID: uuid.New(),
		Payload: jobs.RepoInitialSyncPayload{
			RepositoryID:      uuid.New(),
			OrganizationID:    uuid.New(),
			TriggeredByUserID: uuid.New(),
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return asynq.NewTask(jobs.TaskRepoInitialSync, payload)
}

func TestRepoInitialSyncSuccessLifecycle(t *testing.T) {
	store := &fakeStore{}
	handler := NewRepoInitialSyncHandler(store, &fakeProcessor{})

	if err := handler.Handle(context.Background(), makeTask(t)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !store.runningCalled || !store.succeededCalled {
		t.Fatalf("expected running and succeeded transitions, got running=%v succeeded=%v", store.runningCalled, store.succeededCalled)
	}
}

func TestRepoInitialSyncFailurePersistsError(t *testing.T) {
	store := &fakeStore{}
	handler := NewRepoInitialSyncHandler(store, &fakeProcessor{err: errors.New("sync failed")})

	if err := handler.Handle(context.Background(), makeTask(t)); err == nil {
		t.Fatal("expected handler error")
	}
	if !store.retryableFailCalled && !store.failedCalled {
		t.Fatal("expected failed transition to be persisted")
	}
}
