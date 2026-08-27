## 1. Backend: shared filter helper

- [x] 1.1 In `internal/api/handlers.go`, extract the `content_type`/`state`/`group_title`/`tvg_name`/`tmdb_enriched` `WHERE`-clause building out of `listItems` into a shared helper function that both `listItems` and the new grouped handler can call. Verify by running the existing item-listing tests and confirming `listItems` behavior is unchanged.
- [x] 1.2 Add `movie_id` (int) and `tmdb_id` (int) optional query parameters to `listItems`, filtering `processed_lines` by `movie_id = ?` and, for `tmdb_id`, by `tv_show_id IN (SELECT id FROM tvshows WHERE tmdb_id = ?)`. Verify with a test that `GET /api/v1/items?movie_id=<id>` returns only items linked to that movie, and `GET /api/v1/items?content_type=tvshows&tmdb_id=<id>` returns every item linked to any `TVShow` row sharing that `tmdb_id`, across seasons/episodes.

## 2. Backend: grouped listing endpoint

- [x] 2.1 Add `ItemGroupResponse` DTO (fields: group type — `movie`/`tvshow`/`unmatched_movies`/`unmatched_tvshows`, `movie_id` or `tmdb_id` key where applicable, `title`, `year`, `season_start`, `season_end`, `latest_activity`) alongside the existing response types in `internal/api`.
- [x] 2.2 Implement the `movies` aggregate query (`GROUP BY movies.id`, filtered by the shared helper from 1.1, `latest_activity = MAX(processed_lines.created_at)`), verified by a test seeding one movie with 3 processed lines and asserting a single group row with the max `created_at`.
- [x] 2.3 Implement the `tvshows` aggregate query (`GROUP BY tvshows.tmdb_id`, `season_start = MIN(season)`, `season_end = MAX(season)` ignoring `NULL`, `latest_activity = MAX(processed_lines.created_at)`), verified by a test seeding one show's episodes across 3 different `TVShow` rows (same `tmdb_id`, different season/episode) and asserting exactly one group row with the correct season range.
- [x] 2.4 Implement the two unmatched pseudo-group aggregates (`content_type = 'movies' AND movie_id IS NULL`, `content_type = 'tvshows' AND tv_show_id IS NULL`), each emitted only when its `COUNT(*) > 0`. Verify with a test asserting the pseudo-group is absent when there are zero unmatched items of that type, and present with `latest_activity` set correctly when there are some.
- [x] 2.5 Union the four aggregates from 2.2-2.4, order the union by `latest_activity DESC`, and apply `LIMIT`/`OFFSET`; compute the pagination total as a `COUNT(*)` over the same unioned subquery without the limit. Verify with a test that a show with 100 episodes counts as exactly 1 unit toward `limit`/`offset` and toward the total.
- [x] 2.6 Register `GET /api/v1/items/grouped` in `internal/api/api.go` routing to the new handler, accepting the same `content_type`/`state`/`group_title`/`tvg_name`/`tmdb_enriched` params as `/api/v1/items` (via the shared helper) plus `limit`/`offset`, and rejecting/ignoring any `sort` parameter (per design, ordering is fixed). Verify by an end-to-end handler test covering: mixed movies+shows+unmatched items, a `state` filter that empties out one show's episodes down to a season subset (asserting the returned season range reflects only the matching episodes), and a `content_type=movies` filter that excludes any `tvshows`/unmatched-tvshows entries.
- [x] 2.7 Exclude `channels`/`uncategorized` content types from all aggregates regardless of filters. Verify with a test that seeds a channel item and confirms it never appears in `GET /api/v1/items/grouped` output.

## 3. Frontend: grouped view data layer

- [x] 3.1 Add `MediaGroupItem` (or equivalent) type to `frontend/src/types.ts` matching the `ItemGroupResponse` shape from 2.1.
- [x] 3.2 Add `api.getGroupedPlaylist(...)` to `frontend/src/services/api.ts` calling `GET /api/v1/items/grouped` with the same filter params as `api.getPlaylist`, plus pagination. Verify by a manual `curl`/browser-network check that the request URL and params match an equivalent `/api/v1/items` call for the same filters.
- [x] 3.3 Add a `view` field (`'items' | 'grouped'`, default `'items'`) to the Playlist page's URL-synced state (alongside the existing `usePlaylist` URL schema or a new sibling hook), so the active sub-tab is reflected in the URL and survives a refresh, consistent with the rest of the Playlist filters.
- [x] 3.4 Add a data hook (e.g. `usePlaylistGroups`) that fetches from `api.getGroupedPlaylist` using the shared content-type/state/tmdb/search filters and the `view`-scoped pagination/limit state, mirroring `usePlaylist`'s fetch-on-dependency-change pattern.

## 4. Frontend: grouped view UI

- [x] 4.1 Add the "Items" / "Films & Séries" sub-tab control to `PlaylistTab.tsx` (or a new sibling component it renders), next to the existing All/Movies/TVShows content-type buttons, wired to the `view` state from 3.3.
- [x] 4.2 Render the grouped table (title, year, and season range for TV-show rows only — `S0X` for a single season, `S0X-S0Y` for a range, formatted client-side from `season_start`/`season_end`) with i18n labels for the two unmatched pseudo-groups ("Films non identifiés"/"Séries non identifiées"). Add corresponding keys to `frontend/src/locales/en/playlist.json` and `frontend/src/locales/fr/playlist.json`.
- [x] 4.3 Apply the existing date-group header logic (`getDateGroupLabel`/`getDateGroupStarts`) to the grouped rows, keyed off each row's `latest_activity` instead of `created_at`. Verify by reusing/extending `frontend/src/utils/date.test.ts` with a case for grouped rows.
- [x] 4.4 Implement click-to-expand: clicking a group row toggles an inline section beneath it rendering the existing items-table row markup (columns, state badges, Corriger/Associer and Reset actions), fetching from `/api/v1/items` filtered by `movie_id` for movie groups, `content_type=tvshows&tmdb_id=...` for show groups, and `content_type=movies|tvshows&tmdb_enriched=no` for the two pseudo-groups, with its own independent pagination. Verify manually: expanding a show with 100 episodes shows a first page with pagination controls, not all 100 rows; expanding a movie group shows its item(s); expanding a pseudo-group shows unmatched items of that type.
- [x] 4.5 Ensure the grouped table and its mobile card-list equivalent render without horizontal overflow below the mobile breakpoint, consistent with the existing Items view's mobile treatment (`useIsMobile()`).

## 5. Verification

- [x] 5.1 Run backend tests (`go test ./...`) and confirm they pass.
- [x] 5.2 Run `npm run build` and `npm run lint` in `frontend/` and confirm both succeed with no errors.
- [x] 5.3 Start the app, open the Playlist page, switch to "Films & Séries", and manually verify: a show with many episodes appears as one row; the row shows title/year/season range only; switching content-type and state filters changes which groups appear; expanding a group lists its items with working actions; the two unmatched pseudo-groups appear only when applicable; switching back to "Items" restores the original table.
