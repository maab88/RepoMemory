package handlers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type GenerateMemoryHandler struct{}

func NewGenerateMemoryHandler() *GenerateMemoryHandler {
	return &GenerateMemoryHandler{}
}

func (h *GenerateMemoryHandler) Handle(_ context.Context, task *asynq.Task) error {
	if _, err := jobs.ParseTaskEnvelope[jobs.RepoGenerateMemoryPayload](task.Payload()); err != nil {
		return fmt.Errorf("parse repo generate memory payload: %w", err)
	}
	return nil
}
