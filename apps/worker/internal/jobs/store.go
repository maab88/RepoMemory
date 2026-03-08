package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

type RepositoryForSync struct {
	ID           uuid.UUID
	Organization uuid.UUID
	OwnerLogin   string
	Name         string
}

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
	Labels            []string
	CreatedAtExternal time.Time
	UpdatedAtExternal time.Time
	SyncedAt          time.Time
}

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
	Labels            []string
	CreatedAtExternal time.Time
	UpdatedAtExternal time.Time
	SyncedAt          time.Time
}

var ErrGitHubAccountNotFound = errors.New("github account not found")

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetRepositoryForSync(ctx context.Context, repositoryID uuid.UUID) (RepositoryForSync, error) {
	const query = `
SELECT id, organization_id, owner_login, name
FROM repositories
WHERE id = $1`

	var row RepositoryForSync
	err := s.pool.QueryRow(ctx, query, repositoryID).Scan(&row.ID, &row.Organization, &row.OwnerLogin, &row.Name)
	if err != nil {
		return RepositoryForSync{}, err
	}
	return row, nil
}

func (s *Store) GetGitHubAccessTokenForUser(ctx context.Context, userID uuid.UUID) (string, error) {
	const query = `
SELECT access_token_encrypted
FROM github_accounts
WHERE user_id = $1
ORDER BY connected_at DESC
LIMIT 1`

	var token string
	err := s.pool.QueryRow(ctx, query, userID).Scan(&token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrGitHubAccountNotFound
		}
		return "", err
	}
	return token, nil
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

func (s *Store) SetRepositoryLastPRSyncAt(ctx context.Context, repositoryID uuid.UUID, at time.Time) error {
	const query = `
INSERT INTO repository_sync_states (
  repository_id,
  last_pr_sync_at,
  updated_at
) VALUES (
  $1,
  $2,
  NOW()
)
ON CONFLICT (repository_id)
DO UPDATE SET
  last_pr_sync_at = EXCLUDED.last_pr_sync_at,
  updated_at = NOW()`
	_, err := s.pool.Exec(ctx, query, repositoryID, at)
	return err
}

func (s *Store) SetRepositoryLastIssueSyncAt(ctx context.Context, repositoryID uuid.UUID, at time.Time) error {
	const query = `
INSERT INTO repository_sync_states (
  repository_id,
  last_issue_sync_at,
  updated_at
) VALUES (
  $1,
  $2,
  NOW()
)
ON CONFLICT (repository_id)
DO UPDATE SET
  last_issue_sync_at = EXCLUDED.last_issue_sync_at,
  updated_at = NOW()`
	_, err := s.pool.Exec(ctx, query, repositoryID, at)
	return err
}

func (s *Store) UpsertPullRequest(ctx context.Context, record PullRequestSyncRecord) error {
	labels, err := json.Marshal(record.Labels)
	if err != nil {
		return err
	}
	const query = `
INSERT INTO pull_requests (
  repository_id,
  github_pr_id,
  github_pr_number,
  title,
  body,
  state,
  author_login,
  html_url,
  merged_at,
  closed_at,
  labels,
  created_at_external,
  updated_at_external,
  synced_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW()
)
ON CONFLICT (repository_id, github_pr_id)
DO UPDATE SET
  github_pr_number = EXCLUDED.github_pr_number,
  title = EXCLUDED.title,
  body = EXCLUDED.body,
  state = EXCLUDED.state,
  author_login = EXCLUDED.author_login,
  html_url = EXCLUDED.html_url,
  merged_at = EXCLUDED.merged_at,
  closed_at = EXCLUDED.closed_at,
  labels = EXCLUDED.labels,
  created_at_external = EXCLUDED.created_at_external,
  updated_at_external = EXCLUDED.updated_at_external,
  synced_at = EXCLUDED.synced_at,
  updated_at = NOW()`
	_, err = s.pool.Exec(ctx, query,
		record.RepositoryID,
		record.GitHubPrID,
		record.GitHubPrNumber,
		record.Title,
		nullIfEmpty(record.Body),
		record.State,
		nullIfEmpty(record.AuthorLogin),
		record.HTMLURL,
		record.MergedAt,
		record.ClosedAt,
		labels,
		record.CreatedAtExternal,
		record.UpdatedAtExternal,
		record.SyncedAt,
	)
	return err
}

func (s *Store) UpsertIssue(ctx context.Context, record IssueSyncRecord) error {
	labels, err := json.Marshal(record.Labels)
	if err != nil {
		return err
	}
	const query = `
INSERT INTO issues (
  repository_id,
  github_issue_id,
  github_issue_number,
  title,
  body,
  state,
  author_login,
  html_url,
  closed_at,
  labels,
  created_at_external,
  updated_at_external,
  synced_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW()
)
ON CONFLICT (repository_id, github_issue_id)
DO UPDATE SET
  github_issue_number = EXCLUDED.github_issue_number,
  title = EXCLUDED.title,
  body = EXCLUDED.body,
  state = EXCLUDED.state,
  author_login = EXCLUDED.author_login,
  html_url = EXCLUDED.html_url,
  closed_at = EXCLUDED.closed_at,
  labels = EXCLUDED.labels,
  created_at_external = EXCLUDED.created_at_external,
  updated_at_external = EXCLUDED.updated_at_external,
  synced_at = EXCLUDED.synced_at,
  updated_at = NOW()`
	_, err = s.pool.Exec(ctx, query,
		record.RepositoryID,
		record.GitHubIssueID,
		record.GitHubIssueNumber,
		record.Title,
		nullIfEmpty(record.Body),
		record.State,
		nullIfEmpty(record.AuthorLogin),
		record.HTMLURL,
		record.ClosedAt,
		labels,
		record.CreatedAtExternal,
		record.UpdatedAtExternal,
		record.SyncedAt,
	)
	return err
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
