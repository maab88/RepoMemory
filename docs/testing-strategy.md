# Testing Strategy

## Frontend component tests
- Use Vitest + React Testing Library for rendering and interaction behavior.
- Focus on accessibility-aligned queries (`getByRole`, `getByLabelText`) and user-visible outcomes.
- Keep component tests fast and deterministic (no network).

## Frontend integration tests
- Test route-level behavior that combines components, state, and API boundaries.
- Prefer MSW/mocks for API responses in integration scope.
- Cover loading, empty, success, and error states for major flows.

## Playwright E2E strategy
- Reserve E2E for golden user paths:
  - landing/dashboard availability
  - memory list/search flow
  - digest viewing flow
- Use seeded data and stable selectors.
- Keep E2E suite small and high-signal to reduce flake and runtime.

## Go unit tests
- Primary focus: service logic, validators, mappers, and worker job behavior.
- Use table-driven tests where helpful.
- Avoid external dependencies in unit tests.

## Go handler/repository tests
- Handler tests verify transport contract (status code, payload shape, validation mapping).
- Repository tests verify SQL behavior against test DB (or limited integration fixture).
- Keep integration tests focused on high-risk query paths and regression-prone behavior.

## Coverage philosophy
- Target strong coverage on core backend logic (80%+ in critical service paths).
- Prefer meaningful assertions over line-count gaming.
- Require tests for all non-presentational new features.

## Code that does not need heavy testing
- Static presentational-only UI with no behavior.
- Thin wiring files with no logic (e.g., simple DI/bootstrap) when covered indirectly.
- Generated code that is validated by source contract and smoke tests.