package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type RepoInitialSyncStore interface {
	GetJobAttempts(ctx context.Context, jobID uuid.UUID) (int, error)
	MarkJobRunning(ctx context.Context, jobID uuid.UUID, attempts int) error
	MarkJobSucceeded(ctx context.Context, jobID uuid.UUID, attempts int) error
	MarkJobRetryableFailure(ctx context.Context, jobID uuid.UUID, attempts int, lastError string) error
	MarkJobFailed(ctx context.Context, jobID uuid.UUID, attempts int, lastError string) error
	UpdateRepositorySyncStatus(ctx context.Context, repositoryID uuid.UUID, status string, lastError *string, lastSuccessfulSyncAt *time.Time) error
}

type RepoInitialSyncService interface {
	Run(ctx context.Context, payload jobs.RepoInitialSyncPayload) error
}

type RepoInitialSyncHandler struct {
	store   RepoInitialSyncStore
	service RepoInitialSyncService
}

func NewRepoInitialSyncHandler(store RepoInitialSyncStore, service RepoInitialSyncService) *RepoInitialSyncHandler {
	return &RepoInitialSyncHandler{
		store:   store,
		service: service,
	}
}

func (h *RepoInitialSyncHandler) Handle(ctx context.Context, task *asynq.Task) error {
	envelope, err := jobs.ParseTaskEnvelope[jobs.RepoInitialSyncPayload](task.Payload())
	if err != nil {
		return fmt.Errorf("parse repo initial sync payload: %w", err)
	}

	attempts, err := h.store.GetJobAttempts(ctx, envelope.JobID)
	if err != nil {
		return fmt.Errorf("get job attempts: %w", err)
	}
	attempts++

	if err := h.store.MarkJobRunning(ctx, envelope.JobID, attempts); err != nil {
		return fmt.Errorf("mark job running: %w", err)
	}
	_ = h.store.UpdateRepositorySyncStatus(ctx, envelope.Payload.RepositoryID, jobs.StatusRunning, nil, nil)

	if err := h.service.Run(ctx, envelope.Payload); err != nil {
		lastErr := err.Error()
		retryCount, okRetry := asynq.GetRetryCount(ctx)
		maxRetry, okMax := asynq.GetMaxRetry(ctx)
		exhausted := okRetry && okMax && retryCount >= maxRetry

		_ = h.store.UpdateRepositorySyncStatus(ctx, envelope.Payload.RepositoryID, jobs.StatusFailed, &lastErr, nil)
		if exhausted {
			_ = h.store.MarkJobFailed(ctx, envelope.JobID, attempts, lastErr)
		} else {
			_ = h.store.MarkJobRetryableFailure(ctx, envelope.JobID, attempts, lastErr)
		}
		return err
	}

	now := time.Now().UTC()
	if err := h.store.MarkJobSucceeded(ctx, envelope.JobID, attempts); err != nil {
		return fmt.Errorf("mark job succeeded: %w", err)
	}
	if err := h.store.UpdateRepositorySyncStatus(ctx, envelope.Payload.RepositoryID, jobs.StatusSucceeded, nil, &now); err != nil {
		return fmt.Errorf("update repository sync state: %w", err)
	}

	return nil
}

type DefaultRepoInitialSyncService struct{}

func (s *DefaultRepoInitialSyncService) Run(_ context.Context, _ jobs.RepoInitialSyncPayload) error {
	return nil
}
