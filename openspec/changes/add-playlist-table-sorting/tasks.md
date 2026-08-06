## 1. Backend data model

- [ ] 1.1 Add `DownloadedAt *time.Time` to `ProcessedLine` (`internal/models/processed_line.go`), tagged for `AutoMigrate` (nullable column, e.g. `json:"downloaded_at,omitempty"`).
- [ ] 1.2 In `internal/downloader/downloader.go`, extend the `updateProcessedLineState(..., StateDownloaded)` call path (around line 294) to also set `downloaded_at` to the current time in the same `Updates(...)` call that sets `state`.
- [ ] 1.3 Add `DownloadedAt *string` to `ItemResponse` (`internal/api/dto.go`) and populate it in `toItemResponse` (`internal/api/handlers.go`).

## 2. Backend sort/order handling in `listItems`

- [ ] 2.1 Extend `validSortFields` in `internal/api/handlers.go` to `{tvg_name, group_title, state, created_at, downloaded_at, tmdb_title}`.
- [ ] 2.2 Add a `validSortOrders := {"asc": true, "desc": true}` whitelist (case-insensitive) and return `400 invalid_sort_order` when `sortOrder` (lower-cased) isn't in it, before building the query.
- [ ] 2.3 When `sortBy == "tmdb_title"`, add `LEFT JOIN movies ON movies.id = processed_lines.movie_id` and `LEFT JOIN tv_shows ON tv_shows.id = processed_lines.tv_show_id` to the query, and order by `COALESCE(movies.tmdb_title, tv_shows.tmdb_title) IS NULL, COALESCE(movies.tmdb_title, tv_shows.tmdb_title) <ASC|DESC>` so non-enriched items always sort last.
- [ ] 2.4 For all other accepted `sortBy` values, keep the existing simple `ORDER BY <sort> <order>` path unchanged (no JOIN).

## 3. Backend tests

- [ ] 3.1 Add/extend tests in `internal/api/handlers_frontend_test.go` (or the relevant `listItems` test file) covering: sort by each whitelisted field, default sort when `sort`/`order` omitted, `400 invalid_sort_field` for an unknown field, `400 invalid_sort_order` for a malformed `order` value (including an injection-style payload), and non-enriched items sorting last for both `order=asc` and `order=desc` on `sort=tmdb_title`.
- [ ] 3.2 Add/extend a downloader test verifying `downloaded_at` is set when `state` transitions to `downloaded`, and remains `null` for items in other states.

## 4. Frontend data layer

- [ ] 4.1 Add `sort`/`order` fields to `PLAYLIST_URL_SCHEMA` in `frontend/src/hooks/usePlaylist.ts`, with `isValid` restricted to the six accepted sort values and `asc`/`desc`, following the existing `state`/`type`/`tmdb` pattern.
- [ ] 4.2 Add a `setPlaylistSort(column)` callback in `usePlaylist.ts`: if `column` is already the active sort, toggle `order`; otherwise set `sort=column`, `order='asc'`; in both cases reset `page` to 1 (same behavior as the existing filter-change effect at lines 114-131).
- [ ] 4.3 Update `api.getPlaylist` (`frontend/src/services/api.ts`) to accept and forward `sort`/`order` query params.
- [ ] 4.4 Add `downloaded_at: string | null` to the `PlaylistItem` type (`frontend/src/types.ts`).
- [ ] 4.5 Thread `urlState.sort`/`urlState.order` through `fetchPlaylist` in `usePlaylist.ts` and include them in the hook's returned API (`playlistSort`, `playlistOrder`, `setPlaylistSort`).

## 5. Frontend table UI

- [ ] 5.1 In `PlaylistTab.tsx`, make each sortable `<th>` (media name, group/category, TMDB enrichment, pipeline state, created date) clickable, calling `setPlaylistSort` with the corresponding backend field name.
- [ ] 5.2 Render a sort-direction indicator (e.g. ▲/▼) on the currently active sort header.
- [ ] 5.3 Add a new "Téléchargé le" column (header + cell), sorted via `downloaded_at`, rendered with the same `formatDate` used for "Créé le", showing "—" when `downloaded_at` is `null`.
- [ ] 5.4 Update `colSpan` values on the loading/empty-state rows to account for the new column.

## 6. i18n

- [ ] 6.1 Add the new "Téléchargé le" header key to `frontend/src/locales/{en,fr}/playlist.json` (e.g. `table.headers.downloadedAt`), alongside English equivalent ("Downloaded on").

## 7. Manual verification

- [ ] 7.1 Run the app, sort the playlist table by each column ascending and descending, and confirm the URL query string updates and survives a page refresh.
- [ ] 7.2 Filter by `state=downloaded`, sort by "Téléchargé le" descending, and confirm the most recently downloaded items appear first.
- [ ] 7.3 Sort by TMDB enrichment and confirm non-enriched items appear last in both directions.
