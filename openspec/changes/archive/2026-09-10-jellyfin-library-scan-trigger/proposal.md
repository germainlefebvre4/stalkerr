## Why

Stalkeer already writes newly downloaded movies/episodes straight into the configured Jellyfin-shared library folders (`MoviesPath`/`TVShowsPath`), but Jellyfin only picks them up on its own periodic or manual library scan. That delay breaks the "playlist → watchable media" loop the rest of the pipeline otherwise closes automatically: a user can see an item as `downloaded` in Stalkeer and still not find it in Jellyfin for hours.

## What Changes

- Add a Jellyfin API client (`internal/external/jellyfin`) that reports specific changed library paths via Jellyfin's targeted `Library/Media/Updated` endpoint.
- Add a `jellyfin` configuration section (`url`, `api_key`, `enabled`) following the existing `RadarrConfig`/`SonarrConfig` pattern.
- `stalkeer download` and `stalkeer resume-downloads` each collect the on-disk folder path (movie folder, or season folder for TV) of every item they successfully complete during the run, de-duplicate them, and — once per run, only when Jellyfin is enabled and at least one path was collected — send a single grouped notification to Jellyfin for those paths.
- The Jellyfin notification is best-effort: a failure (timeout, unreachable, bad credentials) is logged as a warning and never changes the run's exit code or its download statistics.

## Capabilities

### New Capabilities
- `jellyfin-library-scan-notification`: configuring a Jellyfin instance, collecting changed library paths during a download run, and sending a single best-effort targeted scan notification per run so newly downloaded media becomes visible in Jellyfin without waiting for its own periodic scan.

### Modified Capabilities
(none — no existing capability's requirements change; `download` and `resume-downloads` gain an additional best-effort side effect but their documented download/scheduling behavior is unchanged)

## Impact

- **New code**: `internal/external/jellyfin` (client), a `JellyfinConfig` struct in `internal/config/config.go`, path-collection plumbing in `cmd/download.go` (`downloadStats`) and `internal/downloader/resume_helper.go` (`ResumeDownloads`).
- **Config surface**: new `jellyfin.url` / `jellyfin.api_key` / `jellyfin.enabled` keys in `config.yml` / env vars, disabled by default. `config.yml.example` and `.env.example` gain documented examples.
- **Out of scope**: Plex and Emby support; a dashboard settings UI (Radarr/Sonarr have none either); wiring the new config keys into the Helm chart's ConfigMap/Secret (can follow later, the app works standalone via `config.yml`/env vars in the meantime).
