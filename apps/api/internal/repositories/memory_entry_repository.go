package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type MemoryEntryRepository struct {
	queries *db.Queries
}

func NewMemoryEntryRepository(queries *db.Queries) *MemoryEntryRepository {
	return &MemoryEntryRepository{queries: queries}
}

func (r *MemoryEntryRepository) ListByRepository(ctx context.Context, repositoryID uuid.UUID) ([]db.MemoryEntry, error) {
	return r.queries.ListMemoryEntriesByRepository(ctx, repositoryID)
}

func (r *MemoryEntryRepository) GetByIDAndRepository(ctx context.Context, memoryID, repositoryID uuid.UUID) (db.MemoryEntry, error) {
	return r.queries.GetMemoryEntryByIDAndRepository(ctx, db.GetMemoryEntryByIDAndRepositoryParams{
		ID:           memoryID,
		RepositoryID: repositoryID,
	})
}
