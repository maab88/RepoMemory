package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type IssueSyncRecord struct {
	RepositoryID      uuid.UUID
	GitHubIssueID     int64
	GitHubIssueNumber int32
	Title             string
	Body              string
	State             string
	AuthorLogin       string
	HTMLURL           string
	ClosedAt          *time.Time
	Labels            []byte
	CreatedAtExternal time.Time
	UpdatedAtExternal time.Time
}

type IssueRepository struct {
	queries *db.Queries
}

func NewIssueRepository(queries *db.Queries) *IssueRepository {
	return &IssueRepository{queries: queries}
}

func (r *IssueRepository) Upsert(ctx context.Context, record IssueSyncRecord) error {
	_, err := r.queries.UpsertIssue(ctx, db.UpsertIssueParams{
		RepositoryID:      record.RepositoryID,
		GithubIssueID:     record.GitHubIssueID,
		GithubIssueNumber: record.GitHubIssueNumber,
		Title:             record.Title,
		Body:              optionalText(record.Body),
		State:             record.State,
		AuthorLogin:       optionalText(record.AuthorLogin),
		HtmlUrl:           record.HTMLURL,
		ClosedAt:          optionalTime(record.ClosedAt),
		Labels:            record.Labels,
		CreatedAtExternal: pgtype.Timestamptz{Time: record.CreatedAtExternal.UTC(), Valid: true},
		UpdatedAtExternal: pgtype.Timestamptz{Time: record.UpdatedAtExternal.UTC(), Valid: true},
	})
	return err
}
