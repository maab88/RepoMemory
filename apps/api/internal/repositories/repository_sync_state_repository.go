package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type RepositorySyncStateRepository struct {
	queries *db.Queries
}

func NewRepositorySyncStateRepository(queries *db.Queries) *RepositorySyncStateRepository {
	return &RepositorySyncStateRepository{queries: queries}
}

func (r *RepositorySyncStateRepository) GetByRepositoryID(ctx context.Context, repositoryID uuid.UUID) (db.RepositorySyncState, error) {
	return r.queries.GetRepositorySyncStateByRepositoryID(ctx, repositoryID)
}
