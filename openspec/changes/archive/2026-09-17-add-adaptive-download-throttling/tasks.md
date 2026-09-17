## 1. Weekly Schedule Storage

- [x] 1.1 Add a `BandwidthScheduleWindow` model (days of week, start time, end time, action) under `internal/models` and register it in `internal/database/database.go`'s `AutoMigrate` call; verify the migration runs cleanly against a test database.
- [x] 1.2 Add CRUD (create/update/delete/list) for schedule windows in a new `internal/settings/bandwidth_schedule.go`, mirroring the `m3u_sources.go` pattern; verify unit tests for create, update, delete, and list.
- [x] 1.3 Add a function resolving the currently active window's action against the current day/time, including overnight (end-before-start) windows and most-restrictive-wins among overlapping windows; verify unit tests covering: no window defined, a single ordinary window, an overnight window crossing midnight, and two overlapping windows with different actions.
- [x] 1.4 Add the new scalar settings fields to `internal/settings/registry.go`: `downloads.throttle_rate_kbps`, `jellyfin.playback_check_enabled`, `jellyfin.playback_action`, `jellyfin.playback_poll_interval_seconds`; verify registry tests cover get/set round-trips for each and that none are added to the restart-required list.

## 2. Jellyfin Playback Detection

- [x] 2.1 Add a `Sessions` method to `internal/external/jellyfin/jellyfin.go` calling Jellyfin's `/Sessions` endpoint; verify a unit test against a mocked server for idle, paused, and actively-playing session payloads.
- [x] 2.2 Add a helper resolving "Jellyfin currently has active playback" (at least one session with a playing, non-paused item) that treats any request error (timeout, non-2xx, auth failure) as no active playback; verify a unit test for the fail-open case.

## 3. Policy Engine

- [x] 3.1 Add `golang.org/x/time/rate` as a dependency; verify `go.mod`/`go.sum` are updated and `go build ./...` succeeds.
- [x] 3.2 Implement a `PolicyEngine` (combining the active schedule window and the cached Jellyfin state via most-restrictive-wins, exposing the current effective policy and a shared `rate.Limiter`); verify unit tests cover the full none/throttle/stop combination table for both signals, including each signal independently disabled.
- [x] 3.3 Implement the background Jellyfin poller (single goroutine, ticks on `jellyfin.playback_poll_interval_seconds`, caches the last result, applies fail-open on error) feeding the `PolicyEngine`; verify a unit test using an injectable clock/ticker.

## 4. Enforcement in the Transfer Path

- [x] 4.1 Wrap `resp.Body` in `downloader.go`'s `downloadFileWithResume` with a policy-aware reader that applies the shared rate limiter per chunk and returns a new `ErrStoppedByPolicy` sentinel when the engine's policy is `stop`; verify a unit test that a simulated mid-transfer stop aborts within one buffer-sized read.
- [x] 4.2 Handle `ErrStoppedByPolicy` in `Download()`'s post-retry error path: introduce `DownloadStatusPolicyStopped` (or equivalent), leave `retry_count` unchanged, and skip the failure-notification path; verify a unit test asserting the status, unchanged retry count, and no notification call.
- [x] 4.3 Thread the shared `PolicyEngine`/rate limiter into `Downloader`/`ParallelDownloader` construction so one instance is shared by every worker of a run; verify existing downloader tests still pass unmodified when no policy is configured (defaults to unrestricted, matching current behavior).
- [x] 4.4 Gate `runDownloadWorkerPool`'s `ClaimNext()` loop in `cmd/download.go` on the `PolicyEngine`'s stop state (a worker returns without claiming, the same as an empty queue); verify a test that no new stream is claimed while stopped and the run ends early, leaving remaining streams for the next invocation.
- [x] 4.5 Apply the same stop-gating to `cmd/resume_downloads.go`/`ResumeHelper`; verify a test asserting no new transfer starts while the policy is `stop`.

## 5. API

- [x] 5.1 Add CRUD endpoints for schedule windows; verify handler tests for create, update, delete, and list.
- [x] 5.2 Add an "effective policy" read endpoint reporting the current policy and which signal(s) are contributing; verify a handler test covering `none`, `throttle`, and `stop`, with schedule-only, Jellyfin-only, and both-contributing cases.
- [x] 5.3 Expose the new scalar settings fields (throttle rate, Jellyfin detection enabled/action/poll interval) through the existing settings-override endpoints; verify the existing app-settings API tests extend cleanly to the new keys.

## 6. Frontend

- [x] 6.1 Add a bandwidth-schedule section to the Configuration page's "Avancé" tab; verify it renders and is reachable from that tab.
- [x] 6.2 Implement the weekly schedule window editor (list, create, edit, delete, with an end time earlier than start time accepted as an overnight window); verify component tests for create, edit, delete, and the overnight-window case.
- [x] 6.3 Implement the Jellyfin throttle settings (enable toggle, throttle/stop action select) and the shared throttle-rate field; verify component tests for each control.
- [x] 6.4 Implement the effective-policy preview reading the new endpoint and refreshing without a page reload; verify a component test that it updates when the underlying policy changes.

## 7. End-to-End Verification

- [x] 7.1 Run the full backend test suite (`go test ./...`) and confirm no regressions.
- [x] 7.2 Run the frontend test suite and confirm no regressions.
- [x] 7.3 Manually verify: with a `stop` window active right now, running `download` claims no new item and aborts an in-progress one without marking it failed or consuming a retry attempt.
- [x] 7.4 Manually verify: with the weekly schedule empty and Jellyfin detection disabled, download behavior is unchanged from before this change.
