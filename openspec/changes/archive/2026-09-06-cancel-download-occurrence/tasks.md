## 1. Model changes

- [x] 1.1 Add `DownloadStatusCancelled = "cancelled"` to `internal/models/download.go` and `StateCancelled = "cancelled"` to `internal/models/processed_line.go`; update `DownloadInfo.IsEligibleForResume` doc comment to note `cancelled` is excluded (it already falls outside the listed eligible statuses, so no logic change is needed there — verify with a unit test asserting `IsEligibleForResume` returns `false` for status `cancelled`).

## 2. Uniform retry accounting

- [x] 2.1 In `internal/downloader/state_manager.go`'s `UpdateState`, remove the `retry_count + 1` increment from the `DownloadStatusRetrying` case and add it to the `DownloadStatusFailed` case instead, so a full external attempt increments the counter exactly once regardless of how many internal HTTP sub-attempts `retry.Do` performed or whether the error was retryable. Verify with a unit test: a `Download()` call that fails on the first HTTP attempt (non-retryable error, no sub-retry) still increments `retry_count` by 1.
- [x] 2.2 In the same `DownloadStatusFailed` branch, after incrementing, compare the resulting `retry_count` against `cfg.Downloads.MaxRetryAttempts` (pass it into `UpdateState` or read it via the existing config accessor) and, when reached, write `status = "cancelled"` instead of `"failed"`, and set the linked `ProcessedLine.State = "cancelled"` in the same operation. Verify with a unit test: a `DownloadInfo` whose `retry_count` reaches `max_retry_attempts` on this failure ends with `status = "cancelled"`, and one below the threshold ends with `status = "failed"`.
- [x] 2.3 Verify by unit test that `retry_count` increments identically whether `Download()` was invoked via the tier-1 direct path (`downloadItem` in `cmd/download.go`) or via the resume path (`ResumeHelper.ResumeDownloads`) — both call the same `Downloader.Download()`/`UpdateState`, so one shared test fixture per path suffices.

## 3. Candidate-selection exclusion (matcher + resume)

- [x] 3.1 Add a unit test to `internal/matcher/matcher_test.go` confirming `FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates` do NOT return a `ProcessedLine` whose `state = "cancelled"` (relying on the existing `state IN ('processed', 'failed')` allow-list already excluding it — no query change expected, this test guards the invariant).
- [x] 3.2 Add a unit test to `internal/downloader/state_manager_test.go` confirming `GetIncompleteDownloads` does NOT return a `DownloadInfo` whose `status = "cancelled"` (same allow-list-exclusion invariant on the resume path).
- [x] 3.3 Add a scheduler-level test to `internal/scheduler/build_test.go` mirroring `TestMergeIncompleteDownloads_SkipsUnmonitoredMovie`, asserting a cancelled/retry-exhausted incomplete download is not resumed even when its movie/series is still monitored.

## 4. Cancel API endpoint

- [x] 4.1 Create `internal/api/cancel_download.go` with a handler for `POST /api/v1/downloads/:id/cancel` (id = `DownloadInfo.id`) following the eligibility/response shape of `internal/api/force_download.go`: load the `DownloadInfo` and its linked `ProcessedLine`(s); `404` if not found.
- [x] 4.2 Refuse with `409 Conflict` and a clear reason when `DownloadInfo.Status` is `completed`, `downloading`, or already `cancelled`; verify with handler tests covering each refused status.
- [x] 4.3 On an eligible status (`pending`, `failed`, `retrying`), set `DownloadInfo.Status = "cancelled"` and `error_message = "Cancelled by user"`, and set every linked `ProcessedLine.State = "cancelled"`, synchronously (no goroutine); return `200 OK` with the updated status. Verify with a handler test asserting the DB rows after a successful call, and that a sibling `ProcessedLine`/`DownloadInfo` for the same movie/show is unchanged.
- [x] 4.4 Register the route in `internal/api/api.go` alongside the other `downloads` routes. Verify with a router-level test that the route resolves.

## 5. Frontend

- [x] 5.1 Add `cancelDownload(id)` to `frontend/src/services/api.ts` calling `POST /api/v1/downloads/:id/cancel`, and add the `cancelled` status to the `DownloadEnriched`/status types in `frontend/src/types.ts`.
- [x] 5.2 In `frontend/src/components/DownloadsTab.tsx`, add an "Annuler" button in the sidepanel's actions section next to Move/Rename, shown only when `selectedItem.status` is `pending`, `failed`, or `retrying`; wire it to `cancelDownload` and update only the selected item on success (matching the existing Move/Rename/Resync update pattern — no full list reload).
- [x] 5.3 Add a `cancelled` status badge/label (French: "Annulé") in the same status-badge switch used for `completed`/`downloading`/`failed`/`pending`/`retrying`, and add `cancelled` as a status filter option. Add the corresponding i18n keys to `frontend/src/locales`.
- [x] 5.4 Verify manually in a running dev instance: open the Downloads tab, cancel a `failed` item, confirm its badge updates to "Annulé" and the action disappears from its own sidepanel, while other items are unaffected.

## 6. Documentation

- [x] 6.1 Update `config.yml.example` / `charts/stalkerr/values.yaml` comments for `max_retry_attempts` to reflect the corrected semantic ("counts one increment per full external attempt, across both the matcher and resume paths") if the existing comment describes the old behavior.
