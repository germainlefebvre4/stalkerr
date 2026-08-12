## Why

The "Filtres de Tri" page only lists runtime filters stored in the database — it never shows the file-based defaults defined in `config.yml` (`filter.group_title` / `filter.tvg_name`), even though those defaults are always loaded and can silently stop applying once a runtime filter exists for the same attribute (`internal/filter/filter.go` picks either the config-based filter set or the runtime one, never both). A user creating a small runtime filter today has no way to see that they just replaced dozens of pre-existing exclusion rules, and no easy way to start from what's currently active instead of retyping it.

## What Changes

- New read-only endpoint exposing the `config.yml` filter patterns (origin config) per attribute (`group_title`, `tvg_name`), so the frontend can display them without a database round-trip.
- **BREAKING**: `POST /api/v1/filters` (and `PATCH /api/v1/filters/:id` when it changes an attribute) now enforces at most one active runtime filter per attribute — creating/updating a filter for an attribute that already has an active runtime override automatically replaces the previous one instead of allowing them to stack. The underlying matching behavior (origin config vs. the single active runtime override, never both merged) is unchanged from today's "runtime replaces origin" semantics, but it's now unambiguous since more than one runtime filter per attribute can no longer coexist.
- `FiltersTab.tsx` groups filters by attribute (Group Title, TVG Name) instead of a flat card grid, showing the origin `config.yml` patterns (read-only, badge "🔧 Origine") alongside the active runtime override for that attribute, if any (badge "✏️ Surcharge active"), making clear which config is actually in effect.
- `CreateFilterDialog.tsx` gains a "Reprendre la config actuelle" button that loads the currently active patterns (the active override if one exists, otherwise the origin config) for the selected attribute into the Include/Exclude fields, and shows a warning when submitting will replace an existing active override.

## Capabilities

### New Capabilities
- `filter-override-policy`: backend rules for exposing origin (`config.yml`) filter config per attribute and enforcing a single active runtime override per attribute, including auto-replacement on create/update.

### Modified Capabilities
- `frontend-filters-management`: filters list view groups by attribute and displays origin vs. active-override provenance; create-filter dialog supports loading the currently active config into the form and warns before replacing an existing override.

## Impact

- Backend: `internal/filter/filter.go` (Manager), `internal/api/handlers.go` (createFilter, updateFilter), `internal/api/dto.go`, `internal/api/api.go` (new endpoint), `internal/config/config.go` (expose filter config getter if needed).
- Frontend: `frontend/src/components/FiltersTab.tsx`, `frontend/src/components/CreateFilterDialog.tsx`, `frontend/src/hooks/useFilters.ts` (or equivalent), `frontend/src/services/api.ts`, `frontend/src/types.ts`.
- Data: no schema migration required — `filter_configs` table gains a uniqueness rule enforced at the application level (one active row per `attribute`), not necessarily a DB constraint.
