# Contracts Package

This package owns the API contract between `apps/api` and `apps/web`.

## Source Of Truth
- OpenAPI spec: `packages/contracts/openapi.yaml`
- Generated TypeScript client: `packages/contracts/generated`

## Regenerate Client
From the repo root:

```bash
make generate-contracts
```

Or directly:

```bash
corepack pnpm --dir packages/contracts generate
```

## Update Workflow
1. Update API handlers/DTOs in `apps/api` and `packages/contracts/openapi.yaml` in the same change.
2. Regenerate the client.
3. Run API + web tests.

Do not hand-edit files in `packages/contracts/generated`.
