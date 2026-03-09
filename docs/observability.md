# Observability and Auditability (v1)

## Request IDs
- Every API request is assigned `X-Request-Id` in middleware.
- Incoming `X-Request-Id` is preserved when provided by clients/proxies.
- Request IDs are attached to request context for downstream service usage.
- All JSON error envelopes include:
  - `error.code`
  - `error.message`
  - `error.requestId`

## API Structured Logging
- API request middleware emits structured completion logs with:
  - `request_id`
  - `method`
  - `path`
  - `status`
  - `duration_ms`
  - `user_id` when available
  - `organization_id` when available from route params

## Worker Structured Logging
- Worker task execution logs include:
  - `task_type`
  - `job_id`
  - `repository_id`
  - `organization_id`
  - `triggered_by_user_id`
  - `attempt`
  - `attempts`
  - `status`
  - `duration_ms`
- Logs are emitted at task start and completion/failure.

## Audit Log Coverage
- `organization.created`
- `github.connection_succeeded`
- `repository.imported`
- `repository.sync_triggered`
- `repository.memory_generation_triggered`
- `repository.digest_generation_triggered`

Worker lifecycle transitions are intentionally **not** recorded to `audit_logs` to avoid noisy business history.

## Security/Privacy Notes
- OAuth access tokens are never logged.
- Audit metadata is concise and business-action oriented.
- Operational logs and audit rows are intentionally separate concerns.

## Dev-only Debug Tooling
- No dedicated debug page was added in this slice.
- Reason: request/job tracing is available through structured logs + persisted jobs/audit tables already, and we avoid exposing internal debug UI paths prematurely.
