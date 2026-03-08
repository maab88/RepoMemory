package handlers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type RecalculateHotspotsHandler struct{}

func NewRecalculateHotspotsHandler() *RecalculateHotspotsHandler {
	return &RecalculateHotspotsHandler{}
}

func (h *RecalculateHotspotsHandler) Handle(_ context.Context, task *asynq.Task) error {
	if _, err := jobs.ParseTaskEnvelope[jobs.RepoRecalculateHotspotsPayload](task.Payload()); err != nil {
		return fmt.Errorf("parse repo recalculate hotspots payload: %w", err)
	}
	return nil
}
