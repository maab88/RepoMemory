package handlers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type GenerateDigestHandler struct{}

func NewGenerateDigestHandler() *GenerateDigestHandler {
	return &GenerateDigestHandler{}
}

func (h *GenerateDigestHandler) Handle(_ context.Context, task *asynq.Task) error {
	if _, err := jobs.ParseTaskEnvelope[jobs.RepoGenerateDigestPayload](task.Payload()); err != nil {
		return fmt.Errorf("parse repo generate digest payload: %w", err)
	}
	return nil
}
