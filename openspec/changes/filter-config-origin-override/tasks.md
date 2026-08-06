## 1. Backend - Origin Config Exposure

- [ ] 1.1 Add a `GetFilterConfig()` (or equivalent) accessor on `internal/config` returning the `FilterConfig` (`group_title`, `tvg_name`) currently loaded from `config.yml`, if not already directly accessible.
- [ ] 1.2 Add `SystemFilterResponse` / `AttributeFilterResponse` DTOs in `internal/api/dto.go` shaping `{"group_title": {"include_patterns": [...], "exclude_patterns": [...]}, "tvg_name": {...}}`.
- [ ] 1.3 Add `listSystemFilters` handler in `internal/api/handlers.go` reading from `config.Get().Filter` and returning the DTO from 1.2.
- [ ] 1.4 Register `GET /api/v1/filters/system` in `internal/api/api.go`.

## 2. Backend - Single Active Override Per Attribute

- [ ] 2.1 In `createFilter` (`internal/api/handlers.go`), wrap the insert in a transaction that first deletes any existing `filter_configs` row with `is_runtime = true` and the same `attribute` as the request, then inserts the new row.
- [ ] 2.2 In `updateFilter`, when the request changes `attribute` to a value different from the filter's current one, apply the same transactional delete-then-write against the target attribute before committing the update.
- [ ] 2.3 Add/extend backend tests in `internal/api/handlers_frontend_test.go` (or a new `handlers_test.go` case) covering: create on attribute with no override, create on attribute with an existing override (old row gone, new row active), update changing attribute onto an attribute with an existing override.
- [ ] 2.4 Add/extend `internal/filter/filter_test.go` if needed to confirm `Matches()`/`LoadFromDatabase()` behavior is unaffected now that at most one runtime row per attribute can exist.

## 3. Frontend - API & Types

- [ ] 3.1 Add `SystemFilterConfig`/`AttributeFilterPatterns` types to `frontend/src/types.ts` mirroring the DTO from 1.2.
- [ ] 3.2 Add `api.getSystemFilters()` in `frontend/src/services/api.ts` calling `GET /api/v1/filters/system`.
- [ ] 3.3 Extend `frontend/src/hooks/useFilters.ts` to also fetch and expose the system/origin config (e.g. `systemFilters`, `systemFiltersLoading`) alongside the existing runtime `filters` state.

## 4. Frontend - Filters List View

- [ ] 4.1 Rework `FiltersTab.tsx` to group by attribute (`group_title`, `tvg_name`) instead of a flat card grid: one section per attribute.
- [ ] 4.2 In each attribute section, render the origin `config.yml` patterns (read-only, "🔧 Origine" badge) sourced from `systemFilters`.
- [ ] 4.3 In each attribute section, render the active runtime override (if `filters` contains a row for that attribute), with an "✏️ Surcharge active" badge, its name, and its own include/exclude patterns.
- [ ] 4.4 When an attribute has no active runtime override, omit the override section entirely (per spec scenario).

## 5. Frontend - Create Filter Dialog

- [ ] 5.1 Pass `systemFilters` and the current active `filters` (or the derived active-override-per-attribute map) into `CreateFilterDialog.tsx`.
- [ ] 5.2 Add a "Reprendre la config actuelle" button next to the attribute selector; on click, populate the Include/Exclude fields with the active override's patterns for the selected attribute if one exists, otherwise the origin patterns.
- [ ] 5.3 Show a warning banner in the dialog when the selected attribute already has an active runtime override, before the form is submitted.
- [ ] 5.4 Verify the dialog re-evaluates the warning/prefill source when the user changes the selected attribute.

## 6. Verification

- [ ] 6.1 Run backend tests (`go test ./...`) covering the new handler and filter-manager behavior.
- [ ] 6.2 Manually exercise the "Filtres de Tri" page: no override, one override, replace-via-create, and the "Reprendre la config actuelle" flow for both attributes.
- [ ] 6.3 Update `openspec/specs/frontend-filters-management/spec.md` expectations are reflected in the running UI (grouped view, badges, warning).
