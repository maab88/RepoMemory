package db_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/maab88/repomemory/apps/api/internal/db"
)

func TestCoreQueryFlows(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, testDatabaseURL())
	if err != nil {
		t.Skipf("skipping DB integration test: %v", err)
	}
	defer func() {
		_ = conn.Close(ctx)
	}()

	schemaName := fmt.Sprintf("it_%d", time.Now().UnixNano())
	if _, err := conn.Exec(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE")
	}()

	if _, err := conn.Exec(ctx, "SET search_path TO "+schemaName); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	migrationSQL, err := loadMigrationSQL()
	if err != nil {
		t.Fatalf("load migration: %v", err)
	}
	if _, err := conn.Exec(ctx, migrationSQL); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	queries := db.New(conn)

	userID := uuid.UUID{}
	if err := conn.QueryRow(ctx,
		`INSERT INTO users (email, display_name) VALUES ($1, $2) RETURNING id`,
		"owner@example.com",
		"Owner",
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	org, err := queries.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Acme",
		Slug: "acme",
	})
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}

	membership, err := queries.CreateMembership(ctx, db.CreateMembershipParams{
		OrganizationID: org.ID,
		UserID:         userID,
		Role:           "owner",
	})
	if err != nil {
		t.Fatalf("create membership: %v", err)
	}
	if membership.OrganizationID != org.ID {
		t.Fatalf("unexpected membership org id: %s", membership.OrganizationID)
	}

	orgs, err := queries.ListOrganizationsForUser(ctx, userID)
	if err != nil {
		t.Fatalf("list orgs for user: %v", err)
	}
	if len(orgs) != 1 {
		t.Fatalf("expected 1 organization, got %d", len(orgs))
	}

	repo, err := queries.UpsertRepository(ctx, db.UpsertRepositoryParams{
		OrganizationID: org.ID,
		GithubRepoID:   12345,
		OwnerLogin:     "acme",
		Name:           "repomemory",
		FullName:       "acme/repomemory",
		Private:        true,
		DefaultBranch:  "main",
		HtmlUrl:        "https://github.com/acme/repomemory",
		Description:    pgtype.Text{String: "RepoMemory repo", Valid: true},
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("upsert repository: %v", err)
	}

	now := time.Now().UTC()
	pr, err := queries.UpsertPullRequest(ctx, db.UpsertPullRequestParams{
		RepositoryID:      repo.ID,
		GithubPrID:        777,
		GithubPrNumber:    12,
		Title:             "Improve sync pipeline",
		Body:              pgtype.Text{String: "details", Valid: true},
		State:             "open",
		AuthorLogin:       pgtype.Text{String: "octocat", Valid: true},
		HtmlUrl:           "https://github.com/acme/repomemory/pull/12",
		MergedAt:          pgtype.Timestamptz{},
		ClosedAt:          pgtype.Timestamptz{},
		Labels:            []byte(`["sync"]`),
		CreatedAtExternal: pgtype.Timestamptz{Time: now.Add(-2 * time.Hour), Valid: true},
		UpdatedAtExternal: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("upsert pull request: %v", err)
	}

	issue, err := queries.UpsertIssue(ctx, db.UpsertIssueParams{
		RepositoryID:      repo.ID,
		GithubIssueID:     888,
		GithubIssueNumber: 34,
		Title:             "Sync state visibility",
		Body:              pgtype.Text{String: "issue body", Valid: true},
		State:             "open",
		AuthorLogin:       pgtype.Text{String: "hubot", Valid: true},
		HtmlUrl:           "https://github.com/acme/repomemory/issues/34",
		ClosedAt:          pgtype.Timestamptz{},
		Labels:            []byte(`["observability"]`),
		CreatedAtExternal: pgtype.Timestamptz{Time: now.Add(-3 * time.Hour), Valid: true},
		UpdatedAtExternal: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("upsert issue: %v", err)
	}

	entry, err := queries.InsertMemoryEntry(ctx, db.InsertMemoryEntryParams{
		OrganizationID: org.ID,
		RepositoryID:   repo.ID,
		Type:           "weekly_hotspot",
		Title:          "Sync pipeline pressure",
		Summary:        "Multiple sync retries observed in the last 24h.",
		WhyItMatters:   pgtype.Text{String: "Can delay memory freshness.", Valid: true},
		ImpactedAreas:  []byte(`["sync","worker"]`),
		Risks:          []byte(`["stale-memory"]`),
		FollowUps:      []byte(`["add-metrics"]`),
		SourceKind:     pgtype.Text{String: "pull_request", Valid: true},
		SourceID:       &pr.ID,
		SourceUrl:      pgtype.Text{String: pr.HtmlUrl, Valid: true},
		GeneratedBy:    "ai",
	})
	if err != nil {
		t.Fatalf("insert memory entry: %v", err)
	}

	if _, err := queries.LinkMemoryEntrySource(ctx, db.LinkMemoryEntrySourceParams{
		MemoryEntryID:  entry.ID,
		SourceType:     "pull_request",
		SourceRecordID: pr.ID,
	}); err != nil {
		t.Fatalf("link memory source pr: %v", err)
	}

	if _, err := queries.LinkMemoryEntrySource(ctx, db.LinkMemoryEntrySourceParams{
		MemoryEntryID:  entry.ID,
		SourceType:     "issue",
		SourceRecordID: issue.ID,
	}); err != nil {
		t.Fatalf("link memory source issue: %v", err)
	}

	entries, err := queries.ListMemoryEntriesForRepository(ctx, db.ListMemoryEntriesForRepositoryParams{
		RepositoryID: repo.ID,
		Limit:        20,
		Offset:       0,
	})
	if err != nil {
		t.Fatalf("list memory entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 memory entry, got %d", len(entries))
	}
}

func testDatabaseURL() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5432/repomemory?sslmode=disable"
}

func loadMigrationSQL() (string, error) {
	candidates := []string{
		"../../../../infra/migrations/0001_v1_schema.up.sql",
		"../../../infra/migrations/0001_v1_schema.up.sql",
		"infra/migrations/0001_v1_schema.up.sql",
	}

	for _, candidate := range candidates {
		b, err := os.ReadFile(filepath.Clean(candidate))
		if err == nil {
			return string(b), nil
		}
	}

	return "", fmt.Errorf("migration file not found")
}
