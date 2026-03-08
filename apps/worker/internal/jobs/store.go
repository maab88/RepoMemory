package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetJobAttempts(ctx context.Context, jobID uuid.UUID) (int, error) {
	const query = `SELECT attempts FROM jobs WHERE id = $1`
	var attempts int
	err := s.pool.QueryRow(ctx, query, jobID).Scan(&attempts)
	return attempts, err
}

func (s *Store) MarkJobRunning(ctx context.Context, jobID uuid.UUID, attempts int) error {
	const query = `
UPDATE jobs
SET status = $2, attempts = $3, last_error = NULL, started_at = NOW(), finished_at = NULL, updated_at = NOW()
WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, jobID, StatusRunning, attempts)
	return err
}

func (s *Store) MarkJobSucceeded(ctx context.Context, jobID uuid.UUID, attempts int) error {
	const query = `
UPDATE jobs
SET status = $2, attempts = $3, last_error = NULL, finished_at = NOW(), updated_at = NOW()
WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, jobID, StatusSucceeded, attempts)
	return err
}

func (s *Store) MarkJobRetryableFailure(ctx context.Context, jobID uuid.UUID, attempts int, lastError string) error {
	const query = `
UPDATE jobs
SET status = $2, attempts = $3, last_error = $4, finished_at = NULL, updated_at = NOW()
WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, jobID, StatusQueued, attempts, lastError)
	return err
}

func (s *Store) MarkJobFailed(ctx context.Context, jobID uuid.UUID, attempts int, lastError string) error {
	const query = `
UPDATE jobs
SET status = $2, attempts = $3, last_error = $4, finished_at = NOW(), updated_at = NOW()
WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, jobID, StatusFailed, attempts, lastError)
	return err
}

func (s *Store) UpdateRepositorySyncStatus(ctx context.Context, repositoryID uuid.UUID, status string, lastError *string, lastSuccessfulSyncAt *time.Time) error {
	const query = `
INSERT INTO repository_sync_states (
  repository_id,
  last_sync_status,
  last_sync_error,
  last_successful_sync_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  NOW()
)
ON CONFLICT (repository_id)
DO UPDATE SET
  last_sync_status = EXCLUDED.last_sync_status,
  last_sync_error = EXCLUDED.last_sync_error,
  last_successful_sync_at = EXCLUDED.last_successful_sync_at,
  updated_at = NOW()`
	_, err := s.pool.Exec(ctx, query, repositoryID, status, lastError, lastSuccessfulSyncAt)
	return err
}
