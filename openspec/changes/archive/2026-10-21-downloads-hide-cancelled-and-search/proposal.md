# Proposal

## Why

The Downloads tab lists every `download_info` row ordered by `updated_at DESC`. Cancelled occurrences (manual cancel, retry-exhausted, sibling occurrences cancelled by the duplicate cleanup) are all touched recently, so they flood the top of the list and push the real downloads (`completed`, `downloading`, ...) to later pages. For a title with several occurrences (e.g. "Peacemaker") the useful rows are buried under cancelled duplicates. The tab also has no name search: the `searchName` URL parameter belongs to the Playlist tab and does nothing on Downloads, so there is no way to narrow the list to one title.

## What Changes

- Status filter gains an **"Actifs"** option (`active` = every status except `cancelled`) which becomes the **default** when no status is chosen. "Tous" remains available (explicit opt-in to see cancelled rows) and `pending` / `retrying` are added as filterable statuses.
- `GET /api/v1/downloads` accepts `status=active` and returns a `cancelled_hidden` count (cancelled downloads matching the other active filters but excluded by `status=active`). The Downloads tab shows a "N annulés masqués" link under the list that switches the filter to "Annulés".
- `GET /api/v1/downloads` always orders by status rank (`downloading`/`retrying`/`pending`, then `completed`, then `failed`, then `cancelled`) and then `updated_at DESC`, so cancelled rows are last whenever they are shown.
- New **title search** on the Downloads tab (`dlSearch` in the URL, `search` on the API) matching the movie/TV show title, the M3U `tvg_name`, or the download file name, case-insensitively. `?tab=downloads&searchName=…` links are read as an alias of `dlSearch` so existing links keep working.
- **BREAKING (behavioral)**: a Downloads URL without `dlStatus` now shows "Actifs" instead of "Tous"; cancelled rows are only visible after choosing "Tous" or "Annulés".

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `downloads-enrichment-api`: `status` query accepts `active`; new `search` query parameter; deterministic status-rank ordering; `cancelled_hidden` field in the paginated response.
- `downloads-display-ui`: new status filter option set and "Actifs" default, title search input persisted as `dlSearch`, "N annulés masqués" indicator.

## Impact

- Backend: `internal/api/handlers_frontend.go` (`listDownloadsEnriched`: status/search predicates, ordering, `cancelled_hidden`), `internal/api/dto.go` (response shape), tests in `internal/api/`.
- Frontend: `frontend/src/hooks/useDownloads.ts` (URL schema, default status, search), `frontend/src/services/api.ts` (`getDownloads`), `frontend/src/components/DownloadsTab.tsx` (filter options, search input, hidden-cancelled link), `frontend/src/locales/{fr,en}/downloads.json`, `frontend/src/types.ts`, and their tests.
- No schema migration, no new dependency. Existing API clients that omit `status` keep the previous "all rows" behavior; only the sort order changes.
