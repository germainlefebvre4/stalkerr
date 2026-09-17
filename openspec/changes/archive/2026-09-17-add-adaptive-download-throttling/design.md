## Context

See `proposal.md` - Why. Key constraints established during exploration, grounded in the current code:

- `download` and `resume-downloads` are short-lived batch commands (typically cron-triggered), not a persistent daemon. `runDownloadWorkerPool` (`cmd/download.go`) starts `parallel` goroutines that each loop `ClaimNext()` → drain a stream → release, until the scheduler reports nothing left, then the process exits. There is no "wait for the schedule window to open" mode today, and this design intentionally does not add one (see Decisions).
- Each `Downloader.Download()` call already threads a `context.Context` through `http.NewRequestWithContext`, so cancelling that context already aborts the in-flight `io.Copy` cleanly at the transport level (`internal/downloader/downloader.go`).
- A policy-aborted transfer is **not** the same problem `cancel-download-occurrence` solves. That capability explicitly does not interrupt a live transfer and is user/API-driven; this design's abort is internal to the same OS process performing the transfer and does not touch or extend the cancel API. Both mechanisms can coexist unmodified.
- `Download()` always calls `downloadFile(ctx, ..., 0, ...)` (`downloader.go:349`) — the byte-range resume path (`downloadFileWithResume` with `startByte > 0`) exists but is never reached from the current call graph, and every call unconditionally `os.RemoveAll`s its temp directory on return (`downloader.go:180`). Per the agreed v1 scope, this design does not change that: a policy-aborted transfer restarts from 0% next time.
- `isRetryableError` (`downloader.go:97`) only treats specific `*apperrors.AppError` codes (plus `ErrFileTooSmall`) as retryable; a plain `context.Canceled`-derived error is already non-retryable today, so aborting via context cancellation cannot cause `retry.Do` to spin — but without a dedicated status, it would still fall through to the existing failure path (counted status `failed`, notified), which this design must avoid.
- Settings storage has two established patterns: a scalar KV override registry (`internal/settings/registry.go`, e.g. `radarr.url`) for simple fields, and dedicated model+table CRUD (`internal/settings/m3u_sources.go`) for structured/list data. The weekly schedule is structured (a list of windows) and follows the second pattern; the shared throttle rate and the Jellyfin-detection toggle/action/poll-interval are scalars and follow the first.
- `internal/external/jellyfin` currently only pushes (`NotifyPathsUpdated`) and reachability-checks (`SystemStatus`); it has no session/playback query today.

## Goals / Non-Goals

**Goals:**
- One shared, per-run "policy engine" that both `download` and `resume-downloads` construct, combining the weekly schedule and Jellyfin playback state into an effective policy (`none`/`throttle`/`stop`), re-evaluated live during the run.
- A single shared token-bucket rate limiter enforcing the aggregate cap across all concurrently active transfers of that run.
- A clean, non-failure abort path for a transfer stopped by policy.

**Non-Goals:**
- Byte-level resume of a policy-aborted transfer (explicit v1 scope decision — see proposal).
- A "wait for the window to reopen" run mode; a stopped run simply ends early and the next scheduled invocation picks up remaining work.
- Per-window or per-signal throttle rates (a single shared rate, per proposal).
- Debounce/hysteresis on rapid Jellyfin playback start/stop flapping.

## Decisions

**A single `PolicyEngine`, constructed once per `download`/`resume-downloads` invocation, owns both signals.**
It holds: (a) the weekly schedule, loaded once at run start (a handful of rows, cheap to hold in memory and evaluate against `time.Now()` on every check — no I/O), and (b) the last-known Jellyfin playback state, refreshed by a single background goroutine on `jellyfin.playback_poll_interval_seconds` (default 20s), not per-worker. Every worker reads the same cached Jellyfin state and computes the same schedule state, so "most restrictive wins" is evaluated consistently across all concurrent workers at any instant, and Jellyfin is never polled more than once per interval regardless of `max_parallel`.
*Alternative considered*: each worker/transfer independently polling Jellyfin. Rejected — multiplies Jellyfin API calls by `max_parallel` for no benefit, and risks workers observing inconsistent policy at the same instant.

**Rate limiting via a shared `golang.org/x/time/rate.Limiter`, wrapping each transfer's response body reader.**
The `PolicyEngine` owns one limiter instance for the whole run; `SetLimit()` is called whenever the effective policy's rate changes (unlimited under `none`, the configured shared rate under `throttle`). Each `Downloader.Download()` call wraps `resp.Body` in a policy-aware reader (alongside the existing `progressReader`) that calls `limiter.WaitN(ctx, n)` before each chunk, so the cap is aggregate across all concurrent transfers rather than per-transfer.
*Alternative considered*: a per-worker limiter sized at `sharedRate/max_parallel`. Rejected — under-utilizes the cap when fewer than `max_parallel` transfers are active, and doesn't match "single shared aggregate cap" from the proposal.

**Stop is enforced by the same policy-aware reader returning a dedicated sentinel error, not by cancelling the run's context.**
Before each chunk read, the reader checks the engine's current policy; if `stop`, it returns `ErrStoppedByPolicy` immediately instead of reading further. This bounds the abort delay to one buffer-sized read (io.Copy's default ~32KB), independent of file size, without cancelling the broader run context (which must stay alive so other, already-stopped workers can still exit their loops and the process can wind down normally).
`isRetryableError` already treats an unrecognized error as non-retryable, so `retry.Do` stops immediately; `Download()`'s post-retry handling gets one new branch: `errors.Is(err, ErrStoppedByPolicy)` sets a new status (e.g. `DownloadStatusPolicyStopped`) instead of `failed`, leaves `retry_count` untouched, and is excluded from the failure-notification path (`internal/notifier`). This status is also excluded from `cancel-download-occurrence`'s terminal states — it is a resumable, non-terminal outcome, since the same item should be attempted again once the policy clears.
*Alternative considered*: cancelling a per-item `context.Context` derived from the run's context. Rejected as equivalent in effect but more moving parts (a cancel-func registry keyed by in-flight item) for no behavioral gain, since the reader already has a natural per-chunk check point.

**The worker-claim loop gates on the same engine, mirroring "no work left".**
`runDownloadWorkerPool`'s `for { stream, ok := sched.ClaimNext(); if !ok { return } ... }` gains one condition: a worker also returns when the engine's policy is `stop`, without releasing anything back (nothing was claimed). Streams already claimed by other workers drain normally except for the abort behavior above. The run ends early; unclaimed items are simply picked up by the next scheduled invocation, since `BuildStreams` recomputes "what's missing" fresh each time. No in-process "wait" loop is introduced (see Non-Goals).

**Weekly schedule storage mirrors `m3u_sources.go`: a dedicated model/table with its own CRUD in `internal/settings`, not the scalar registry.** Each row is one window: day(s) of week, start time, end time, action. The shared throttle rate and the three Jellyfin-detection fields (`jellyfin.playback_check_enabled`, `jellyfin.playback_action`, `jellyfin.playback_poll_interval_seconds`) are added to `registry.go` as ordinary scalar fields under the existing `Jellyfin`/`Downloads` sections — no change to `app-settings`'s spec text, since it already covers those sections generically (see proposal.md - Modified Capabilities). None of the new scalar fields are added to the restart-required list: each `download`/`resume-downloads` invocation reads `settings.Effective()` fresh at process start, so there is no live long-running process for these fields to be stale in.

**Jellyfin playback detection is a new `Sessions`-polling method on the existing client, additive alongside `NotifyPathsUpdated`/`SystemStatus`.** It calls Jellyfin's `/Sessions` endpoint and treats a session as active playback when it has a `NowPlayingItem` and `PlayState.IsPaused == false`. Any error (timeout, non-2xx, auth failure) is treated the same as "no sessions" — fail-open, per the Active-Playback/Fail-Open requirements.

**New DB table registered in `internal/database/database.go`'s `AutoMigrate` call**, alongside the existing model list — an additive migration, no change to existing tables.

## Risks / Trade-offs

- **A stop that fires often on a large in-flight file can repeatedly discard its progress and never let it finish** (accepted v1 trade-off — see proposal's "Restart From Zero"). → Mitigation: none in v1; flagged as a known limitation, revisitable with real byte-resume in a follow-up change if it proves painful in practice.
- **Adding `golang.org/x/time/rate` as a new dependency.** → Low risk: it's a standard, minimal, well-maintained extended-stdlib package with no transitive dependencies of its own.
- **Misclassifying the new sentinel error** (e.g. accidentally matching `isRetryableError` and looping, or falling through to the ordinary failure path and spamming notifications) would defeat both the "not a failure" and "bounded abort" requirements. → Mitigation: a single, explicitly-checked sentinel (`errors.Is`) at both classification points; covered by unit tests on `Download()`'s error handling.
- **A worker mid-transfer still takes up to one read-buffer's worth of time to notice `stop`.** → Acceptable: bounded and independent of file size, matches the "short, bounded delay" requirement; not worth a more invasive mid-buffer cancellation mechanism.

## Migration Plan

Purely additive: a new table (empty by default → `download-bandwidth-schedule`'s "Default Schedule Is Unrestricted" requirement) and new scalar settings fields defaulting to disabled/unset. Existing installations see no behavior change until a user configures a schedule window or enables Jellyfin detection. No rollback beyond a standard prior-version redeploy is needed since nothing existing is altered.
