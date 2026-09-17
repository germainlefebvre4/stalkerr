## 1. Backend query change

- [x] 1.1 In `internal/api/handlers_grouped.go`, rewrite the `tvshowsArm` query per design.md's window-function decision: compute `group_latest_log_id` via `MAX(pl.processing_log_id) OVER (PARTITION BY t.tmdb_id)` in an inner subquery, then aggregate `season_start`/`season_end` in the outer `GROUP BY t.tmdb_id` using a `CASE` that includes a row only when `group_latest_log_id IS NULL OR pl.processing_log_id = group_latest_log_id`. Verify with `go build ./...`.
- [x] 1.2 Verify the movies, unmatched-movies, and unmatched-tvshows arms are unchanged, and that the four arms still combine as bare `SELECT`s under `UNION ALL` (no added parens) so SQLite's compound-select grammar still accepts the query.

## 2. Backend tests

- [x] 2.1 Add a test (e.g. `TestListItemGroups_SeasonRangeScopedToLatestRun` in `internal/api/handlers_grouped_test.go`) reproducing the bug: seed a show with episodes across seasons 1-12 under an older `processing_log_id`, then episodes for seasons 13-14 under a newer `processing_log_id`, and assert the group's `season_start`/`season_end` are `13`/`14`, not `1`/`14`.
- [x] 2.2 Add a test asserting the no-run-attributed fallback: seed a show whose episodes all have a nil `processing_log_id` (or reuse `TestListItemGroups_NullProcessingLogIDReportsNoValue`'s pattern extended with multiple seasons), and assert the season range still spans all of that show's episodes.
- [x] 2.3 Run `go test ./internal/api/...` and confirm `TestListItemGroups_TVShowAggregation`, `TestListItemGroups_EndToEndFiltersAndExclusions`, and `TestListItemGroups_ReportsLatestProcessingLogID` still pass unmodified (their fixtures don't mix multiple `processing_log_id` values per show, so they should be unaffected), updating any fixture/expectation that turns out to rely on the old full-history aggregation.

## 3. Verification

- [x] 3.1 Run `go test ./...` for the backend and confirm all tests pass.
- [x] 3.2 Manually verify against a dev database (or the reproduction from the bug report) that a show whose latest run only touched a subset of its historical seasons now shows that subset's range in the grouped view, matching what expanding the group reveals.
