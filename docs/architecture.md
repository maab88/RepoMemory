# Architecture

## Monorepo structure
- `apps/web`: user-facing web UI built with Next.js App Router, TypeScript, Tailwind, and TanStack-compatible client patterns.
- `apps/api`: synchronous HTTP API with chi; validates requests, applies business rules, and exposes JSON DTOs.
- `apps/worker`: asynchronous background processing using Asynq and Redis queues.
- `packages/contracts`: OpenAPI source-of-truth and generated TypeScript client artifacts.
- `infra/migrations`: immutable SQL migrations for PostgreSQL schema evolution.

## Responsibilities by package
- Web:
  - Render UI and call the API through explicit contract DTOs.
  - No direct DB or queue access.
- API:
  - Handle auth boundary (mock now, real later), validation, orchestration, persistence, and enqueue background work.
  - Own synchronous request latency and error semantics.
- Worker:
  - Execute retry-safe/idempotent background tasks.
  - Isolate long-running or high-variance workloads from request path.
- Contracts:
  - Define REST operations and schemas.
  - Keep frontend/backend type alignment explicit and reviewable.
- Infra:
  - Provide deterministic local dependencies and schema history.

## Request flow
1. Browser calls `apps/web` route or action.
2. Web sends REST request to `apps/api` using OpenAPI-backed DTO expectations.
3. API performs validation/business logic and reads/writes PostgreSQL.
4. API enqueues non-blocking work to Redis/Asynq when needed.
5. Worker consumes jobs, performs side effects/computations, and persists results back to PostgreSQL.
6. Web fetches updated state from API.

## Why REST + OpenAPI
- Stable, explicit contract between independently deployable frontend and backend.
- Easy client generation and schema review in pull requests.
- Works well for low-cost MVP hosting and tooling.

## Why PostgreSQL + Redis
- PostgreSQL:
  - Durable relational source of truth for memory entries, digests, and sync state.
  - Strong indexing/query capabilities for searchable product needs.
- Redis:
  - Fast queue broker for deferred work and retries.
  - Keeps API response times predictable.

## Why separate background jobs from API
- Protects user-facing latency from slow/variable workloads.
- Enables retries, backoff, and idempotency without duplicating HTTP behavior.
- Allows independent horizontal scaling and failure isolation for async tasks.