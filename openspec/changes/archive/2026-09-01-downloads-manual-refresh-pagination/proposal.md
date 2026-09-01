## Why

The Downloads tab currently re-fetches the full list every 5 seconds via a fixed `setInterval`, regardless of whether the user is actively looking at it or has any download in progress. This makes the UI visibly "flicker" (list re-renders, scroll position and open filters feel unstable) and does unnecessary work. At the same time, the list is capped at a fixed 20 items with no way to see more — even though the backend endpoint already returns pagination metadata (`total`, `total_pages`) that the frontend never uses. Users have no way to browse past the first 20 downloads or control how many they see at once.

## What Changes

- Remove the 5-second auto-refresh polling from the Downloads tab. The list SHALL only refresh on initial tab activation and when the user clicks the existing manual "refresh" button — no background timer.
- Add pagination to the Downloads list: page navigation controls plus a configurable items-per-page selector, reusing the existing `<Pagination />` component (already used on the Playlist tab and in `RunItemsDialog`) and the `offset`/`limit`/`total` contract the backend already exposes.
- Fix `GET /api/v1/downloads`'s pagination metadata (`total`/`total_pages`) to be correct when the `problem` filter is active: the `problem` filter is evaluated on parsed file metadata (not a DB column), so it currently runs in Go *after* `LIMIT`/`OFFSET`/`COUNT`, making the reported total reflect the pre-filter count instead of the actual filtered result set. The fix keeps filtering in application code (no schema change) but filters the full status/type-matched set before computing `total` and slicing the requested page in memory.
- **BREAKING** (spec-level only, not user-facing): the `downloads-enrichment-api` requirement that `problem`-filtered responses report "total reflecting pre-filter count" is replaced by a requirement that `total`/`total_pages` reflect the actual filtered result count.

## Capabilities

### Modified Capabilities
- `downloads-display-ui`: replaces the "auto-refresh every 5 seconds" requirement with manual-refresh-only behavior, and adds pagination controls (page navigation + items-per-page selector) to the Download List Summary Row requirement.
- `downloads-enrichment-api`: changes the `problem` query parameter's pagination behavior so `total`/`total_pages` reflect the post-filter count instead of the pre-filter count.

## Impact

Frontend:
- `frontend/src/hooks/useDownloads.ts` — remove the `setInterval` polling effect; add page/limit/total state; pass `offset` to `api.getDownloads` and read `total`/`total_pages` from the response; reset to page 1 when filters change.
- `frontend/src/services/api.ts` — `getDownloads` gains an `offset` parameter, mirroring `getRunItems`/`getGroupedPlaylist`.
- `frontend/src/components/DownloadsTab.tsx` — render the existing `<Pagination />` component below the downloads list, wired to the new page/limit/total state.

Backend:
- `internal/api/handlers_frontend.go` (`listDownloadsEnriched`) — when `problem` is set, fetch the full status/type-matched set, apply the problem filter in Go, compute `total` from the filtered set, and slice the requested page before building the enriched response (instead of applying `LIMIT`/`OFFSET`/`COUNT` at the DB level and filtering afterward).

Specs:
- `openspec/specs/downloads-display-ui/spec.md` — update the auto-refresh requirement and its scenario/manual-test references; add pagination requirement/scenarios.
- `openspec/specs/downloads-enrichment-api/spec.md` — update the `problem` filter requirement's pagination behavior.

No database migration, no new dependencies.
