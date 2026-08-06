## Why

Runtime filter overrides silently never take effect in the real M3U processing pipeline: `createFilter`/`updateFilter` persist `include_patterns`/`exclude_patterns` as raw comma-separated strings (the format used everywhere else — API responses, the DB column, the UI), but `filter.Manager.LoadFromDatabase()` expects a JSON-encoded array and fails to parse every single row ever written. Compounding this, `filter.Manager.LoadFromConfig()` aborts entirely the moment one attribute's pattern fails to compile (as `config.yml`'s `tvg_name.include_patterns: ["*"]` currently does), which also prevents `LoadFromDatabase()` from ever running — so one bad pattern on either attribute silently disables every filter, origin and runtime, for both attributes, while `internal/processor/processor.go` only logs a generic "continuing without filters" warning that gives no indication of scope or cause. The net effect: the runtime-override feature has never actually been enforced by the real pipeline, only by its CRUD/display layer.

## What Changes

- **BREAKING**: Align the serialization format for runtime filter patterns between the write path (`createFilter`/`updateFilter`) and the read path (`filter.Manager.LoadFromDatabase()`) so a pattern persisted through the API can actually be parsed back and applied during processing. Which side changes to match the other is a design decision (see design.md); existing `filter_configs` rows that are currently unparsable will need a stated migration/backfill approach.
- `filter.Manager.LoadFromConfig()` and `LoadFromDatabase()` isolate failures per attribute: a pattern that fails to compile for `group_title` no longer prevents `tvg_name`'s origin filter or any runtime override (on either attribute) from loading, and vice versa.
- `createFilter`/`updateFilter` validate include/exclude patterns as valid regex before persisting (via the existing but currently-unused `filter.ValidatePattern()`), returning a distinct `invalid_pattern` error instead of silently accepting a pattern that only breaks later, deep in the processing pipeline.
- Loading failures that do still occur are logged with the affected attribute and pattern instead of today's generic, scope-free "continuing without filters" warning.

## Capabilities

### New Capabilities
- `filter-loading-engine`: the contract for how filter patterns are serialized between the API/DB layer and `filter.Manager`, and how `LoadFromConfig`/`LoadFromDatabase` isolate and report per-attribute loading failures.

### Modified Capabilities
- `frontend-filters-management`: the Create Filter Configuration requirement gains pattern-validation behavior — an invalid regex is rejected with a specific error before it is persisted, instead of being silently accepted.

## Impact

- Backend: `internal/filter/filter.go` (serialization format, per-attribute isolation, logging), `internal/api/handlers.go` (createFilter, updateFilter — pattern validation), `internal/api/dto.go` (possible new error code), `internal/processor/processor.go` and `internal/dryrun/dryrun.go` (unaffected interfaces, but their observed logging/error behavior changes as a consequence of the isolation fix).
- Frontend: `frontend/src/components/CreateFilterDialog.tsx` (surface the new `invalid_pattern` error), locale files (en/fr) for the new error message.
- Data: existing `filter_configs` rows written before this fix are unparsable under the current (broken) contract; the design must state whether they are migrated, left as dead rows, or something else.
