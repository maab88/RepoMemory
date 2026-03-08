package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type MemorySearchRepository struct {
	queries *db.Queries
}

func NewMemorySearchRepository(queries *db.Queries) *MemorySearchRepository {
	return &MemorySearchRepository{queries: queries}
}

type MemorySearchInput struct {
	OrganizationID uuid.UUID
	RepositoryID   *uuid.UUID
	Query          string
	Limit          int32
	Offset         int32
}

type MemorySearchResultRow struct {
	ID             uuid.UUID
	RepositoryID   uuid.UUID
	RepositoryName string
	Type           string
	Title          string
	Summary        string
	SourceURL      string
	CreatedAt      time.Time
	TotalCount     int32
}

func (r *MemorySearchRepository) Search(ctx context.Context, input MemorySearchInput) ([]MemorySearchResultRow, error) {
	rows, err := r.queries.SearchMemoryEntries(ctx, db.SearchMemoryEntriesParams{
		OrganizationID: input.OrganizationID,
		RepositoryID:   input.RepositoryID,
		QueryText:      pgtype.Text{String: input.Query, Valid: true},
		LimitCount:     input.Limit,
		OffsetCount:    input.Offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]MemorySearchResultRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, MemorySearchResultRow{
			ID:             row.ID,
			RepositoryID:   row.RepositoryID,
			RepositoryName: row.RepositoryName,
			Type:           row.Type,
			Title:          row.Title,
			Summary:        row.Summary,
			SourceURL:      row.SourceUrl.String,
			CreatedAt:      row.CreatedAt.Time.UTC(),
			TotalCount:     row.TotalCount,
		})
	}

	return result, nil
}
