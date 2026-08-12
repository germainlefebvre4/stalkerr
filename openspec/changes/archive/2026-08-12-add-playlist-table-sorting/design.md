## Context

`GET /api/v1/items` (`internal/api/handlers.go:listItems`) already reads `sort`/`order` query params and applies them via `query.Order(fmt.Sprintf("%s %s", sortBy, strings.ToUpper(sortOrder)))`, but only whitelists `tvg_name`, `created_at`, `processed_at`, `group_title`, and never validates `sortOrder` itself. The frontend (`usePlaylist.ts`, `PlaylistTab.tsx`) never sends `sort`/`order` at all — headers are static text.

There is no download-completion timestamp on `ProcessedLine`. The closest existing field is `DownloadInfo.CompletedAt`, set in `downloader.go` (`Download()`) via a separate, non-transactional `UPDATE` a few lines before `ProcessedLine.State` is set to `downloaded`. That table has no index on `completed_at`, and no query today joins `ProcessedLine` to `DownloadInfo`. PostgreSQL is the only real runtime target (SQLite is test-only).

There's also no existing SQL JOIN from `processed_lines` to `movies`/`tv_shows` anywhere in the codebase — TMDB data is only ever fetched via GORM `Preload`.

## Goals / Non-Goals

**Goals:**
- Server-side sort for every playlist table column that maps to a well-defined, single value per item.
- A reliable `downloaded_at` that is set atomically with the `downloaded` state transition, not derived from a separately-written, unindexed field on another table.
- Close the `sortOrder` SQL-injection gap while extending the same code path.
- Non-enriched items always sort to the end of a TMDB-title sort, independent of direction.

**Non-Goals:**
- Sorting by fields with no single natural value (e.g. `resolution`, which is nullable and not currently a visible column) or by nested `Movie`/`TVShow` fields other than title (e.g. year) — can be added later using the same JOIN if requested.
- Backfilling `downloaded_at` for already-downloaded historical items (see Migration Plan).
- Changing `DownloadInfo.CompletedAt` or the downloader's non-transactional write pattern — out of scope; `downloaded_at` is deliberately independent of it (see Decisions).
- Multi-column ("sort by A then B") sorting — single active sort column only, matching the single-click-header UX.

## Decisions

**1. Denormalize `downloaded_at` onto `ProcessedLine` instead of joining `DownloadInfo.CompletedAt`.**
Set it in the same `updateProcessedLineState(..., StateDownloaded)` call (`downloader.go:294`) that already flips `state`, in the same UPDATE statement (`state = ?, downloaded_at = ?`). This guarantees `state=downloaded` and `downloaded_at IS NOT NULL` are always consistent — no cross-table join, no dependency on `DownloadInfo.CompletedAt`'s separate, currently non-transactional write, no need for a new index on `download_info.completed_at`. Alternative considered: join `DownloadInfo.CompletedAt` at query time — rejected because it reuses a field that (a) is already set by a different, best-effort code path that can drift from `state`, (b) is unindexed, and (c) would require preloading/joining a 1-to-many relation from `ProcessedLine`, which GORM's `Preload` cannot sort by directly (would need a raw JOIN regardless, at which point denormalizing is simpler and cheaper).

**2. Compute TMDB-title sort via `LEFT JOIN movies` + `LEFT JOIN tv_shows` with `COALESCE`.**
```sql
SELECT processed_lines.*
FROM processed_lines
LEFT JOIN movies ON movies.id = processed_lines.movie_id
LEFT JOIN tv_shows ON tv_shows.id = processed_lines.tv_show_id
ORDER BY COALESCE(movies.tmdb_title, tv_shows.tmdb_title) IS NULL,
         COALESCE(movies.tmdb_title, tv_shows.tmdb_title) <ASC|DESC>
```
The `IS NULL` boolean-as-first-sort-key trick (false=0 sorts before true=1 in Postgres) forces non-enriched rows last in both directions, per the explicit product decision that non-enriched items must never lead a TMDB sort (they aren't filterable/actionable from that ordering). This JOIN is only added to the query when `sort=tmdb_title`; other sort fields keep the existing simple `ORDER BY` with no JOIN, to avoid a needless join cost on the common case.

**3. Validate `sortOrder` against an explicit `{"asc": true, "desc": true}` whitelist (case-insensitive), same pattern as the existing `sortBy` whitelist.**
Any other value → `400 invalid_sort_order`, query never executes. This is a direct fix for the pre-existing injection gap, using the same validation style already established for `sortBy` — no new pattern introduced.

**4. Extend `validSortFields` to `{tvg_name, group_title, state, created_at, downloaded_at, tmdb_title}` and branch the query builder on `sortBy` to decide whether to add the JOIN and/or explicit NULL handling.**
Keeps the non-JOIN path (four of six fields) exactly as simple/fast as today. `downloaded_at` gets its own branch — see Decision 6 below — and does not follow the plain `ORDER BY <sort> <order>` path despite needing no JOIN.

**6. Explicit `NULLS LAST` handling for `downloaded_at`, regardless of sort direction: `ORDER BY downloaded_at IS NULL, downloaded_at <ASC|DESC>`.**
Corrects an assumption in this document's original Migration Plan (below), which claimed PostgreSQL sorts NULLs last on `DESC` by default. It does not: PostgreSQL's actual default is NULLs-first on `DESC` and NULLs-last on `ASC`. Verified against a live PostgreSQL instance: with one row holding a real `downloaded_at` and the rest `NULL` (the expected post-launch shape, per the no-backfill decision), a bare `ORDER BY downloaded_at DESC` placed every `NULL` row *before* the one real timestamp — the opposite of "most recently downloaded first," defeating the feature's primary use case. The fix mirrors the `IS NULL`-as-first-sort-key technique already used for `tmdb_title` (Decision 2), which happens to work correctly on both PostgreSQL and SQLite (the test suite's backend) since boolean-valued `IS NULL` sorts identically on both, unlike each engine's differing implicit NULL-ordering default.

**5. Frontend: add `sort`/`order` to `PLAYLIST_URL_SCHEMA` in `usePlaylist.ts`, following the exact pattern already used for `state`/`type`/`tmdb`.**
`isValid` restricts `sort` to the six accepted values and `order` to `asc`/`desc`, mirroring backend validation client-side (defense in depth / immediate UI feedback on a hand-edited URL). Header clicks call a new `setPlaylistSort(column)` that toggles direction if the column is already active, else defaults to ascending, then resets `page` to 1 (same "filter changed → reset page" behavior already applied to other filters in `usePlaylist.ts:114-131`).

## Risks / Trade-offs

- [Risk] Sorting by `tmdb_title` adds two `LEFT JOIN`s on a table (`processed_lines`) that can be large → potential query slowdown on that specific sort. → Mitigation: JOIN is added conditionally, only when `sort=tmdb_title`; `movies.id`/`tv_shows.id` are primary keys (indexed by default), so the join itself is cheap — the risk is limited to the `ORDER BY` on an unindexed `tmdb_title` column, acceptable given current data volumes (playlist sizes are in the thousands, not millions).
- [Risk] `downloaded_at` will be `null` for every item that reached `downloaded` state before this change ships (no backfill). → Mitigation: acceptable per product decision (recent downloads are the use case); documented in Migration Plan below. A future one-off backfill from `DownloadInfo.CompletedAt` remains possible if needed later, since that data isn't being removed.
- [Risk] `downloaded_at` and `DownloadInfo.CompletedAt` will now be two independent timestamps that could, in rare partial-failure cases, disagree with each other (same non-transactional-write caveat that already exists for `state`/`CompletedAt` today). → Mitigation: not a regression — this already exists between `state` and `CompletedAt`; `downloaded_at` is written in the exact same statement as `state`, so it is *more* consistent with `state` than `CompletedAt` is, not less.

## Migration Plan

1. Add nullable `downloaded_at timestamp` column to `processed_lines` via GORM `AutoMigrate` (adding the field to the `ProcessedLine` struct is sufficient given `runMigrations()`'s existing `AutoMigrate` call).
2. No backfill: existing `downloaded` items keep `downloaded_at = NULL` until they're re-processed, or forever if never touched again. Sorting by `downloaded_at` explicitly forces `NULL`s last regardless of direction (Decision 6) — required for launch precisely because the no-backfill decision guarantees `NULL`s coexist with real timestamps from day one, and PostgreSQL's actual default (`NULL`s first on `DESC`) would otherwise bury real, recent downloads underneath every never-touched legacy row.
3. Rollback: dropping the column is safe (nullable, no other code reads it) if this needs to be reverted; the sort-whitelist/JOIN/validation changes in `listItems` are equally safe to revert independently since they're additive to existing logic.
