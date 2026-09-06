## Context

`listItemGroups` (`internal/api/handlers_grouped.go`) builds each group row from a `UNION ALL` of four aggregate arms (movies, tvshows, unmatched movies, unmatched tvshows), each computing `MAX(pl.created_at) AS latest_activity` over the group's contributing `processed_lines`. The frontend expand (`PlaylistGroupedView.tsx` → `api.getGroupItems`) re-fetches from `GET /api/v1/items` filtered only by `movie_id`/`tmdb_id`/pseudo-group content type — with no run or date filter — so it returns the group's full history rather than just what earned it its current `latest_activity`.

Capability `processing-run-items` already attributes every `processed_lines` row to the `processing_logs` run that created or last updated it (`processing_log_id`), and `GET /api/v1/items` already accepts `processing_log_id` as a filter (used today by `RunItemsDialog`). See proposal.md for why this is scoped to run attribution rather than calendar-day bucketing.

## Goals / Non-Goals

**Goals:**
- Scope an expanded group's item list to the run that produced its `latest_activity`.
- Reuse the existing `processing_log_id` filter and column rather than introducing a new filtering mechanism.

**Non-Goals:**
- Changing how groups are listed, ordered, dated, or paginated.
- Backfilling `processing_log_id` for legacy rows that predate run attribution.
- Changing the date-group headers ("Aujourd'hui"/"Hier") shown above the grouped table.

## Decisions

**Compute the group's run id as `MAX(pl.processing_log_id)`, not a correlated lookup of "the run id belonging to the row with `MAX(created_at)`".**
`processing_logs.id` is an auto-incrementing primary key and every run's `processed_lines` rows are written with `created_at` timestamps at or after that run started, with runs executing strictly in id order. So within a single group's contributing rows, the row with the highest `created_at` and the row with the highest `processing_log_id` are the same row, and a plain `MAX()` aggregate — computed the same way `latest_activity` already is — is sufficient. A correlated subquery or window function would express the same guarantee more defensively, but at the cost of a materially more complex query across four UNION arms for no behavioral difference given how runs are created (`internal/processor/processor.go`: one `ProcessingLog` row created before processing starts, all lines in that run get its id, `checkDuplicate` skips writing to already-seen lines rather than touching their attribution — see capability `processing-run-items`).

**Field naming: `latest_processing_log_id` on the group response, alongside the existing `latest_activity`.**
Symmetric with the existing field and self-describing; avoids colliding with the per-item `processing_log_id` already used as an `/api/v1/items` query parameter and response field.

**Fallback when a group's most recent item predates run attribution (`latest_processing_log_id` absent):**
The frontend omits the `processing_log_id` query parameter entirely for that expand request, falling back to today's unscoped behavior (list everything) rather than sending a request that would deterministically return an empty list. This only affects data that predates the `processing-run-items` capability; new data always carries an attribution.

## Risks / Trade-offs

- [A group's `latest_processing_log_id` is null for legacy data, so its expand briefly reverts to showing full history for that group only] → Acceptable and temporary: it only affects groups whose most recent touch predates run attribution; every group touched by a run since than capability shipped is unaffected, and this matches how `processing-run-items` already documents pre-migration runs (empty-state scenario) rather than inventing new fallback semantics.
- [Adding a new aggregated column to all four UNION arms touches a query already noted as intentionally kept portable between SQLite and Postgres] → Low risk: `MAX()` over an existing indexed column is standard SQL supported identically by both backends, unlike the arms' existing constraint against parenthesized compound-select branches.
