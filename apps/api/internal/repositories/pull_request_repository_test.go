package repositories

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
	"github.com/maab88/repomemory/apps/api/internal/db"
)

func TestPullRequestUpsertIsIdempotent(t *testing.T) {
	ctx := context.Background()
	conn := openTestConn(t, ctx)
	defer conn.Close(ctx)

	queries := db.New(conn)
	repositoryID := seedRepository(t, ctx, conn, queries)
	repo := NewPullRequestRepository(queries)

	now := time.Now().UTC()
	record := PullRequestSyncRecord{
		RepositoryID:      repositoryID,
		GitHubPrID:        101,
		GitHubPrNumber:    7,
		Title:             "First title",
		Body:              "body",
		State:             "open",
		AuthorLogin:       "octocat",
		HTMLURL:           "https://github.com/octo/repo/pull/7",
		Labels:            []byte(`["sync"]`),
		CreatedAtExternal: now.Add(-time.Hour),
		UpdatedAtExternal: now,
	}
	if err := repo.Upsert(ctx, record); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	record.Title = "Updated title"
	if err := repo.Upsert(ctx, record); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	var count int
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM pull_requests WHERE repository_id = $1 AND github_pr_id = $2`, repositoryID, int64(101)).Scan(&count); err != nil {
		t.Fatalf("count pull requests: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 pull request row, got %d", count)
	}
}

func openTestConn(t *testing.T, ctx context.Context) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(ctx, testDatabaseURL())
	if err != nil {
		t.Skipf("skipping DB integration test: %v", err)
	}

	schemaName := fmt.Sprintf("it_repo_pr_%d", time.Now().UnixNano())
	if _, err := conn.Exec(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = conn.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE")
	})
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
	return conn
}

func seedRepository(t *testing.T, ctx context.Context, conn *pgx.Conn, queries *db.Queries) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	if _, err := conn.Exec(ctx, `INSERT INTO users (id, display_name) VALUES ($1, 'Owner')`, userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	org, err := queries.CreateOrganization(ctx, db.CreateOrganizationParams{Name: "Acme", Slug: "acme"})
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if _, err := queries.CreateMembership(ctx, db.CreateMembershipParams{OrganizationID: org.ID, UserID: userID, Role: "owner"}); err != nil {
		t.Fatalf("create membership: %v", err)
	}
	repo, err := queries.UpsertRepository(ctx, db.UpsertRepositoryParams{
		OrganizationID: org.ID,
		GithubRepoID:   1,
		OwnerLogin:     "octo",
		Name:           "repo",
		FullName:       "octo/repo",
		Private:        true,
		DefaultBranch:  "main",
		HtmlUrl:        "https://github.com/octo/repo",
		Description:    pgtype.Text{},
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}
	return repo.ID
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
