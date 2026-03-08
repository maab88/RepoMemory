package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type PullRequestSyncRecord struct {
	RepositoryID      uuid.UUID
	GitHubPrID        int64
	GitHubPrNumber    int32
	Title             string
	Body              string
	State             string
	AuthorLogin       string
	HTMLURL           string
	MergedAt          *time.Time
	ClosedAt          *time.Time
	Labels            []byte
	CreatedAtExternal time.Time
	UpdatedAtExternal time.Time
}

type PullRequestRepository struct {
	queries *db.Queries
}

func NewPullRequestRepository(queries *db.Queries) *PullRequestRepository {
	return &PullRequestRepository{queries: queries}
}

func (r *PullRequestRepository) Upsert(ctx context.Context, record PullRequestSyncRecord) error {
	_, err := r.queries.UpsertPullRequest(ctx, db.UpsertPullRequestParams{
		RepositoryID:      record.RepositoryID,
		GithubPrID:        record.GitHubPrID,
		GithubPrNumber:    record.GitHubPrNumber,
		Title:             record.Title,
		Body:              optionalText(record.Body),
		State:             record.State,
		AuthorLogin:       optionalText(record.AuthorLogin),
		HtmlUrl:           record.HTMLURL,
		MergedAt:          optionalTime(record.MergedAt),
		ClosedAt:          optionalTime(record.ClosedAt),
		Labels:            record.Labels,
		CreatedAtExternal: pgtype.Timestamptz{Time: record.CreatedAtExternal.UTC(), Valid: true},
		UpdatedAtExternal: pgtype.Timestamptz{Time: record.UpdatedAtExternal.UTC(), Valid: true},
	})
	return err
}

func optionalText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func optionalTime(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}
