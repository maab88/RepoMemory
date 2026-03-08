package jobs

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUpsertMemoryEntryForSourceIsIdempotent(t *testing.T) {
	ctx := context.Background()
	pool := openWorkerTestPool(t, ctx)
	defer pool.Close()

	store := NewStore(pool)

	orgID := uuid.New()
	repoID := uuid.New()
	prID := uuid.New()

	if _, err := pool.Exec(ctx, `INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)`, orgID, "Acme", fmt.Sprintf("acme-%d", time.Now().UnixNano())); err != nil {
		t.Fatalf("insert organization: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO repositories (id, organization_id, github_repo_id, owner_login, name, full_name, private, default_branch, html_url, is_active) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		repoID, orgID, int64(101), "octo", "repo", "octo/repo", true, "main", "https://github.com/octo/repo", true); err != nil {
		t.Fatalf("insert repository: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO pull_requests (id, repository_id, github_pr_id, github_pr_number, title, state, html_url, labels, created_at_external, updated_at_external) VALUES ($1,$2,$3,$4,$5,$6,$7,'[]'::jsonb,NOW(),NOW())`,
		prID, repoID, int64(777), int32(12), "Initial PR", "open", "https://github.com/octo/repo/pull/12"); err != nil {
		t.Fatalf("insert pull request: %v", err)
	}

	record := MemoryEntryUpsertRecord{
		OrganizationID: orgID,
		RepositoryID:   repoID,
		Type:           "pr_summary",
		Title:          "PR #12: Initial PR",
		Summary:        "Initial summary",
		WhyItMatters:   "Important",
		ImpactedAreas:  []string{"sync"},
		Risks:          []string{"Async risk"},
		FollowUps:      []string{"Monitor"},
		SourceKind:     "pull_request",
		SourceID:       prID,
		SourceURL:      "https://github.com/octo/repo/pull/12",
		GeneratedBy:    "deterministic",
		SourceType:     "pull_request",
		SourceRecordID: prID,
	}

	entryID, created, err := store.UpsertMemoryEntryForSource(ctx, record)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if !created {
		t.Fatal("expected first upsert to create row")
	}

	record.Title = "PR #12: Updated title"
	record.Summary = "Updated summary"
	entryID2, created2, err := store.UpsertMemoryEntryForSource(ctx, record)
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if created2 {
		t.Fatal("expected second upsert to update existing row")
	}
	if entryID != entryID2 {
		t.Fatalf("expected same entry id, got %s and %s", entryID, entryID2)
	}

	var entryCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM memory_entries WHERE repository_id = $1 AND type = 'pr_summary'`, repoID).Scan(&entryCount); err != nil {
		t.Fatalf("count memory entries: %v", err)
	}
	if entryCount != 1 {
		t.Fatalf("expected 1 memory entry, got %d", entryCount)
	}

	var sourceCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM memory_entry_sources WHERE memory_entry_id = $1 AND source_type = 'pull_request' AND source_record_id = $2`, entryID, prID).Scan(&sourceCount); err != nil {
		t.Fatalf("count source rows: %v", err)
	}
	if sourceCount != 1 {
		t.Fatalf("expected 1 source row, got %d", sourceCount)
	}
}

func openWorkerTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	baseURL := workerTestDatabaseURL()
	adminConn, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Skipf("skipping DB integration test: %v", err)
	}

	schemaName := fmt.Sprintf("it_worker_memory_%d", time.Now().UnixNano())
	if _, err := adminConn.Exec(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	if _, err := adminConn.Exec(ctx, "SET search_path TO "+schemaName); err != nil {
		t.Fatalf("set search_path: %v", err)
	}
	migrationSQL, err := loadWorkerMigrationSQL()
	if err != nil {
		t.Fatalf("load migration: %v", err)
	}
	if _, err := adminConn.Exec(ctx, migrationSQL); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(withSearchPath(baseURL, schemaName))
	if err != nil {
		t.Fatalf("parse pool config: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("open test pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		_, _ = adminConn.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE")
		_ = adminConn.Close(context.Background())
	})

	return pool
}

func withSearchPath(rawURL, schema string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	return u.String()
}

func workerTestDatabaseURL() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5432/repomemory?sslmode=disable"
}

func loadWorkerMigrationSQL() (string, error) {
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
