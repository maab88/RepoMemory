package handlers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type RepoIncrementalSyncHandler struct{}

func NewRepoIncrementalSyncHandler() *RepoIncrementalSyncHandler {
	return &RepoIncrementalSyncHandler{}
}

func (h *RepoIncrementalSyncHandler) Handle(_ context.Context, task *asynq.Task) error {
	if _, err := jobs.ParseTaskEnvelope[jobs.RepoIncrementalSyncPayload](task.Payload()); err != nil {
		return fmt.Errorf("parse repo incremental sync payload: %w", err)
	}
	return nil
}
