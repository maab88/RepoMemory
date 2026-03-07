-- 0001_init.sql
CREATE TABLE IF NOT EXISTS memory_entries (
  id BIGSERIAL PRIMARY KEY,
  org_id TEXT NOT NULL,
  repo_name TEXT NOT NULL,
  source_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_memory_entries_org_repo_created
  ON memory_entries (org_id, repo_name, created_at DESC);