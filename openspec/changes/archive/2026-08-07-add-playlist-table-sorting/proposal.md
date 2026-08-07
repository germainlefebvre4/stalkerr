## Why

The playlist M3U table cannot be sorted on any column today: rows always come back ordered by `created_at desc` from the backend, even though the API already accepts `sort`/`order` query params that the frontend never sends. Users cannot, for instance, list the most recently downloaded items (`state=downloaded`) by download date, because no download-date field is even exposed yet.

## What Changes

- Add server-side sorting to `GET /api/v1/items`, driven by clickable column headers in the playlist table (media name, group, pipeline state, TMDB enrichment, created date, downloaded date).
- Add a new `downloaded_at` timestamp on `ProcessedLine`, set at the same time the item's `state` transitions to `downloaded`, and expose it via `ItemResponse` and a new "Téléchargé le" column.
- Extend the backend's sort whitelist to cover `group_title` (already there), `state`, `downloaded_at`, and a computed `tmdb_title` (via `LEFT JOIN` on `movies`/`tv_shows`, `COALESCE`d across both).
- For the TMDB-title sort, force non-enriched items (no linked movie/tvshow) to always sort last, regardless of ascending/descending direction.
- **BREAKING (internal only)**: fix `sortOrder` validation in `listItems` — currently interpolated into raw SQL unchecked (`ORDER BY <sort> <order>`), a latent SQL-injection vector. Only `asc`/`desc` (case-insensitive) will be accepted; anything else returns `400 invalid_sort_order`.
- Persist the active sort column/direction in the URL query string, consistent with the existing `page`/`limit`/filter persistence pattern (`usePlaylist.ts`).

## Capabilities

### New Capabilities
- `playlist-item-sorting`: server-side sort/order contract for `GET /api/v1/items` (whitelist, validation, NULL-ordering rule for the TMDB sort), and the new `downloaded_at` field.

### Modified Capabilities
- `frontend-ihm-dashboard`: the playlist table gains clickable, sortable column headers (with a visual sort-direction indicator) and a new "Téléchargé le" column; sort state is persisted in the URL like the existing filters.

## Impact

- Backend: `internal/models/processed_line.go` (new `DownloadedAt` column + migration), `internal/downloader/downloader.go` (set `DownloadedAt` alongside `StateDownloaded`), `internal/api/handlers.go` (`listItems`: expanded/validated sort whitelist, `LEFT JOIN` for TMDB sort, `sortOrder` validation fix), `internal/api/dto.go` (`ItemResponse.DownloadedAt`).
- Frontend: `frontend/src/components/PlaylistTab.tsx` (sortable `<th>` headers, sort indicators, new column), `frontend/src/hooks/usePlaylist.ts` (`sort`/`order` added to `PLAYLIST_URL_SCHEMA`), `frontend/src/services/api.ts` (`getPlaylist` passes `sort`/`order`), `frontend/src/types.ts` (`PlaylistItem.downloaded_at`), locale files (`frontend/src/locales/{en,fr}/playlist.json`, new column header).
- No breaking change for API consumers using the default sort (`created_at desc` unchanged); only malformed/previously-unvalidated `order` values now get rejected.
