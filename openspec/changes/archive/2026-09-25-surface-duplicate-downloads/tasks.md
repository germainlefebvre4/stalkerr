# Tasks

## 1. Forced-origin marker on DownloadInfo (design D1)

- [x] 1.1 Add `IsForced bool` (`gorm:"default:false;not null"`) to `DownloadInfo` in `internal/models/download.go`. Verify `go build ./...` succeeds and a fresh `AutoMigrate` run creates the column with a `false` default on existing rows (existing `database` package tests, or a small migration test).
- [x] 1.2 In `internal/api/force_download.go`'s `persistForceDownloadPath`, set `IsForced = true` on both the create branch and the reuse-existing-record branch. Verify with a unit test: a forced download that creates a new `DownloadInfo` has `IsForced == true`; a forced download that reuses an existing `DownloadInfo` (resume-after-restart case) also ends with `IsForced == true`.
- [x] 1.3 Verify no automatic-pipeline code path (tier-1, tier-2, incomplete-download resume in `internal/scheduler/build.go`) sets `IsForced`, so it stays `false` by default for those records. Cover with a unit test asserting an automatically-created `DownloadInfo` has `IsForced == false`.

## 2. Shared duplicate-group detection query (design D2)

- [x] 2.1 Add `FindDuplicateDownloadGroups(db *gorm.DB) ([]DuplicateGroup, error)` (new file, e.g. `internal/database/duplicates.go`) implementing the Duplicate Group Definition: group `processed_lines` by `movie_id` and separately by `tv_show_id`, restricted to `state = "downloaded"` rows whose `download_info.status = "completed"`, excluding any row whose `download_info.is_forced = true`; a group qualifies only when 2+ remaining rows have distinct `download_info.download_path` values. Verify with unit tests covering the spec's four scenarios: different-path duplicate, same-path tier-2 upgrade (not a duplicate), forced download excluded, single completed occurrence (not a duplicate).
- [x] 2.2 Verify the query requires no Radarr/Sonarr call (pure local DB), and is efficient against the existing `movie_id`/`tv_show_id` indexes on `processed_lines` (`EXPLAIN`/index-usage check or a comment confirming the existing indexes cover it).

## 3. Continuous detection in the scheduled reconciliation pass (design D3, report-only)

- [x] 3.1 In `cmd/download.go`'s `reconcileDownloadPaths`, after the existing per-item path/file checks, call `FindDuplicateDownloadGroups` once and log each returned group as a distinguishable WARN (structured fields: movie/episode identity, each sibling's `download_info` id and `download_path`), mirroring the existing `file_missing` log shape. Verify with a test asserting the log output/fields for a detected duplicate group, and that no `download_info`/`processed_lines` record is modified as a result of this call.

## 4. Duplicate resolve action (design D3, mutating path)

- [x] 4.1 Add `POST /api/v1/downloads/:id/resolve-duplicate` (new file `internal/api/duplicate_download.go`, route registered in `internal/api/api.go` alongside `resync-path`/`cancel`). Given the `download_info` id, look up its duplicate group via `FindDuplicateDownloadGroups`; if the id isn't part of any group, respond without changing anything (404, distinct machine-readable code, e.g. `not_a_duplicate`). Verify with a handler test for the not-a-duplicate case.
- [x] 4.2 For each occurrence in the resolved group, `os.Stat` its `download_path`; for any occurrence whose file is missing, apply the same transaction shape `cancelDownload` (`internal/api/cancel_download.go`) uses to reach `cancelled` (`download_info.status` and every associated `processed_lines.state`). Leave occurrences whose file is present unchanged. Verify with unit tests covering: one-missing-one-present (only the missing one is cancelled), both-present (nothing changed, group still reported as duplicate), both-missing (both cancelled).
- [x] 4.3 Return the per-occurrence outcome (`cancelled` / `unchanged`) for the whole group in the response body. Verify with a handler test asserting the response shape.
- [x] 4.4 Verify the endpoint never runs automatically — confirm (by code inspection/test) that neither the scheduled reconciliation pass (Task 3) nor any other background process calls this resolve logic; it is only reachable via this HTTP endpoint.

## 5. Enrichment API integration (design D2/D4)

- [x] 5.1 In `internal/api/handlers_frontend.go`, extend `matchesSingleProblem`/`matchesProblem` to support `duplicate_download`, backed by `FindDuplicateDownloadGroups`. Verify with a handler test: `GET /api/v1/downloads?problem=duplicate_download` returns only downloads in a duplicate group.
- [x] 5.2 Add an optional `duplicate_siblings` field (`download_info_id`, `download_path`, `completed_at` per sibling) to `DownloadEnrichedResponse`, populated only for rows in a duplicate group and omitted otherwise (matching how `file_info`/`target_path`/`staging_path` are already omitted). Verify with a test asserting the field's presence/shape for a duplicate row and its absence for a non-duplicate row.
- [x] 5.3 Verify the default (no `problem` filter) and combined-`problem` requests also correctly include/exclude `duplicate_download` matches and the `duplicate_siblings` field, consistent with the other three existing problem values.

## 6. Frontend: Erreurs tab badge, filter, and diagnostic sidepanel

- [x] 6.1 Add `duplicate_download` to `VALID_REASONS` and the combined `COMBINED_PROBLEM_FILTER` in `frontend/src/hooks/useErrorsTab.ts`. Verify with `useErrorsTab.test.ts` covering selection of the new reason and the updated combined filter.
- [x] 6.2 Add the `duplicate_download` badge key to `ErrorReasonBadge`/`getErrorReasons` in `frontend/src/utils/errorReasons.ts`, sourced from the new `duplicate_siblings` field. Verify with a unit test asserting a download with `duplicate_siblings` produces the "téléchargement en double" badge.
- [x] 6.3 Add "Téléchargement en double" to the Erreurs tab's reason filter UI and render the new badge in `ErrorsTable.tsx`. Verify with `ErrorsTable.test.tsx` covering the new badge rendering and `ErrorsTab.test.tsx` covering the new filter option.
- [x] 6.4 Add the duplicate-group diagnostic section (sibling occurrences: path, completed date) to the Erreurs sidepanel, shown first alongside the other diagnostic reasons per the existing "diagnostic section renders first" requirement. Verify with a sidepanel component test.
- [x] 6.5 Add the resolve action to the sidepanel for `duplicate_download` rows only (calling the Task 4 endpoint), showing the per-occurrence outcome once the request completes, and confirm no other reason renders any action control. Verify with a component test covering: action present only for duplicate rows, outcome displayed after a successful call, action absent for the other three reasons.
- [x] 6.6 Add French/English translation keys for the new badge, filter option, diagnostic labels, and resolve action in `frontend/src/locales/{en,fr}` (errors/downloads namespace as appropriate).

## 7. Full verification

- [x] 7.1 Run `go test ./...` and confirm all existing and new backend tests pass.
- [x] 7.2 Run the frontend test suite and confirm all existing and new tests pass.
- [x] 7.3 Run `openspec validate surface-duplicate-downloads --strict` and confirm it reports the change as valid.
- [x] 7.4 Manually exercise the Erreurs tab against a dev/staging database seeded with a duplicate (two completed occurrences of the same movie at different paths): confirm the badge appears, the sidepanel shows both occurrences, deleting one file on disk and triggering the resolve action cancels only that occurrence, and the movie stops appearing in the duplicate list afterward.
