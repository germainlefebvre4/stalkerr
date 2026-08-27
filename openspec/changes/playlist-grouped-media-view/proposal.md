## Why

The Playlist page's "Items" view lists one row per M3U line, which for TV shows means one row per episode. A single show ingested overnight with 5 seasons of 20 episodes produces 100 rows, which can span several pages and bury other movies/shows fetched the same night. There is no way to see, at a glance, which movies or TV shows (as a whole, ignoring season/episode) were fetched recently without scrolling through every episode row.

## What Changes

- Add a new "Films & Séries" sub-tab next to the existing "Items" view on the Playlist page, showing one row per distinct movie or per distinct TV show (grouped by TMDB identity), instead of one row per M3U line.
- Each grouped row shows the title, year, and — for TV shows only — the season range covered by the episodes fetched so far (e.g. `S01-S05`).
- Grouped rows are paginated and sorted server-side by the group's most recent underlying item activity (`MAX(created_at)`) descending by default, so shows/movies with new content bubble to the top — matching the "what came in last night" use case. This requires a new server-side aggregation query; grouping cannot be done by aggregating an already-paginated page of items, since that would keep the same "one big show buries everything else" problem within the new view.
- The existing item-level filters (content type, pipeline state, TMDB-enrichment, group/title search) continue to apply: a group is included in the grouped view if at least one of its underlying items matches the active filters.
- Clicking a grouped row expands an inline sub-list of the underlying items belonging to that group, rendered with the same columns, badges, and per-item actions (Corriger/Associer, Reset) as today's Items table.
- Items with no TMDB match (`movie_id`/`tvshow_id` both null) are not distributed into per-title groups; they are collected into two fixed pseudo-groups, "Films non identifiés" and "Séries non identifiées", one per content type, which expand the same way to list their underlying items.
- The existing day-based date-group headers ("Aujourd'hui" / "Hier" / full date) are reused for the grouped view, keyed off each group's most recent activity date, consistent with how the Items view already groups by `created_at` day.

## Capabilities

### New Capabilities
- `playlist-grouped-media-view`: defines the grouped (movie/TV-show-level) alternative to the Playlist Items view — the sub-tab toggle, the server-side grouping/pagination/sorting contract, how existing filters interact with groups, the non-enriched pseudo-groups, and the click-to-expand item sub-list.

### Modified Capabilities
(none — this is a net-new, additive view; it does not change the behavior of the existing Items listing endpoint or table.)

## Impact

- Backend: a new `GET /api/v1/items/grouped` endpoint (`internal/api/handlers.go`, route registered in `internal/api/api.go`) running a `GROUP BY` aggregation over `processed_lines` joined to `movies`/`tvshows`, exposing per-group title, year, season range, and most-recent-activity timestamp, with the same filter query params as `/api/v1/items` and its own pagination.
- Frontend: a new sub-tab control on the Playlist page (`frontend/src/components/PlaylistTab.tsx` or a new sibling component), a new data hook analogous to `usePlaylist` for the grouped endpoint, new response types in `frontend/src/types.ts`, new translation keys in `frontend/src/locales/{en,fr}/playlist.json`, and reuse of the existing date-group utilities (`frontend/src/utils/date.ts`) and the existing items-table rendering for the expanded sub-list.
