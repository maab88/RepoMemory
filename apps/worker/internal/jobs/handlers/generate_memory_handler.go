package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
	"github.com/maab88/repomemory/apps/worker/internal/services"
)

type GenerateMemoryStore interface {
	GetJobAttempts(ctx context.Context, jobID uuid.UUID) (int, error)
	MarkJobRunning(ctx context.Context, jobID uuid.UUID, attempts int) error
	MarkJobSucceeded(ctx context.Context, jobID uuid.UUID, attempts int) error
	MarkJobRetryableFailure(ctx context.Context, jobID uuid.UUID, attempts int, lastError string) error
	MarkJobFailed(ctx context.Context, jobID uuid.UUID, attempts int, lastError string) error
}

type GenerateMemoryService interface {
	GenerateAndPersistForRepository(ctx context.Context, payload jobs.RepoGenerateMemoryPayload) (services.MemoryGenerationResult, error)
}

type GenerateMemoryHandler struct {
	store   GenerateMemoryStore
	service GenerateMemoryService
}

func NewGenerateMemoryHandler(store GenerateMemoryStore, service GenerateMemoryService) *GenerateMemoryHandler {
	return &GenerateMemoryHandler{
		store:   store,
		service: service,
	}
}

func (h *GenerateMemoryHandler) Handle(ctx context.Context, task *asynq.Task) error {
	envelope, err := jobs.ParseTaskEnvelope[jobs.RepoGenerateMemoryPayload](task.Payload())
	if err != nil {
		return fmt.Errorf("parse repo generate memory payload: %w", err)
	}

	attempts, err := h.store.GetJobAttempts(ctx, envelope.JobID)
	if err != nil {
		return fmt.Errorf("get job attempts: %w", err)
	}
	attempts++

	if err := h.store.MarkJobRunning(ctx, envelope.JobID, attempts); err != nil {
		return fmt.Errorf("mark job running: %w", err)
	}

	if _, err := h.service.GenerateAndPersistForRepository(ctx, envelope.Payload); err != nil {
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
