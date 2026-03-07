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
