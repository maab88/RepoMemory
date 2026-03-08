package org

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/maab88/repomemory/apps/api/internal/db"
)

func TestCreateOrganizationWithOwnerCreatesAuditLog(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	baseURL := testDatabaseURL()
	conn, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Skipf("skipping DB integration test: %v", err)
	}
	defer func() {
		_ = conn.Close(ctx)
	}()

	schemaName := fmt.Sprintf("org_it_%d", time.Now().UnixNano())
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

	userID := uuid.UUID{}
	if err := conn.QueryRow(ctx,
		`INSERT INTO users (email, display_name) VALUES ($1, $2) RETURNING id`,
		"owner@example.com",
		"Owner",
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(baseURL)
	if err != nil {
		t.Fatalf("parse pool config: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	defer pool.Close()

	store := NewStore(pool, db.New(pool))
	orgRow, err := store.CreateOrganizationWithOwner(ctx, userID, "Acme Org", Slugify("Acme Org"))
	if err != nil {
		t.Fatalf("create org with owner: %v", err)
	}

	var count int
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE action = 'organization.created' AND entity_id = $1`, orgRow.ID).Scan(&count); err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 audit log row, got %d", count)
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
