# Deterministic Hotspots (v1)

## Scope
- Hotspots are generated from persisted local `pull_requests` and `issues` only.
- No GitHub API calls are made in hotspot recalculation.
- The worker task is `repo.recalculate_hotspots`.

## Time Window
- Fixed deterministic window: last **30 days** based on `updated_at_external`.

## Heuristic Rules
- Candidate themes are derived from explicit keyword groups:
  - `auth`, `billing`, `queue`, `retry`, `sync`, `permissions`, `migration`, `notifications`, `webhooks`
- A hotspot is emitted only when the same theme appears in **3+ distinct sources** (PRs/issues) in the window.
- Labels contribute only when they include keywords for the theme.
- Bug-oriented signals (`bug`, `fix`, `incident`, `regression`, etc.) influence risk/why-it-matters wording.

## Duplicate Strategy
- One active hotspot memory entry per `(repository, hotspot_key)` where `hotspot_key = hotspot:<theme>`.
- Reruns update the same `memory_entries` row (type = `hotspot`) instead of inserting duplicates.
- Source links are inserted with conflict-safe semantics; reruns do not duplicate `memory_entry_sources` rows.

## Source Linking
- Up to 5 sources are linked per hotspot per run.
- Source selection is deterministic: newest `updated_at_external` first, then stable tie-breakers.
- `source_type` remains explicit (`pull_request` / `issue`).

## Low-Signal Behavior
- If no theme reaches threshold, no hotspot entries are created.
- This is treated as a successful no-op, not an error.

## Known v1 Limitations
- Keyword/theme coverage is intentionally narrow and English-centric.
- There is no temporal trend charting yet (only bounded window detection).
- No cross-repository clustering yet; hotspots are repository-local.
