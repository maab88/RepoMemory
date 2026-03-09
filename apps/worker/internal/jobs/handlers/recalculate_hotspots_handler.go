package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
	"github.com/maab88/repomemory/apps/worker/internal/services/hotspots"
)

type RecalculateHotspotsStore interface {
	GetJobAttempts(ctx context.Context, jobID uuid.UUID) (int, error)
	MarkJobRunning(ctx context.Context, jobID uuid.UUID, attempts int) error
	MarkJobSucceeded(ctx context.Context, jobID uuid.UUID, attempts int) error
	MarkJobRetryableFailure(ctx context.Context, jobID uuid.UUID, attempts int, lastError string) error
	MarkJobFailed(ctx context.Context, jobID uuid.UUID, attempts int, lastError string) error
}

type RecalculateHotspotsService interface {
	RecalculateForRepository(ctx context.Context, payload jobs.RepoRecalculateHotspotsPayload) (hotspots.RecalculationResult, error)
}

type RecalculateHotspotsHandler struct {
	store   RecalculateHotspotsStore
	service RecalculateHotspotsService
}

func NewRecalculateHotspotsHandler(store RecalculateHotspotsStore, service RecalculateHotspotsService) *RecalculateHotspotsHandler {
	return &RecalculateHotspotsHandler{
		store:   store,
		service: service,
	}
}

func (h *RecalculateHotspotsHandler) Handle(ctx context.Context, task *asynq.Task) error {
	envelope, err := jobs.ParseTaskEnvelope[jobs.RepoRecalculateHotspotsPayload](task.Payload())
	if err != nil {
		return fmt.Errorf("parse repo recalculate hotspots payload: %w", err)
	}

	attempts, err := h.store.GetJobAttempts(ctx, envelope.JobID)
	if err != nil {
		return fmt.Errorf("get job attempts: %w", err)
	}
	attempts++

	if err := h.store.MarkJobRunning(ctx, envelope.JobID, attempts); err != nil {
		return fmt.Errorf("mark job running: %w", err)
	}

	if _, err := h.service.RecalculateForRepository(ctx, envelope.Payload); err != nil {
		lastErr := err.Error()
		retryCount, okRetry := asynq.GetRetryCount(ctx)
		maxRetry, okMax := asynq.GetMaxRetry(ctx)
		exhausted := okRetry && okMax && retryCount >= maxRetry

		if exhausted {
			_ = h.store.MarkJobFailed(ctx, envelope.JobID, attempts, lastErr)
		} else {
			_ = h.store.MarkJobRetryableFailure(ctx, envelope.JobID, attempts, lastErr)
		}
		return err
	}

	if err := h.store.MarkJobSucceeded(ctx, envelope.JobID, attempts); err != nil {
		return fmt.Errorf("mark job succeeded: %w", err)
	}

	return nil
}
