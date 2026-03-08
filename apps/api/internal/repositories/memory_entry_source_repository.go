package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type MemoryEntrySourceRepository struct {
	queries *db.Queries
}

func NewMemoryEntrySourceRepository(queries *db.Queries) *MemoryEntrySourceRepository {
	return &MemoryEntrySourceRepository{queries: queries}
}

func (r *MemoryEntrySourceRepository) ListByMemoryEntryID(ctx context.Context, memoryEntryID uuid.UUID) ([]db.ListMemoryEntrySourcesByMemoryEntryIDRow, error) {
	return r.queries.ListMemoryEntrySourcesByMemoryEntryID(ctx, memoryEntryID)
}
