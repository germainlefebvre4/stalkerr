## Why

Investigation of the production database (Vaiana (2026), La Pat' Patrouille : Le film mission Dino (2026)) showed the same movie downloaded twice, in two different resolutions, roughly two `download` cron ticks (2h) apart. Root cause: `buildTier1MovieStreams`/`buildTier1SeriesStreams` decide whether a movie/episode still needs downloading *solely* from Radarr's/Sonarr's "missing" status, and stalkeer never tells Radarr/Sonarr a file just arrived. Radarr/Sonarr only notice via their own periodic library scan, which can lag well behind stalkeer's own download cadence. Any `download` run that falls in that lag window re-adds the movie/episode as tier-1 and downloads the next best untried candidate (different resolution/language), wasting bandwidth/storage and producing duplicate files in the same folder even though the user only expects one download per request. This affects both the Radarr (movies) and Sonarr (series) paths identically.

## What Changes

- Add a local, Radarr/Sonarr-independent guard to tier-1 stream building: a movie, or a specific episode within a series-season stream, that already has a `downloaded` `ProcessedLine` locally SHALL NOT be re-added as tier-1 content, even while Radarr/Sonarr still report it missing. Only tier-2 (strictly-better-candidate upgrade, already gated separately) may touch it from then on.
- After a movie or episode's download completes successfully within the `download` command, notify Radarr/Sonarr to rescan/refresh that specific movie/series so its missing/has-file status catches up promptly instead of waiting for the next periodic library scan — closing the lag window at the source, in addition to the local guard above. Applies to both Radarr and Sonarr.
- A rescan/refresh notification failure (unreachable service, timeout, error response) SHALL be logged and SHALL NOT fail the run or change its reported statistics, mirroring the existing Jellyfin notification's fault tolerance.

## Capabilities

### New Capabilities
- `radarr-sonarr-post-download-rescan`: after a movie/episode download completes in the `download` command, request a targeted Radarr/Sonarr rescan for that specific movie/series so its "missing" status updates promptly instead of relying solely on Radarr's/Sonarr's own periodic scan.

### Modified Capabilities
- `media-download-scheduling`: tier-1 stream building for a movie or series-episode SHALL also check the local database and skip a work unit that already has a `downloaded` `ProcessedLine`, instead of relying solely on Radarr's/Sonarr's reported missing status.

## Impact

- `internal/scheduler/build.go`: `buildTier1MovieStreams` (movie-level skip) and `buildTier1SeriesStreams` (per-episode skip within a season stream), reusing the existing `findDownloadedLine` helper already used by the tier-2 builders.
- `cmd/download.go`: collect per-run successfully-completed movies/episodes (mirroring `changedPaths` tracking for Jellyfin) and issue the Radarr/Sonarr rescan notifications after the run, or per item — to be settled in design.md.
- `internal/external/radarr/radarr.go`, `internal/external/sonarr/sonarr.go`: new client method(s) to trigger a targeted rescan/refresh command.
- No database schema changes; no changes to `internal/downloader` transfer mechanics.
