## Context

Today `GET /api/v1/items` paginates `processed_lines` rows one-to-one with M3U lines. `Movie` rows are already unique per `(tmdb_title, tmdb_year)`, but `TVShow` rows are **not** unique per show — there is one `TVShow` row per `(tmdb_title, tmdb_year, season, episode)`, so a single show's episodes are scattered across many `processed_lines`/`tvshows` rows with no existing show-level entity to query. See `proposal.md` for why this makes the current Items view unusable for spotting "what came in last night" once a show has many episodes.

## Goals / Non-Goals

**Goals:**
- Paginate the grouped view over distinct movies/shows, computed at query time (`GROUP BY`), so pagination reflects the number of distinct titles, not the number of underlying items.
- Reuse the existing `/api/v1/items` listing (filters, table rendering, per-item actions) to power the click-to-expand sub-list, instead of introducing a second item-listing endpoint.
- Keep the new endpoint's filter query parameters identical in name/semantics to `/api/v1/items` so the two views stay in sync when a filter is active.

**Non-Goals:**
- No configurable sort for the grouped view in this change — it is always ordered by most-recent underlying-item activity, descending. (Column-based sort like the Items table can be a later addition.)
- No grouping for `channels`/`uncategorized` content types — the "Films & Séries" sub-tab only covers movies and TV shows.
- No schema/migration changes — grouping is computed at query time from existing tables.
- No poster/synopsis or per-group state breakdown in the grouped row (only title, year, season range) — matches the agreed row content.

## Decisions

### Grouping key
- Movie groups: `movies.id` (equivalently, TMDB identity — `movies` already has a unique index on `(tmdb_title, tmdb_year)`).
- TV show groups: `tvshows.tmdb_id`, **not** `tvshows.id`, since every episode has its own `TVShow` row. All `TVShow` rows sharing a `tmdb_id` belong to the same show regardless of season/episode or accidental duplicate rows for the same episode.
- Unmatched: two fixed pseudo-groups, one for `content_type = 'movies' AND movie_id IS NULL`, one for `content_type = 'tvshows' AND tv_show_id IS NULL`. A pseudo-group is only returned when it has at least one matching item.

### Query shape
A single new endpoint `GET /api/v1/items/grouped` builds the result as a `UNION ALL` of four shapes, all filtered by the same `WHERE` clause the existing `listItems` handler already builds (content_type, state, group_title, tvg_name, tmdb_enriched — extracted into a shared filter-building helper so both handlers stay in sync):
1. `movies` aggregate: `GROUP BY movies.id` → title, year, `season_start`/`season_end` = `NULL`, `latest_activity` = `MAX(processed_lines.created_at)`.
2. `tvshows` aggregate: `GROUP BY tvshows.tmdb_id` → title, year (from any row in the group — identical across the group by construction), `season_start` = `MIN(season)`, `season_end` = `MAX(season)` (ignoring `NULL` seasons), `latest_activity` = `MAX(processed_lines.created_at)`.
3. Unmatched-movies pseudo-group (present only if non-empty): fixed type, no title/year/season, `latest_activity` = `MAX(created_at)` over `movie_id IS NULL` rows.
4. Unmatched-tvshows pseudo-group (present only if non-empty): same, over `tv_show_id IS NULL` rows.

The outer query orders the union by `latest_activity DESC` and applies `LIMIT`/`OFFSET`; total count is a `COUNT(*)` over the same unioned subquery without the limit.

**Alternative considered**: aggregate client-side over an already-paginated `/api/v1/items` page. Rejected — it reproduces the exact problem this change fixes (a single show's episodes filling a page and hiding everything else), since grouping must happen before pagination, not after.

### Filter semantics carry over unchanged
The `WHERE` clause (state, tmdb_enriched, content_type, group_title, tvg_name) is applied to `processed_lines` **before** the `GROUP BY`. This means: a group appears iff at least one underlying item matches (satisfies the agreed "at least one item matches" rule for free), and a season range or activity timestamp reflects only the matching subset — e.g. filtering by `state=downloaded` on a show with 99 pending + 1 downloaded episode shows that show with a season range derived from just the downloaded episode. This is intentional and consistent with the filter applying at item level first.

### Expanding a group reuses `/api/v1/items`, extended with two optional filters
Rather than embedding items in the grouped response (which would need its own nested pagination for a 100-episode show) or adding a second items-listing endpoint, `/api/v1/items` gains two optional query params:
- `movie_id` (int): restrict to items linked to that movie.
- `tmdb_id` (int, only meaningful with `content_type=tvshows`): restrict to items whose linked `TVShow.TMDBID` matches, across all its season/episode rows.

Expanding the two unmatched pseudo-groups needs no new params — they're already expressible today as `content_type=movies&tmdb_enriched=no` / `content_type=tvshows&tmdb_enriched=no`.

The grouped response therefore only needs to carry the type + identifying key (`movie_id` for movie groups, `tmdb_id` for show groups, absent for the two pseudo-groups whose expansion filter is static) — not a full item list.

### Frontend structure
- The Playlist page gains a `view` URL parameter (`items` default, `grouped`) alongside the existing content-type buttons, read/written the same way other Playlist URL state is (`useURLState`), so the chosen view is shareable/bookmarkable and survives a refresh like the rest of the page's filters.
- A new component renders the grouped table and owns its own paginated fetch against `/api/v1/items/grouped`, reusing the existing date-group header utilities (`getDateGroupLabel`/`getDateGroupStarts`) keyed off each group's `latest_activity`.
- Clicking a group row expands an inline accordion section directly beneath it, rendering the existing items-table row markup (same columns, badges, actions) for that group's items, fetched from `/api/v1/items` with the appropriate filter and its own independent pagination (so an expanded 100-episode show doesn't dump all rows at once) — rather than a second modal/drawer, so the user keeps visual context of which group they expanded.

## Risks / Trade-offs

- **Query cost**: the union-of-aggregates plus an outer `ORDER BY`/`LIMIT` is more expensive than the current flat paginated `SELECT`. → Existing indexes on `movie_id`, `tv_show_id`, `content_type`, and `created_at` cover the join/filter columns; given this is a single-tenant self-hosted tracker (not a multi-tenant, high-QPS service), this is not expected to need further tuning for v1. Revisit if it becomes measurably slow.
- **NULL seasons**: `TVShow.Season`/`Episode` are nullable in the model. → `MIN`/`MAX` naturally ignore `NULL`s; if every row in a show's group has a `NULL` season, `season_start`/`season_end` are both `NULL` and the frontend renders no season range for that row (same visual treatment as a movie row).
- **Shared filter logic duplication**: extracting the `WHERE`-building logic out of `listItems` into a helper shared with the new handler risks subtly diverging behavior between the two endpoints over time. → Keep it as a single shared function from the start rather than copy-pasting, so any future filter change (new param, new validation rule) is made once.
