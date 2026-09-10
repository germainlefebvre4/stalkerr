## Why

Diagnosing whether a problem comes from the app itself, Radarr, Sonarr, TMDB, or disk space currently requires reading logs. There is no single place to get a quick, on-demand read of these dependencies' health.

## What Changes

- New backend endpoint aggregating, on a single request: database connectivity (reusing the existing `database.HealthCheck()`), Radarr/Sonarr/TMDB reachability, and disk usage for every configured storage path.
- Each of Radarr, Sonarr, and TMDB reports one of three states: **OK**, **KO with a short human-readable reason** (unreachable, invalid credentials, timeout, etc.), or **not configured** (when disabled in `config.yml`).
- Disk usage is reported per distinct mounted volume backing `downloads.movies_path`, `downloads.tvshows_path`, `downloads.temp_dir` (when set), and `m3u.download.archive_dir` — paths sharing the same mounted volume are deduplicated into a single entry.
- New header entry point: an icon button next to the language selector, present identically on desktop and mobile, with no health-reflecting badge — it is a static navigation affordance, not an alert.
- Clicking the icon opens a `SystemStatusDialog` (following the existing `*Dialog.tsx` / Radix Dialog pattern already used elsewhere in the app) that fetches the aggregated status on open and offers a manual "Refresh" action inside the dialog.
- All checks are on-demand only: nothing is fetched before the dialog is opened, and there is no background polling or caching between opens.

## Capabilities

### New Capabilities
- `system-status-api`: backend endpoint that aggregates DB/Radarr/Sonarr/TMDB connectivity and deduplicated disk usage into a single on-demand response.
- `system-status-view`: header icon entry point and `SystemStatusDialog` that display that aggregated status to the user, on demand.

### Modified Capabilities
(none — no existing capability's requirements change)

## Impact

- `internal/api`: new route and handler for the aggregation endpoint.
- `internal/external/radarr`, `internal/external/sonarr`, `internal/external/tmdb`: new lightweight reachability check per client (none currently expose one).
- `internal/downloader/diskspace.go` (or a new shared package): disk usage lookup extended to support deduplication by mounted volume.
- `internal/config`: read to determine each integration's configured/enabled state and resolve the storage paths to check.
- `frontend/src/components/FloatingHeader.tsx`: add the header icon trigger.
- `frontend/src/components`: new `SystemStatusDialog.tsx`.
- `frontend/src/hooks`: new hook to fetch the status on demand.
- `frontend/src/locales/{en,fr}`: new translation strings for the dialog.
