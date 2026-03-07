# RepoMemory

RepoMemory is a SaaS MVP that turns GitHub engineering activity into searchable memory entries and weekly digests.

This repository is a monorepo with:
- `apps/web`: Next.js (App Router, TypeScript, Tailwind)
- `apps/api`: Go API (chi)
- `apps/worker`: Go worker (Asynq + Redis)
- `packages/contracts`: OpenAPI contract and generated TypeScript client placeholder
- `infra/migrations`: SQL migrations

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
2. Install frontend deps:
   - `pnpm install`
3. Start infrastructure:
   - `make up`
4. Run apps in separate terminals:
   - `make dev-api`
   - `make dev-worker`
   - `make dev-web`

## Services
- Web: http://localhost:3000
- API health: http://localhost:8080/health
- Postgres: localhost:5432
- Redis: localhost:6379

## Commands
- `make up`: start postgres + redis
- `make down`: stop containers
- `make logs`: tail container logs
- `make test`: run API and web tests
- `make lint`: run basic lint checks
- `make dev-web`: run Next.js app
- `make dev-api`: run API server
- `make dev-worker`: run worker bootstrap

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

### Worker (`apps/worker/.env`)
- `WORKER_ENV`: environment name
- `REDIS_ADDR`: Redis address for Asynq

### Web (`apps/web/.env.local`)
- `NEXT_PUBLIC_API_BASE_URL`: browser-facing API base URL

## Notes
- SQL migrations are managed via files in `infra/migrations`.
- The OpenAPI contract is in `packages/contracts/openapi/openapi.yaml`.
- `packages/contracts/generated/client.ts` is a safe placeholder until generation is added.