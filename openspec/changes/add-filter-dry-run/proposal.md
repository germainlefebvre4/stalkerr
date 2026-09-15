## Why

Filter patterns (`group_title` / `tvg_name` include/exclude regex) are currently saved blind: the user cannot see what they will actually match or exclude against real playlist content until after saving, reprocessing, and checking the results downstream. A slightly wrong regex can silently drop wanted channels or let unwanted ones through, and the only way to notice is after the fact. Users need a way to test a filter — whether still being drafted or already saved — against real, unfiltered playlist data before it takes effect.

## What Changes

- Add a new stateless, on-demand backend endpoint that tests a caller-supplied (not necessarily saved) `attribute` + include/exclude pattern combination against a chosen M3U source's most recently downloaded (unfiltered) playlist archive, without touching the database or `config.yml`.
- The endpoint returns an aggregate summary by default: total lines scanned, how many would match vs. be excluded, and the top distinct values on each side.
- The endpoint also supports an optional content search: given a substring, it returns the individual lines whose value (on the attribute being tested) contains it, each tagged as would-match or would-be-excluded, so the user can verify that a specific channel/line disappears or survives as expected.
- If the selected source has no downloaded archive yet, the endpoint reports that explicitly rather than triggering a live download.
- In the frontend "Contenu" tab: `CreateFilterDialog` gains a source picker and a "Test" action that dry-runs the include/exclude values currently typed in the form (before saving), showing the summary and search results inline.
- Each filter card in `FiltersSection` (both the origin and the active runtime override, when present) gains its own "Test" action that dry-runs the patterns already in effect for that card, read-only, without opening the create dialog.

## Capabilities

### New Capabilities
- `filter-dry-run-test`: on-demand backend check that evaluates a caller-supplied filter attribute/pattern combination against a source's latest downloaded M3U archive and reports aggregate match/exclude statistics plus line-level search, without persisting anything or depending on previously saved filters.

### Modified Capabilities
- `frontend-filters-management`: the "Filtres" section and its create dialog gain the ability to trigger a dry-run test (for in-progress and already-saved patterns) and to display its results, including a content search to verify specific lines.

## Impact

- **Backend**: new handler (e.g. `POST /api/v1/filters/dryrun`) in `internal/api`, wired in `internal/api/api.go`. Reuses `internal/filter` (pattern compilation/matching), `internal/m3udownloader` (`SourcePaths`, `ArchiveManager.GetLatestArchive`) to locate the latest archive per source, and `internal/parser` to parse it. New request/response DTOs in `internal/api/dto.go`. No database or `config.yml` changes; no writes of any kind.
- **Frontend**: `CreateFilterDialog.tsx` (source select, Test action, results panel) and `FiltersSection.tsx` (per-card Test action) in `frontend/src/components`; new `api.ts` client method; new types in `types.ts`; new `en`/`fr` locale strings for the dialogs/filters namespaces. `useM3uSources` source list is passed down to the content tab's filter components.
- **No breaking changes**: purely additive; existing filter creation, listing, and deletion behavior is unchanged.
