package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type taskEnvelope struct {
	JobID   uuid.UUID       `json:"jobId"`
	Payload taskPayloadMeta `json:"payload"`
}

type taskPayloadMeta struct {
	RepositoryID      uuid.UUID `json:"repositoryId"`
	OrganizationID    uuid.UUID `json:"organizationId"`
	TriggeredByUserID uuid.UUID `json:"triggeredByUserId"`
}

func WithJobLogging(taskType string, next func(ctx context.Context, task *asynq.Task) error) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		started := time.Now()
		meta := parseEnvelope(task.Payload())
		attempt, attempts := resolveAttemptFields(ctx)

		base := log.With().
			Str("task_type", taskType).
			Str("job_id", meta.jobID).
			Str("repository_id", meta.repositoryID).
			Str("organization_id", meta.organizationID).
			Str("triggered_by_user_id", meta.triggeredByUserID).
			Int("attempt", attempt).
			Int("attempts", attempts).
			Logger()

		base.Info().Msg("worker task started")
		err := next(ctx, task)
		duration := time.Since(started).Milliseconds()
		if err != nil {
			base.Error().Err(err).Str("status", "failed").Int64("duration_ms", duration).Msg("worker task finished")
			return err
		}
		base.Info().Str("status", "succeeded").Int64("duration_ms", duration).Msg("worker task finished")
		return nil
	}
}

type envelopeMeta struct {
	jobID             string
	repositoryID      string
	organizationID    string
	triggeredByUserID string
}

func parseEnvelope(raw []byte) envelopeMeta {
	meta := envelopeMeta{}
	var envelope taskEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return meta
	}

	if envelope.JobID != uuid.Nil {
		meta.jobID = envelope.JobID.String()
	}
	if envelope.Payload.RepositoryID != uuid.Nil {
		meta.repositoryID = envelope.Payload.RepositoryID.String()
	}
	if envelope.Payload.OrganizationID != uuid.Nil {
		meta.organizationID = envelope.Payload.OrganizationID.String()
	}
	if envelope.Payload.TriggeredByUserID != uuid.Nil {
		meta.triggeredByUserID = envelope.Payload.TriggeredByUserID.String()
	}
	return meta
}

func resolveAttemptFields(ctx context.Context) (int, int) {
	retryCount, okRetry := asynq.GetRetryCount(ctx)
	maxRetry, okMax := asynq.GetMaxRetry(ctx)
	if !okRetry || !okMax {
		return 0, 0
	}
	return retryCount + 1, maxRetry + 1
}
