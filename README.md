# RepoMemory

RepoMemory is a SaaS MVP that turns GitHub engineering activity into searchable memory entries and weekly digests.

## Monorepo Structure
- `apps/web`: Next.js App Router frontend
- `apps/api`: Go REST API (chi) + sqlc query layer
- `apps/worker`: Go background worker (Asynq)
- `packages/contracts`: OpenAPI source + generated TS client
- `infra/migrations`: SQL migrations
- `docs`: architecture, conventions, testing strategy, and ERD

## Prerequisites
- Node.js 20+
- pnpm 9+
- Go 1.23+
- Docker + Docker Compose

## Quick Start
1. Copy env files:
   - `cp .env.example .env`
   - `cp apps/api/.env.example apps/api/.env`
   - `cp apps/worker/.env.example apps/worker/.env`
   - `cp apps/web/.env.example apps/web/.env.local`
2. Install frontend deps: `corepack pnpm install`
3. Start infrastructure: `make up`
4. Run apps in separate terminals:
   - `make dev-api`
   - `make dev-worker`
   - `make dev-web`

## Services
- Web: http://localhost:3000
- API health: http://localhost:8080/health
- Postgres: localhost:55432
- Redis: localhost:6379

## Quality Commands
- `make test`: run all tests
- `make test-web`: run web tests
- `make test-api`: run API tests
- `make test-worker`: run worker tests
- `make lint`: run Go + web lint
- `make format`: run Go + web formatters
- `make ci`: run lint + test
- `make generate-contracts`: regenerate TypeScript client from OpenAPI contract

## Data Layer Commands
- Apply schema migration locally:
  - `Get-Content infra/migrations/0001_v1_schema.up.sql | docker compose --env-file .env exec -T postgres psql -U postgres -d repomemory`
- Generate sqlc code:
  - `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.28.0 generate -f apps/api/sqlc.yaml`
- Generate OpenAPI TypeScript client:
  - `make generate-contracts`
- Run DB integration tests:
  - `cd apps/api && go test ./...`
  - Optional override: set `TEST_DATABASE_URL`

## GitHub OAuth (v1)
RepoMemory v1 uses GitHub OAuth App flow (not GitHub App installation flow).

1. Create a GitHub OAuth App in your GitHub account settings.
2. Set callback URL to `http://localhost:3000/integrations/github/callback`.
3. Set API env vars in `apps/api/.env`:
   - `GITHUB_CLIENT_ID`
   - `GITHUB_CLIENT_SECRET`
   - `GITHUB_STATE_SECRET` (random secret for signed OAuth state)
   - Optional overrides:
     - `GITHUB_REDIRECT_URL` (default `http://localhost:3000/integrations/github/callback`)
     - `GITHUB_OAUTH_SCOPE` (default `repo read:user user:email`)
4. Restart API after env updates.

Connect flow routes:
- `POST /v1/github/connect/start`
- `GET /v1/github/callback` (JSON callback completion endpoint, consumed by web callback page)

Repository flow routes:
- `GET /v1/github/repositories`
- `POST /v1/github/repositories/import`

Import behavior:
- Import upserts on `(organization_id, github_repo_id)` and returns imported rows.
- Re-import updates metadata safely; it does not create duplicate repository rows.
- Initial repository sync state is upserted and an audit log entry is created per imported repository.
- GitHub repository listing uses v1 single-page fetch (`per_page=100`), no API pagination response yet.

## API Contract Workflow
- Source of truth: `packages/contracts/openapi.yaml`
- Generated client output: `packages/contracts/generated`
- Keep spec and implementation aligned in the same change:
  - update API handlers/DTOs and OpenAPI spec together
  - regenerate client
  - run API + web tests before merge

## Docs
- `docs/architecture.md`
- `docs/conventions.md`
- `docs/testing-strategy.md`
- `docs/erd.md`

## Environment Variables
### Root (`.env`)
- `POSTGRES_USER`: postgres username
- `POSTGRES_PASSWORD`: postgres password
- `POSTGRES_DB`: default database name
- `POSTGRES_PORT`: local postgres port
- `REDIS_PORT`: local redis port

### API (`apps/api/.env`)
- `API_PORT`: API HTTP port
- `API_ENV`: environment name
- `DATABASE_URL`: Postgres connection string
- `GITHUB_CLIENT_ID`: GitHub OAuth app client ID
- `GITHUB_CLIENT_SECRET`: GitHub OAuth app client secret
- `GITHUB_STATE_SECRET`: secret used to sign/validate OAuth state
- `GITHUB_REDIRECT_URL` (optional): callback URL used in OAuth exchange
- `GITHUB_OAUTH_SCOPE` (optional): OAuth scopes requested during connect

### Worker (`apps/worker/.env`)
- `WORKER_ENV`: environment name
- `REDIS_ADDR`: Redis address for Asynq

### Web (`apps/web/.env.local`)
- `NEXT_PUBLIC_API_BASE_URL`: browser-facing API base URL

### API tests
- `TEST_DATABASE_URL` (optional): Postgres connection string for integration tests.
  - Default: `postgres://postgres:postgres@127.0.0.1:55432/repomemory?sslmode=disable`
