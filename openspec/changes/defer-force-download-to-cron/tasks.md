## 1. Backend: defer the transfer

- [ ] 1.1 In `internal/api/force_download.go`, delete the goroutine (currently lines ~129-142) that calls `s.downloader.Download(...)` right after `persistForceDownloadPath` succeeds, leaving the `DownloadInfo` row it already created in status `pending`. Verify `go build ./...` succeeds and no goroutine/`s.downloader.Download` call remains in `forceDownloadItem`.
- [ ] 1.2 Confirm `forceDownloadItem` still responds `202 {status: "queued", processed_line_id}` immediately once eligibility, existence, and monitored checks pass — verified by the rewritten tests in section 3.

## 2. Backend: monitored gate

- [ ] 2.1 Extend `resolveForceDownloadMoviePath` to check the `Monitored` field already present on the `radarr.Movie` returned by `GetMovieByTMDBID`; when `false`, return a `422` (consistent with the existing static-ineligibility refusals `not_matched`/`no_stream_url`, as distinct from the live-call-dependent `404 media_not_found`/`502 existence_check_failed`) with error code `not_monitored`, before any `DownloadInfo` is persisted.
- [ ] 2.2 Extend `resolveForceDownloadEpisodePath` the same way, checking the series-level `Monitored` field on the `sonarr.Series` returned by `FindEpisodeByTVDBID` (not the episode-level field), for consistency with `media-download-scheduling`'s existing series-level monitored gate.
- [ ] 2.3 Verify both refusals happen before `persistForceDownloadPath` is called, so no `DownloadInfo` row is created or reused on a `not_monitored` refusal.

## 3. Tests: update existing fixtures and expectations

- [ ] 3.1 Add `"monitored": true` to the radarr movie fixtures in `TestForceDownloadItem_MovieSuccess_AcceptedBeforeTransferCompletes`, `TestForceDownloadItem_MovieSuccess_LanguageAndVariantTagged`, `TestForceDownloadItem_ConcurrentRequest_SecondRejectedWithoutDuplicateTransfer`, and `TestForceDownloadItem_SiblingOccurrenceUnaffected` in `internal/api/force_download_test.go`, so they keep passing once the monitored gate lands (their current fixtures omit the field, which unmarshals to `false`).
- [ ] 3.2 Rewrite `TestForceDownloadItem_MovieSuccess_AcceptedBeforeTransferCompletes`: assert the request returns `202` with a `pending` `DownloadInfo` carrying the resolution-suffixed `DownloadPath`, and that the transfer is never attempted — the blocking download source receives no hit within a short timeout, and `ProcessedLine.State` is unchanged. Drop the now-unreachable "wait for `DownloadStatusCompleted`" assertion.
- [ ] 3.3 Apply the same rewrite shape to `TestForceDownloadItem_MovieSuccess_LanguageAndVariantTagged`: assert the tagged `DownloadPath` right after the `202` response, without waiting for a transfer.
- [ ] 3.4 Rewrite `TestForceDownloadItem_ConcurrentRequest_SecondRejectedWithoutDuplicateTransfer` into an idempotent-reuse test: submit two force-download requests back-to-back against the same still-`processed` occurrence, both returning `202`, and assert exactly one `DownloadInfo` row exists afterward (the second request reuses/updates the first's row via `persistForceDownloadPath`'s existing `item.DownloadInfoID != nil` branch) instead of expecting a `409` mid-transfer.
- [ ] 3.5 Add a new test confirming the existing "already downloading" `409` refusal (spec's "Forced download is refused for ineligible occurrences" requirement, unchanged by this proposal) still works: set a `ProcessedLine.State` directly to `StateDownloading` via fixture (no in-flight transfer needed) and confirm a force-download request against it is refused with `409`.
- [ ] 3.6 Rewrite `TestForceDownloadItem_SiblingOccurrenceUnaffected` to assert immediately after the `202` response — target has a `pending` `DownloadInfo` distinct from the sibling's, sibling's `DownloadInfo`/`ProcessedLine` untouched — instead of waiting for the target to reach `StateDownloaded`.
- [ ] 3.7 After 3.2-3.6, check whether `newBlockingDownloadSource`, `waitForDownloadStatus`, `waitForProcessedLineState`, and the `hitCount`/`release` plumbing in `internal/api/force_download_test.go` still have callers; remove any that are now unused. Verify with `go build ./... ` / `go vet ./...`.

## 4. Tests: new monitored-gate coverage

- [ ] 4.1 Add `TestForceDownloadItem_MovieNotMonitored`: radarr fixture with `"monitored": false`, expect the `422 not_monitored` refusal and zero `DownloadInfo` rows created.
- [ ] 4.2 Add `TestForceDownloadItem_EpisodeSeriesNotMonitored`: sonarr fixture with the series' `"monitored": false`, expect the same refusal shape and zero `DownloadInfo` rows created.

## 5. Verification

- [ ] 5.1 Run `go test ./internal/api/... ./internal/scheduler/...` and confirm all tests pass.
- [ ] 5.2 Using the dev server, force-download a monitored occurrence and confirm: `202` response, "queued" badge shown immediately, and no file transfer starts until a manual `download` run is triggered; then force-download an unmonitored occurrence and confirm the drawer surfaces the `not_monitored` error.
