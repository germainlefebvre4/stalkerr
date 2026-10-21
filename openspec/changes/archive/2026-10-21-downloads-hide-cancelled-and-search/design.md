# Design

## Context

`listDownloadsEnriched` (`internal/api/handlers_frontend.go`) builds a GORM query on `download_info`, optionally `JOIN processed_lines ... .Distinct()` for the `type` filter, orders by `download_info.updated_at desc`, then either paginates in SQL or (when `problem` is set) loads everything, filters in Go and slices. The database is PostgreSQL in production and in-memory SQLite in tests; `getLikeOp(db)` already picks `ILIKE` / `LIKE` accordingly. The frontend keeps Downloads filters in URL query params via `useURLState` (`useDownloads.ts`, schema `dlStatus`, `dlType`, `dlProblem`, `dlPage`, `dlLimit`); a default of `''` serializes as an absent param. See proposal.md - Why for motivation.

## Goals / Non-Goals

**Goals:**
- Real downloads are visible first by default; cancelled rows stay reachable and counted.
- One deterministic ordering, identical on every code path (SQL pagination and the `problem` Go path).
- Title search that works from the Downloads tab, combinable with every other filter.

**Non-Goals:**
- Grouping occurrences by media (a possible later change).
- Changing how or when downloads become `cancelled`.
- A server-side persisted "default filter" preference.

## Decisions

**1. `active` is an API pseudo-status, not a frontend-only expansion.**
`status=active` maps to `status <> 'cancelled'` in the backend. Alternative: frontend sends a list of statuses. Rejected: it would need a multi-value `status` parameter, larger URLs, and the `cancelled_hidden` count would still need a backend rule for "what was excluded". Omitting `status` keeps the old all-rows behavior, so API clients are unaffected.

**2. Default and "Tous" are distinguished by a URL sentinel.**
`useURLState` treats the schema default as "absent". The `dlStatus` default becomes `'active'`; "Tous" serializes to `dlStatus=all`; the hook translates `all` to "no `status` param" when calling `api.getDownloads`. Alternative: keep `''` as the default and have the backend interpret empty as `active`. Rejected: it would silently change the meaning of omitting `status` for every API client and break `status`-less callers.

**3. Status-rank ordering with a `CASE` expression, always applied.**
`ORDER BY CASE download_info.status WHEN 'downloading' THEN 0 WHEN 'retrying' THEN 0 WHEN 'pending' THEN 0 WHEN 'completed' THEN 1 WHEN 'failed' THEN 2 WHEN 'cancelled' THEN 4 ELSE 3 END, download_info.updated_at DESC`. Applying it unconditionally keeps the single-status case trivially correct and removes a code branch. The Go-side `problem` path already preserves the SQL order, so it needs no change. Alternative: sort in Go. Rejected: breaks SQL pagination.

**4. Replace `JOIN ... Distinct()` by `EXISTS` subqueries for `type` and `search`.**
On PostgreSQL, `SELECT DISTINCT` requires every `ORDER BY` expression to appear in the select list, so the `CASE` rank would fail with the current `Joins(...).Distinct()` type filter. `EXISTS (SELECT 1 FROM processed_lines pl WHERE pl.download_info_id = download_info.id AND ...)` avoids `DISTINCT`, avoids duplicate rows when several lines match, and keeps `COUNT` correct. The `type` predicate moves to the same mechanism for consistency. This is a refactor of existing behavior covered by existing tests.

**5. Search predicate.**
`search` is matched with `getLikeOp(db)` against `movies.tmdb_title`, `tv_shows.tmdb_title`, `processed_lines.tvg_name` (all via the same `EXISTS` on `processed_lines` left-joined to `movies`/`tv_shows`) and `download_info.download_path` (file name part is covered by matching the full path; a path substring match is accepted). Wildcard characters (`%`, `_`) in user input are escaped. Alternative: full-text search. Rejected: overkill for a single-user tool and not portable to SQLite tests.

**6. `cancelled_hidden` computed with a second `COUNT`.**
Only when `status=active`: count rows matching the same `type`/`search`/`problem`-independent SQL predicates plus `status = 'cancelled'`. When `problem` is set, the Go filter path counts matching cancelled rows after `matchesProblem` so the number stays consistent with what "Annulés" would show. One extra indexed count per request; acceptable under the existing response-time budget.

**7. `searchName` alias is read once on load.**
The frontend reads `searchName` only as a fallback when `dlSearch` is absent, then writes `dlSearch` and does not touch the Playlist's own `searchName`. Search input is debounced (300 ms) and resets `dlPage` to 1, mirroring the existing filter-change effect in `useDownloads`.

## Risks / Trade-offs

- [Existing bookmarks and the previous "Tous" default change meaning] → The indicator "N annulés masqués" and the "Tous" option keep cancelled rows one click away; documented as **BREAKING (behavioral)** in the proposal and CHANGELOG.
- [`EXISTS` + `LIKE '%x%'` on `download_path` cannot use an index] → The `download_info` table is small (hundreds to low thousands of rows) and the existing `problem` path already scans everything; acceptable.
- [`problem` path loads all rows to compute `cancelled_hidden`] → Only when both `problem` and `status=active` are set, a rare combination already paying a full scan.
- [Status rank hard-codes status names] → Unknown statuses fall into a middle rank rather than erroring; the rank list lives next to the `DownloadStatus` constants with a test that every constant is ranked.

## Migration Plan

No schema change, no data migration. Deploy backend and frontend together (the frontend sends `status=active`, which an old backend would treat as an unknown status and return an empty list). Rollback is reverting the release; no state is written.
