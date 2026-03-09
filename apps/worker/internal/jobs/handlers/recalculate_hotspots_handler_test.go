package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
	"github.com/maab88/repomemory/apps/worker/internal/services/hotspots"
)

type fakeRecalculateHotspotsStore struct {
	attempts            int
	runningCalled       bool
	succeededCalled     bool
	failedCalled        bool
	retryableFailCalled bool
}

func (f *fakeRecalculateHotspotsStore) GetJobAttempts(context.Context, uuid.UUID) (int, error) {
	return f.attempts, nil
}
func (f *fakeRecalculateHotspotsStore) MarkJobRunning(context.Context, uuid.UUID, int) error {
	f.runningCalled = true
	return nil
}
func (f *fakeRecalculateHotspotsStore) MarkJobSucceeded(context.Context, uuid.UUID, int) error {
	f.succeededCalled = true
	return nil
}
func (f *fakeRecalculateHotspotsStore) MarkJobRetryableFailure(context.Context, uuid.UUID, int, string) error {
	f.retryableFailCalled = true
	return nil
}
func (f *fakeRecalculateHotspotsStore) MarkJobFailed(context.Context, uuid.UUID, int, string) error {
	f.failedCalled = true
	return nil
}

type fakeRecalculateHotspotsService struct {
	err error
}

func (f *fakeRecalculateHotspotsService) RecalculateForRepository(context.Context, jobs.RepoRecalculateHotspotsPayload) (hotspots.RecalculationResult, error) {
	if f.err != nil {
		return hotspots.RecalculationResult{}, f.err
	}
	return hotspots.RecalculationResult{}, nil
}

func makeRecalculateHotspotsTask(t *testing.T) *asynq.Task {
	t.Helper()
	payload, err := json.Marshal(jobs.TaskEnvelope[jobs.RepoRecalculateHotspotsPayload]{
		JobID: uuid.New(),
		Payload: jobs.RepoRecalculateHotspotsPayload{
			RepositoryID:      uuid.New(),
			OrganizationID:    uuid.New(),
			TriggeredByUserID: uuid.New(),
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return asynq.NewTask(jobs.TaskRepoRecalculateHotspots, payload)
}

func TestRecalculateHotspotsHandlerSuccessLifecycle(t *testing.T) {
	store := &fakeRecalculateHotspotsStore{}
	handler := NewRecalculateHotspotsHandler(store, &fakeRecalculateHotspotsService{})

	if err := handler.Handle(context.Background(), makeRecalculateHotspotsTask(t)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !store.runningCalled || !store.succeededCalled {
		t.Fatalf("expected running and succeeded transitions, got running=%v succeeded=%v", store.runningCalled, store.succeededCalled)
	}
}

func TestRecalculateHotspotsHandlerFailureLifecycle(t *testing.T) {
	store := &fakeRecalculateHotspotsStore{}
	handler := NewRecalculateHotspotsHandler(store, &fakeRecalculateHotspotsService{err: errors.New("hotspot calculation failed")})

	if err := handler.Handle(context.Background(), makeRecalculateHotspotsTask(t)); err == nil {
		t.Fatal("expected handler error")
	}
	if !store.retryableFailCalled && !store.failedCalled {
		t.Fatal("expected failed transition to be persisted")
	}
}
