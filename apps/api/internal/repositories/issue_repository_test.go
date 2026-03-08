package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/maab88/repomemory/apps/api/internal/db"
)

func TestIssueUpsertIsIdempotent(t *testing.T) {
	ctx := context.Background()
	conn := openTestConn(t, ctx)
	defer conn.Close(ctx)

	queries := db.New(conn)
	repositoryID := seedRepository(t, ctx, conn, queries)
	repo := NewIssueRepository(queries)

	now := time.Now().UTC()
	record := IssueSyncRecord{
		RepositoryID:      repositoryID,
		GitHubIssueID:     202,
		GitHubIssueNumber: 11,
		Title:             "Issue title",
		Body:              "issue body",
		State:             "open",
		AuthorLogin:       "hubot",
		HTMLURL:           "https://github.com/octo/repo/issues/11",
		Labels:            []byte(`["bug"]`),
		CreatedAtExternal: now.Add(-time.Hour),
		UpdatedAtExternal: now,
	}
	if err := repo.Upsert(ctx, record); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	record.Title = "Updated issue title"
	if err := repo.Upsert(ctx, record); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	var count int
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM issues WHERE repository_id = $1 AND github_issue_id = $2`, repositoryID, int64(202)).Scan(&count); err != nil {
		t.Fatalf("count issues: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 issue row, got %d", count)
	}
}
