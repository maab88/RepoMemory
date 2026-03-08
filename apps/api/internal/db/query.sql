-- name: CreateOrganization :one
INSERT INTO organizations (
  name,
  slug
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: ListOrganizationsForUser :many
SELECT o.*
FROM organizations o
INNER JOIN memberships m ON m.organization_id = o.id
WHERE m.user_id = $1
ORDER BY o.created_at DESC;

-- name: GetOrganizationByID :one
SELECT *
FROM organizations
WHERE id = $1;

-- name: CreateMembership :one
INSERT INTO memberships (
  organization_id,
  user_id,
  role
) VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: UpsertRepository :one
INSERT INTO repositories (
  organization_id,
  github_repo_id,
  owner_login,
  name,
  full_name,
  private,
  default_branch,
  html_url,
  description,
  is_active,
  imported_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  NOW(),
  NOW()
)
ON CONFLICT (organization_id, github_repo_id)
DO UPDATE SET
  owner_login = EXCLUDED.owner_login,
  name = EXCLUDED.name,
  full_name = EXCLUDED.full_name,
  private = EXCLUDED.private,
  default_branch = EXCLUDED.default_branch,
  html_url = EXCLUDED.html_url,
  description = EXCLUDED.description,
  is_active = EXCLUDED.is_active,
  updated_at = NOW()
RETURNING *;

-- name: ListRepositoriesForOrganization :many
SELECT *
FROM repositories
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: GetRepositoryByID :one
SELECT *
FROM repositories
WHERE id = $1;

-- name: GetRepositorySyncStateByRepositoryID :one
SELECT *
FROM repository_sync_states
WHERE repository_id = $1;

-- name: ListRepositorySummariesForOrganization :many
SELECT
  r.id,
  r.organization_id,
  r.github_repo_id,
  r.owner_login,
  r.name,
  r.full_name,
  r.private,
  r.default_branch,
  r.html_url,
  r.description,
  r.imported_at,
  rss.last_sync_status,
  rss.last_successful_sync_at AS last_sync_time,
  (
    SELECT COUNT(*)
    FROM pull_requests pr
    WHERE pr.repository_id = r.id
  )::INT AS pull_request_count,
  (
    SELECT COUNT(*)
    FROM issues i
    WHERE i.repository_id = r.id
  )::INT AS issue_count,
  (
    SELECT COUNT(*)
    FROM memory_entries me
    WHERE me.repository_id = r.id
  )::INT AS memory_entry_count
FROM repositories r
LEFT JOIN repository_sync_states rss ON rss.repository_id = r.id
WHERE r.organization_id = $1
ORDER BY r.imported_at DESC, r.created_at DESC;

-- name: ListRepositorySummariesForUser :many
SELECT
  r.id,
  r.organization_id,
  r.github_repo_id,
  r.owner_login,
  r.name,
  r.full_name,
  r.private,
  r.default_branch,
  r.html_url,
  r.description,
  r.imported_at,
  rss.last_sync_status,
  rss.last_successful_sync_at AS last_sync_time,
  (
    SELECT COUNT(*)
    FROM pull_requests pr
    WHERE pr.repository_id = r.id
  )::INT AS pull_request_count,
  (
    SELECT COUNT(*)
    FROM issues i
    WHERE i.repository_id = r.id
  )::INT AS issue_count,
  (
    SELECT COUNT(*)
    FROM memory_entries me
    WHERE me.repository_id = r.id
  )::INT AS memory_entry_count
FROM repositories r
INNER JOIN memberships m ON m.organization_id = r.organization_id
LEFT JOIN repository_sync_states rss ON rss.repository_id = r.id
WHERE m.user_id = $1
ORDER BY r.imported_at DESC, r.created_at DESC;

-- name: UpsertRepositorySyncState :one
INSERT INTO repository_sync_states (
  repository_id,
  last_pr_sync_at,
  last_issue_sync_at,
  last_successful_sync_at,
  last_sync_status,
  last_sync_error,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  NOW()
)
ON CONFLICT (repository_id)
DO UPDATE SET
  last_pr_sync_at = EXCLUDED.last_pr_sync_at,
  last_issue_sync_at = EXCLUDED.last_issue_sync_at,
  last_successful_sync_at = EXCLUDED.last_successful_sync_at,
  last_sync_status = EXCLUDED.last_sync_status,
  last_sync_error = EXCLUDED.last_sync_error,
  updated_at = NOW()
RETURNING *;

-- name: UpsertPullRequest :one
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
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  NOW(),
  NOW()
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
  synced_at = NOW(),
  updated_at = NOW()
RETURNING *;

-- name: UpsertIssue :one
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
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  NOW(),
  NOW()
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
  synced_at = NOW(),
  updated_at = NOW()
RETURNING *;

-- name: InsertMemoryEntry :one
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
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  NOW()
)
RETURNING *;

-- name: LinkMemoryEntrySource :one
INSERT INTO memory_entry_sources (
  memory_entry_id,
  source_type,
  source_record_id
) VALUES (
  $1,
  $2,
  $3
)
ON CONFLICT (memory_entry_id, source_type, source_record_id)
DO UPDATE SET source_record_id = EXCLUDED.source_record_id
RETURNING *;

-- name: ListMemoryEntriesForRepository :many
SELECT *
FROM memory_entries
WHERE repository_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListMemoryEntriesByRepository :many
SELECT *
FROM memory_entries
WHERE repository_id = $1
ORDER BY created_at DESC;

-- name: GetMemoryEntryByIDAndRepository :one
SELECT *
FROM memory_entries
WHERE id = $1
  AND repository_id = $2;

-- name: ListMemoryEntrySourcesByMemoryEntryID :many
SELECT
  mes.source_type,
  mes.source_record_id,
  CASE
    WHEN mes.source_type = 'pull_request' THEN pr.html_url
    WHEN mes.source_type = 'issue' THEN i.html_url
    ELSE ''
  END AS source_url
FROM memory_entry_sources mes
LEFT JOIN pull_requests pr
  ON mes.source_type = 'pull_request'
 AND mes.source_record_id = pr.id
LEFT JOIN issues i
  ON mes.source_type = 'issue'
 AND mes.source_record_id = i.id
WHERE mes.memory_entry_id = $1
ORDER BY mes.created_at ASC;

-- name: SearchMemoryEntries :many
SELECT
  me.id,
  me.repository_id,
  r.name AS repository_name,
  me.type,
  me.title,
  me.summary,
  me.source_url,
  me.created_at,
  COUNT(*) OVER()::INT AS total_count
FROM memory_entries me
INNER JOIN repositories r ON r.id = me.repository_id
WHERE me.organization_id = sqlc.arg(organization_id)
  AND (
    sqlc.narg(repository_id)::uuid IS NULL
    OR me.repository_id = sqlc.narg(repository_id)::uuid
  )
  AND (
    me.title ILIKE ('%' || sqlc.arg(query_text) || '%')
    OR me.summary ILIKE ('%' || sqlc.arg(query_text) || '%')
  )
ORDER BY
  CASE
    WHEN me.title ILIKE ('%' || sqlc.arg(query_text) || '%') THEN 0
    ELSE 1
  END ASC,
  me.created_at DESC,
  me.id DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: InsertDigest :one
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
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  NOW()
)
RETURNING *;

-- name: ListDigestsForRepository :many
SELECT *
FROM digests
WHERE repository_id = $1
ORDER BY period_start DESC;

-- name: InsertJob :one
INSERT INTO jobs (
  organization_id,
  repository_id,
  job_type,
  status,
  queue_name,
  attempts,
  last_error,
  payload,
  started_at,
  finished_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  NOW()
)
RETURNING *;

-- name: GetJobByID :one
SELECT *
FROM jobs
WHERE id = $1;

-- name: UpdateJobStatus :one
UPDATE jobs
SET
  status = $2,
  attempts = $3,
  last_error = $4,
  started_at = $5,
  finished_at = $6,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: InsertAuditLog :one
INSERT INTO audit_logs (
  organization_id,
  actor_user_id,
  action,
  entity_type,
  entity_id,
  metadata
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
)
RETURNING *;
-- name: UpsertUserByID :one
INSERT INTO users (
  id,
  email,
  display_name,
  avatar_url,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  NOW()
)
ON CONFLICT (id)
DO UPDATE SET
  email = COALESCE(EXCLUDED.email, users.email),
  display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), users.display_name),
  avatar_url = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
  updated_at = NOW()
RETURNING *;

-- name: ListOrganizationsWithRoleForUser :many
SELECT
  o.id,
  o.name,
  o.slug,
  o.created_at,
  o.updated_at,
  m.role
FROM organizations o
INNER JOIN memberships m ON m.organization_id = o.id
WHERE m.user_id = $1
ORDER BY o.created_at DESC;

-- name: GetOrganizationForUser :one
SELECT
  o.id,
  o.name,
  o.slug,
  o.created_at,
  o.updated_at,
  m.role
FROM organizations o
INNER JOIN memberships m ON m.organization_id = o.id
WHERE o.id = $1 AND m.user_id = $2;

-- name: UserHasMembership :one
SELECT EXISTS(
  SELECT 1
  FROM memberships
  WHERE user_id = $1 AND organization_id = $2
);

-- name: UpsertGithubAccount :one
INSERT INTO github_accounts (
  user_id,
  github_user_id,
  github_login,
  access_token_encrypted,
  token_scope,
  connected_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  NOW(),
  NOW()
)
ON CONFLICT (user_id, github_user_id)
DO UPDATE SET
  github_login = EXCLUDED.github_login,
  access_token_encrypted = EXCLUDED.access_token_encrypted,
  token_scope = EXCLUDED.token_scope,
  connected_at = NOW(),
  updated_at = NOW()
RETURNING *;

-- name: GetLatestGithubAccountForUser :one
SELECT *
FROM github_accounts
WHERE user_id = $1
ORDER BY connected_at DESC
LIMIT 1;
