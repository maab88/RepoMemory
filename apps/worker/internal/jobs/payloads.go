package jobs

import (
	"encoding/json"

	"github.com/google/uuid"
)

const (
	TaskRepoInitialSync         = "repo.initial_sync"
	TaskRepoIncrementalSync     = "repo.incremental_sync"
	TaskRepoGenerateMemory      = "repo.generate_memory"
	TaskRepoGenerateDigest      = "repo.generate_digest"
	TaskRepoRecalculateHotspots = "repo.recalculate_hotspots"

	QueueDefault = "default"
)

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

type RepoInitialSyncPayload struct {
	RepositoryID      uuid.UUID `json:"repositoryId"`
	OrganizationID    uuid.UUID `json:"organizationId"`
	TriggeredByUserID uuid.UUID `json:"triggeredByUserId"`
}

type RepoIncrementalSyncPayload struct {
	RepositoryID      uuid.UUID `json:"repositoryId"`
	OrganizationID    uuid.UUID `json:"organizationId"`
	TriggeredByUserID uuid.UUID `json:"triggeredByUserId"`
}

type RepoGenerateMemoryPayload struct {
	RepositoryID      uuid.UUID `json:"repositoryId"`
	OrganizationID    uuid.UUID `json:"organizationId"`
	TriggeredByUserID uuid.UUID `json:"triggeredByUserId"`
}

type RepoGenerateDigestPayload struct {
	RepositoryID      uuid.UUID `json:"repositoryId"`
	OrganizationID    uuid.UUID `json:"organizationId"`
	TriggeredByUserID uuid.UUID `json:"triggeredByUserId"`
}

type RepoRecalculateHotspotsPayload struct {
	RepositoryID      uuid.UUID `json:"repositoryId"`
	OrganizationID    uuid.UUID `json:"organizationId"`
	TriggeredByUserID uuid.UUID `json:"triggeredByUserId"`
}

type TaskEnvelope[T any] struct {
	JobID   uuid.UUID `json:"jobId"`
	Payload T         `json:"payload"`
}

func ParseTaskEnvelope[T any](payload []byte) (TaskEnvelope[T], error) {
	var envelope TaskEnvelope[T]
	err := json.Unmarshal(payload, &envelope)
	return envelope, err
}
