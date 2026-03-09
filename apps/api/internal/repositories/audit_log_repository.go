package repositories

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type CreateAuditLogInput struct {
	OrganizationID *uuid.UUID
	ActorUserID    *uuid.UUID
	Action         string
	EntityType     string
	EntityID       *uuid.UUID
	Metadata       map[string]any
}

type AuditLogRepository struct {
	queries *db.Queries
}

func NewAuditLogRepository(queries *db.Queries) *AuditLogRepository {
	return &AuditLogRepository{queries: queries}
}

func (r *AuditLogRepository) Create(ctx context.Context, input CreateAuditLogInput) error {
	metadata := map[string]any{}
	for k, v := range input.Metadata {
		metadata[k] = v
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = r.queries.InsertAuditLog(ctx, db.InsertAuditLogParams{
		OrganizationID: input.OrganizationID,
		ActorUserID:    input.ActorUserID,
		Action:         input.Action,
		EntityType:     input.EntityType,
		EntityID:       input.EntityID,
		Metadata:       metadataJSON,
	})
	return err
}
