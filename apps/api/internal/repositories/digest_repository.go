package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type DigestRepository struct {
	queries *db.Queries
}

func NewDigestRepository(queries *db.Queries) *DigestRepository {
	return &DigestRepository{queries: queries}
}

func (r *DigestRepository) ListByRepository(ctx context.Context, repositoryID uuid.UUID) ([]db.Digest, error) {
	return r.queries.ListDigestsForRepository(ctx, repositoryID)
}
