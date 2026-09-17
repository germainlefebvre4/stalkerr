## Why

Stalkerr's only download-speed lever today is the static `downloads.max_parallel` worker count — there is no way to reduce or stop downloads during specific hours (e.g. work-from-home time) or while someone is actively streaming from Jellyfin, the one signal that actually indicates the connection is contended right now.

## What Changes

- Add a configurable weekly bandwidth schedule: per day, time windows each tagged `none` / `throttle` / `stop`.
- Add Jellyfin active-playback detection (polling Jellyfin's sessions) as a second, independent signal — distinct from the existing library-scan push notification. Only an actual in-progress playback counts, not merely an open session; Jellyfin being unreachable is treated as "no active playback" (fail-open).
- Combine both signals into one effective policy with a "most restrictive wins" rule: `stop` > `throttle` > `none`.
- Enforce that policy in the `download` and `resume-downloads` commands: a single shared global rate limit (one configured throttle rate, not one per window or per signal) applies across all concurrent transfers; when the effective policy is `stop`, no new item is claimed and any transfer already in progress is aborted.
- A transfer aborted by policy is a new, distinct outcome, not a failure: it does not consume the download's retry budget and does not trigger failure notifications. **v1 scope**: an aborted transfer restarts from 0% on its next attempt — byte-level resume of a policy-aborted transfer is out of scope for this change.
- New settings: a shared throttle rate and Jellyfin playback-check enable/poll-interval, added to the existing app-settings override registry. The weekly schedule itself is new structured storage (like M3U sources), not a scalar override.
- New Configuration-page UI to edit the weekly schedule and the Jellyfin-based throttle settings.

## Capabilities

### New Capabilities
- `download-bandwidth-schedule`: weekly schedule storage, CRUD, and resolving the currently-active window's action (`none` / `throttle` / `stop`).
- `adaptive-download-throttling`: combines the weekly schedule and Jellyfin active-playback state into one effective policy and enforces it (shared global rate limit, gating new stream claims, aborting in-flight transfers without treating them as failures) across the `download` and `resume-downloads` commands.
- `frontend-bandwidth-schedule-management`: Configuration-page UI for the weekly schedule editor and the Jellyfin-throttle settings.

### Modified Capabilities
- None. The new scalar settings (shared throttle rate, Jellyfin playback-check enabled, Jellyfin poll interval) are new fields within the `Downloads` and `Jellyfin` sections `app-settings` already covers generically — its existing requirements need no text change, the same way adding a field to those sections has never required one. The new fields' own keys/defaults/behavior are specified as part of `adaptive-download-throttling` below, the same way `m3u-source-overrides` specifies its own fields without modifying `app-settings`.

## Impact

- `internal/downloader`: a rate-limited/interruptible reader wraps the transfer's copy loop; a new non-failure status is introduced for a policy-aborted transfer. Existing retry/resume mechanics (`internal/downloader/resume.go`, `resume_helper.go`) are unchanged.
- `internal/external/jellyfin`: new session-polling method, additive alongside the existing `NotifyPathsUpdated`/`SystemStatus`.
- `internal/settings`: new dedicated model/table for the weekly schedule (mirrors `m3u_sources.go`'s pattern), plus new scalar registry entries.
- `internal/api`, `cmd/download.go`, `cmd/resume_downloads.go`: new endpoints for schedule CRUD and effective-policy visibility; the worker claim loop (`runDownloadWorkerPool`) is gated on the effective policy.
- `frontend/`: new Configuration-page section for the schedule editor and Jellyfin-throttle settings.
- New dependency: a token-bucket rate limiter (e.g. `golang.org/x/time/rate`), not currently in `go.mod`.
