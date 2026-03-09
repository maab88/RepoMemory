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

type PullRequestForMemory struct {
	ID             uuid.UUID
	RepositoryID   uuid.UUID
	GitHubPrNumber int32
	Title          string
	Body           string
	State          string
	AuthorLogin    string
	HTMLURL        string
	MergedAt       *time.Time
	ClosedAt       *time.Time
	Labels         []string
}

type IssueForMemory struct {
	ID                uuid.UUID
	RepositoryID      uuid.UUID
	GitHubIssueNumber int32
	Title             string
	Body              string
	State             string
	AuthorLogin       string
	HTMLURL           string
	ClosedAt          *time.Time
	Labels            []string
}

type MemoryEntryUpsertRecord struct {
	OrganizationID uuid.UUID
	RepositoryID   uuid.UUID
	Type           string
	Title          string
	Summary        string
	WhyItMatters   string
	ImpactedAreas  []string
	Risks          []string
	FollowUps      []string
	SourceKind     string
	SourceID       uuid.UUID
	SourceURL      string
	GeneratedBy    string
	SourceType     string
	SourceRecordID uuid.UUID
}

type PullRequestForDigest struct {
	ID             uuid.UUID
	GitHubPrNumber int32
	Title          string
	State          string
	AuthorLogin    string
	HTMLURL        string
	MergedAt       *time.Time
}

type IssueForDigest struct {
	ID                uuid.UUID
	GitHubIssueNumber int32
	Title             string
	State             string
	AuthorLogin       string
	HTMLURL           string
	UpdatedAtExternal time.Time
}

type MemoryEntryForDigest struct {
	ID            uuid.UUID
	Type          string
	Title         string
	Summary       string
	WhyItMatters  string
	ImpactedAreas []string
	FollowUps     []string
	CreatedAt     time.Time
}

type DigestUpsertRecord struct {
	OrganizationID uuid.UUID
	RepositoryID   uuid.UUID
	PeriodStart    time.Time
	PeriodEnd      time.Time
	Title          string
	Summary        string
	BodyMarkdown   string
	GeneratedBy    string
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

func (s *Store) ListPullRequestsForRepository(ctx context.Context, repositoryID uuid.UUID) ([]PullRequestForMemory, error) {
	const query = `
SELECT id, repository_id, github_pr_number, title, COALESCE(body, ''), state, COALESCE(author_login, ''), html_url, merged_at, closed_at, labels
FROM pull_requests
WHERE repository_id = $1
ORDER BY updated_at_external DESC, github_pr_number DESC`

	rows, err := s.pool.Query(ctx, query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]PullRequestForMemory, 0)
	for rows.Next() {
		var record PullRequestForMemory
		var labelsRaw []byte
		if err := rows.Scan(
			&record.ID,
			&record.RepositoryID,
			&record.GitHubPrNumber,
			&record.Title,
			&record.Body,
			&record.State,
			&record.AuthorLogin,
			&record.HTMLURL,
			&record.MergedAt,
			&record.ClosedAt,
			&labelsRaw,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(labelsRaw, &record.Labels); err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

func (s *Store) ListIssuesForRepository(ctx context.Context, repositoryID uuid.UUID) ([]IssueForMemory, error) {
	const query = `
SELECT id, repository_id, github_issue_number, title, COALESCE(body, ''), state, COALESCE(author_login, ''), html_url, closed_at, labels
FROM issues
WHERE repository_id = $1
ORDER BY updated_at_external DESC, github_issue_number DESC`

	rows, err := s.pool.Query(ctx, query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]IssueForMemory, 0)
	for rows.Next() {
		var record IssueForMemory
		var labelsRaw []byte
		if err := rows.Scan(
			&record.ID,
			&record.RepositoryID,
			&record.GitHubIssueNumber,
			&record.Title,
			&record.Body,
			&record.State,
			&record.AuthorLogin,
			&record.HTMLURL,
			&record.ClosedAt,
			&labelsRaw,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(labelsRaw, &record.Labels); err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

func (s *Store) ListRecentPullRequestsForHotspots(ctx context.Context, repositoryID uuid.UUID, since time.Time) ([]PullRequestForHotspot, error) {
	const query = `
SELECT id, repository_id, github_pr_number, title, COALESCE(body, ''), state, html_url, labels, updated_at_external
FROM pull_requests
WHERE repository_id = $1
  AND updated_at_external >= $2
ORDER BY updated_at_external DESC, github_pr_number DESC`

	rows, err := s.pool.Query(ctx, query, repositoryID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]PullRequestForHotspot, 0)
	for rows.Next() {
		var record PullRequestForHotspot
		var labelsRaw []byte
		if err := rows.Scan(
			&record.ID,
			&record.RepositoryID,
			&record.GitHubPrNumber,
			&record.Title,
			&record.Body,
			&record.State,
			&record.HTMLURL,
			&labelsRaw,
			&record.UpdatedAtExternal,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(labelsRaw, &record.Labels); err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

func (s *Store) ListRecentIssuesForHotspots(ctx context.Context, repositoryID uuid.UUID, since time.Time) ([]IssueForHotspot, error) {
	const query = `
SELECT id, repository_id, github_issue_number, title, COALESCE(body, ''), state, html_url, labels, updated_at_external
FROM issues
WHERE repository_id = $1
  AND updated_at_external >= $2
ORDER BY updated_at_external DESC, github_issue_number DESC`

	rows, err := s.pool.Query(ctx, query, repositoryID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]IssueForHotspot, 0)
	for rows.Next() {
		var record IssueForHotspot
		var labelsRaw []byte
		if err := rows.Scan(
			&record.ID,
			&record.RepositoryID,
			&record.GitHubIssueNumber,
			&record.Title,
			&record.Body,
			&record.State,
			&record.HTMLURL,
			&labelsRaw,
			&record.UpdatedAtExternal,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(labelsRaw, &record.Labels); err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

func (s *Store) UpsertMemoryEntryForSource(ctx context.Context, record MemoryEntryUpsertRecord) (uuid.UUID, bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, false, err
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	const findQuery = `
SELECT me.id
FROM memory_entries me
INNER JOIN memory_entry_sources mes ON mes.memory_entry_id = me.id
WHERE me.repository_id = $1
  AND me.type = $2
  AND mes.source_type = $3
  AND mes.source_record_id = $4
LIMIT 1`

	var existingID uuid.UUID
	findErr := tx.QueryRow(ctx, findQuery, record.RepositoryID, record.Type, record.SourceType, record.SourceRecordID).Scan(&existingID)
	created := false

	impactedAreas, err := json.Marshal(record.ImpactedAreas)
	if err != nil {
		return uuid.Nil, false, err
	}
	risks, err := json.Marshal(record.Risks)
	if err != nil {
		return uuid.Nil, false, err
	}
	followUps, err := json.Marshal(record.FollowUps)
	if err != nil {
		return uuid.Nil, false, err
	}

	if findErr == nil {
		const updateQuery = `
UPDATE memory_entries
SET
  title = $2,
  summary = $3,
  why_it_matters = $4,
  impacted_areas = $5,
  risks = $6,
  follow_ups = $7,
  source_kind = $8,
  source_id = $9,
  source_url = $10,
  generated_by = $11,
  updated_at = NOW()
WHERE id = $1`
		if _, err := tx.Exec(ctx, updateQuery,
			existingID,
			record.Title,
			record.Summary,
			nullIfEmpty(record.WhyItMatters),
			impactedAreas,
			risks,
			followUps,
			nullIfEmpty(record.SourceKind),
			record.SourceID,
			nullIfEmpty(record.SourceURL),
			record.GeneratedBy,
		); err != nil {
			return uuid.Nil, false, err
		}
	} else {
		if !errors.Is(findErr, pgx.ErrNoRows) {
			return uuid.Nil, false, findErr
		}
		const insertQuery = `
INSERT INTO memory_entries (
  organization_id,
  repository_id,
  type,
  title,
  summary,
  why_it_matters,
  impacted_areas,
  risks,
  follow_ups,
  source_kind,
  source_id,
  source_url,
  generated_by,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW()
)
RETURNING id`
		if err := tx.QueryRow(ctx, insertQuery,
			record.OrganizationID,
			record.RepositoryID,
			record.Type,
			record.Title,
			record.Summary,
			nullIfEmpty(record.WhyItMatters),
			impactedAreas,
			risks,
			followUps,
			nullIfEmpty(record.SourceKind),
			record.SourceID,
			nullIfEmpty(record.SourceURL),
			record.GeneratedBy,
		).Scan(&existingID); err != nil {
			return uuid.Nil, false, err
		}
		created = true
	}

	const linkQuery = `
INSERT INTO memory_entry_sources (memory_entry_id, source_type, source_record_id)
VALUES ($1, $2, $3)
ON CONFLICT (memory_entry_id, source_type, source_record_id)
DO NOTHING`
	if _, err := tx.Exec(ctx, linkQuery, existingID, record.SourceType, record.SourceRecordID); err != nil {
		return uuid.Nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, false, err
	}

	return existingID, created, nil
}

func (s *Store) UpsertHotspotMemoryEntry(ctx context.Context, record HotspotMemoryUpsertRecord) (uuid.UUID, bool, error) {
	const findQuery = `
SELECT id
FROM memory_entries
WHERE repository_id = $1
  AND type = 'hotspot'
  AND source_kind = $2
LIMIT 1`

	var entryID uuid.UUID
	findErr := s.pool.QueryRow(ctx, findQuery, record.RepositoryID, record.HotspotKey).Scan(&entryID)
	created := errors.Is(findErr, pgx.ErrNoRows)
	if findErr != nil && !created {
		return uuid.Nil, false, findErr
	}

	impactedAreas, err := json.Marshal(record.ImpactedAreas)
	if err != nil {
		return uuid.Nil, false, err
	}
	risks, err := json.Marshal(record.Risks)
	if err != nil {
		return uuid.Nil, false, err
	}
	followUps, err := json.Marshal(record.FollowUps)
	if err != nil {
		return uuid.Nil, false, err
	}

	if created {
		const insertQuery = `
INSERT INTO memory_entries (
  organization_id,
  repository_id,
  type,
  title,
  summary,
  why_it_matters,
  impacted_areas,
  risks,
  follow_ups,
  source_kind,
  source_id,
  source_url,
  generated_by,
  updated_at
) VALUES (
  $1, $2, 'hotspot', $3, $4, $5, $6, $7, $8, $9, NULL, $10, $11, NOW()
)
RETURNING id`
		if err := s.pool.QueryRow(ctx, insertQuery,
			record.OrganizationID,
			record.RepositoryID,
			record.Title,
			record.Summary,
			nullIfEmpty(record.WhyItMatters),
			impactedAreas,
			risks,
			followUps,
			record.HotspotKey,
			nullIfEmpty(record.SourceURL),
			record.GeneratedBy,
		).Scan(&entryID); err != nil {
			return uuid.Nil, false, err
		}
		return entryID, true, nil
	}

	const updateQuery = `
UPDATE memory_entries
SET
  title = $2,
  summary = $3,
  why_it_matters = $4,
  impacted_areas = $5,
  risks = $6,
  follow_ups = $7,
  source_url = $8,
  generated_by = $9,
  updated_at = NOW()
WHERE id = $1`
	if _, err := s.pool.Exec(ctx, updateQuery,
		entryID,
		record.Title,
		record.Summary,
		nullIfEmpty(record.WhyItMatters),
		impactedAreas,
		risks,
		followUps,
		nullIfEmpty(record.SourceURL),
		record.GeneratedBy,
	); err != nil {
		return uuid.Nil, false, err
	}

	return entryID, false, nil
}

func (s *Store) LinkMemoryEntrySource(ctx context.Context, memoryEntryID uuid.UUID, sourceType string, sourceRecordID uuid.UUID) error {
	const query = `
INSERT INTO memory_entry_sources (memory_entry_id, source_type, source_record_id)
VALUES ($1, $2, $3)
ON CONFLICT (memory_entry_id, source_type, source_record_id)
DO NOTHING`

	_, err := s.pool.Exec(ctx, query, memoryEntryID, sourceType, sourceRecordID)
	return err
}

func (s *Store) ListPullRequestsForDigestPeriod(ctx context.Context, repositoryID uuid.UUID, periodStart, periodEnd time.Time) ([]PullRequestForDigest, error) {
	const query = `
SELECT id, github_pr_number, title, state, COALESCE(author_login, ''), html_url, merged_at
FROM pull_requests
WHERE repository_id = $1
  AND merged_at IS NOT NULL
  AND merged_at >= $2
  AND merged_at <= $3
ORDER BY merged_at DESC, github_pr_number DESC`

	rows, err := s.pool.Query(ctx, query, repositoryID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]PullRequestForDigest, 0)
	for rows.Next() {
		var item PullRequestForDigest
		if err := rows.Scan(
			&item.ID,
			&item.GitHubPrNumber,
			&item.Title,
			&item.State,
			&item.AuthorLogin,
			&item.HTMLURL,
			&item.MergedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) ListOpenIssuesForDigestPeriod(ctx context.Context, repositoryID uuid.UUID, periodEnd time.Time) ([]IssueForDigest, error) {
	const query = `
SELECT id, github_issue_number, title, state, COALESCE(author_login, ''), html_url, updated_at_external
FROM issues
WHERE repository_id = $1
  AND state = 'open'
  AND updated_at_external <= $2
ORDER BY updated_at_external DESC, github_issue_number DESC
LIMIT 20`

	rows, err := s.pool.Query(ctx, query, repositoryID, periodEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]IssueForDigest, 0)
	for rows.Next() {
		var item IssueForDigest
		if err := rows.Scan(
			&item.ID,
			&item.GitHubIssueNumber,
			&item.Title,
			&item.State,
			&item.AuthorLogin,
			&item.HTMLURL,
			&item.UpdatedAtExternal,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) ListMemoryEntriesForDigestPeriod(ctx context.Context, repositoryID uuid.UUID, periodStart, periodEnd time.Time) ([]MemoryEntryForDigest, error) {
	const query = `
SELECT id, type, title, summary, COALESCE(why_it_matters, ''), impacted_areas, follow_ups, created_at
FROM memory_entries
WHERE repository_id = $1
  AND created_at >= $2
  AND created_at <= $3
ORDER BY created_at DESC
LIMIT 100`

	rows, err := s.pool.Query(ctx, query, repositoryID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MemoryEntryForDigest, 0)
	for rows.Next() {
		var item MemoryEntryForDigest
		var impactedRaw []byte
		var followUpsRaw []byte
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Title,
			&item.Summary,
			&item.WhyItMatters,
			&impactedRaw,
			&followUpsRaw,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(impactedRaw, &item.ImpactedAreas); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(followUpsRaw, &item.FollowUps); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) UpsertDigestForPeriod(ctx context.Context, record DigestUpsertRecord) (uuid.UUID, bool, error) {
	const findQuery = `
SELECT id
FROM digests
WHERE repository_id = $1
  AND period_start = $2
  AND period_end = $3
LIMIT 1`

	var existingID uuid.UUID
	findErr := s.pool.QueryRow(ctx, findQuery, record.RepositoryID, record.PeriodStart, record.PeriodEnd).Scan(&existingID)
	created := errors.Is(findErr, pgx.ErrNoRows)
	if findErr != nil && !created {
		return uuid.Nil, false, findErr
	}

	const upsertQuery = `
INSERT INTO digests (
  organization_id,
  repository_id,
  period_start,
  period_end,
  title,
  summary,
  body_markdown,
  generated_by,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, NOW()
)
ON CONFLICT (repository_id, period_start, period_end)
DO UPDATE SET
  title = EXCLUDED.title,
  summary = EXCLUDED.summary,
  body_markdown = EXCLUDED.body_markdown,
  generated_by = EXCLUDED.generated_by,
  updated_at = NOW()
RETURNING id`

	var digestID uuid.UUID
	if err := s.pool.QueryRow(ctx, upsertQuery,
		record.OrganizationID,
		record.RepositoryID,
		record.PeriodStart,
		record.PeriodEnd,
		record.Title,
		record.Summary,
		record.BodyMarkdown,
		record.GeneratedBy,
	).Scan(&digestID); err != nil {
		return uuid.Nil, false, err
	}

	return digestID, created, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
