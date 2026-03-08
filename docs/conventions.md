# Conventions

## Naming conventions
- Go packages/files: short, lower_snake or lower case names, nouns for domains.
- TypeScript files/components: kebab-case file names, PascalCase component names.
- Env vars: UPPER_SNAKE_CASE with app prefix where useful (`API_`, `WORKER_`, `NEXT_PUBLIC_`).
- DTO structs/types: explicit names ending with `Request`/`Response` when crossing process boundaries.

## Folder conventions
- `apps/api/internal/http/handlers`: transport layer only (parse, validate, map errors, call services).
- `apps/api/internal/.../service`: business rules/orchestration.
- `apps/api/internal/.../repository`: persistence logic.
- `apps/worker/internal/worker`: job bootstrapping, handlers, and idempotent execution units.
- `apps/web/app`: routes/layouts.
- `apps/web/components`: reusable UI components.
- `apps/web/lib`: client utilities and data access helpers.

## Service layer vs handler layer
- Use handler layer for HTTP concerns only:
  - decode/encode JSON
  - status codes
  - input validation boundaries
- Use service layer for business behavior:
  - orchestration across repositories/queues
  - invariants and policy decisions
  - retry/idempotency rules

## Test selection rules
- Unit tests:
  - pure logic in services, mappers, validators, job handlers.
- Integration tests:
  - repository/database interactions and selected API handlers with real infra or realistic test doubles.
- E2E tests:
  - critical user journeys spanning web -> api -> persistence outcomes.

## DTO expectations
- DTOs are explicit, versionable, and never implicit map passing.
- Transport DTOs must not leak DB row structs directly to clients.
- Dates/timestamps must use ISO 8601 on all API responses.

## Error response conventions
- JSON error envelope for API responses:
  - `code`: stable machine-readable error code.
  - `message`: safe human-readable message.
  - `requestId` (when available): trace correlation id.
- Avoid leaking internal stack traces or SQL driver messages.

## Environment variable conventions
- Every env var must be documented in README and app-level `.env.example`.
- Config loaders provide safe defaults for local dev where appropriate.
- Missing required secrets should fail fast at startup.

## Migration strategy
- Use forward-only SQL files under `infra/migrations`.
- Never auto-sync schema from ORM.
- One migration per logical schema change, reviewed in PR.

## No hidden coupling rule
- Frontend must only depend on API contract, never API internals.
- API must not rely on frontend assumptions outside contract.
- Worker must communicate through queues/DB, not direct imports from web.
- Shared behavior must be explicit (OpenAPI/contracts, documented schemas, or versioned packages).