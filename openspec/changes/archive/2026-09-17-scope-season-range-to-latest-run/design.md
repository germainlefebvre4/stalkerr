## Context

`listItemGroups` (`internal/api/handlers_grouped.go`) builds the TV-show arm of the grouped listing as a single `GROUP BY t.tmdb_id` query that computes `season_start`/`season_end` as `MIN(t.season)`/`MAX(t.season)` over every matching `processed_lines` row, and separately `MAX(pl.processing_log_id) AS latest_processing_log_id` from the same rows. The frontend later uses `latest_processing_log_id` to scope the expanded item list to one run, but the season range stays unscoped. See proposal.md - Why.

The four arms (`movies`, `tvshows`, `unmatched_movies`, `unmatched_tvshows`) are combined with `UNION ALL`, and each arm must remain a bare `SELECT` (no wrapping parens) for SQLite's compound-select grammar - this constrains how the tvshows arm can be restructured, but not what it can contain internally.

## Goals / Non-Goals

**Goals:**
- Make `season_start`/`season_end` reflect only the episodes from the group's latest processing run, matching what expanding the group reveals.
- Preserve the existing fallback: when a group has no attributable run (`latest_processing_log_id` is null), keep reporting the season range over the group's full history, unchanged from today.
- Keep the query portable across SQLite and Postgres (both currently supported via GORM).

**Non-Goals:**
- No change to the movies arm, the unmatched pseudo-group arms, pagination, ordering, or the existing `content_type`/`state`/`group_title`/`tvg_name`/`tmdb_enriched` filtering.
- No change to the frontend: it already renders whatever `season_start`/`season_end` the API returns.

## Decisions

**Compute the per-group latest `processing_log_id` with a window function, then conditionally aggregate seasons against it, in a single query.**

The tvshows arm becomes a query over a subquery: the inner `SELECT` joins `processed_lines`/`tvshows` and adds `MAX(pl.processing_log_id) OVER (PARTITION BY t.tmdb_id)` as `group_latest_log_id`; the outer `SELECT` does the existing `GROUP BY t.tmdb_id` but computes:

```sql
MIN(CASE WHEN group_latest_log_id IS NULL OR pl.processing_log_id = group_latest_log_id THEN t.season END) AS season_start,
MAX(CASE WHEN group_latest_log_id IS NULL OR pl.processing_log_id = group_latest_log_id THEN t.season END) AS season_end
```

The `group_latest_log_id IS NULL` branch preserves the no-run-attributed fallback (every row in that group has a null `processing_log_id`, so `group_latest_log_id` is null and the `CASE` falls through to including every row - identical to today's unconditional `MIN`/`MAX`).

Alternatives considered:
- **Self-join subquery** (join the tvshows arm to a second aggregated subquery that computes `MAX(processing_log_id) GROUP BY tmdb_id` under the same filters): works, but duplicates the filter clause and its bound args a second time for this arm, adding parameter-binding complexity to `unionArgs`/`dataArgs`. The window-function form needs the filter clause only once (applied on the inner subquery), and there is no second call to `buildItemFilterConditions`.
- **Two round trips** (fetch groups first, then a second query per group or a batched `IN (...)` query to compute latest-run season ranges): avoided - it turns one query into N+1 or requires an extra batched query and response-side merging, adding complexity for no behavioral benefit over doing it in SQL.

Window functions are supported by both target databases (SQLite's `mattn/go-sqlite3 v1.14.22` bundles SQLite well past the 3.25 minimum; Postgres has always supported them), so portability is not a concern here, unlike the UNION ALL arm-shape constraint noted above (which this decision does not touch - the window function lives inside the tvshows arm's own subquery, not at the union level).

## Risks / Trade-offs

- [Slightly more complex SQL in the tvshows arm] -> Mitigated by keeping the change scoped to that one arm and covering it with backend tests asserting the narrowed range, the single-season case, and the no-run-attributed fallback.
- [Existing tests asserting full-history season ranges from multi-run fixtures will fail after this change] -> Expected; update their fixtures/expectations as part of implementation (see proposal.md - Impact).
