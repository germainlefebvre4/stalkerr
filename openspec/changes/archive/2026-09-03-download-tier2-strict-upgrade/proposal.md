## Why

The unified `download` command keeps re-downloading movies and episodes the user already has. Tier-2 ("already-downloaded, upgrade-eligible") streams are rebuilt every run for any leftover untried `ProcessedLine`, not just when a genuinely better one exists, so every remaining quality variant eventually gets pulled down over successive 2-hour cycles. Worse, Tier-2's destination path is recomputed locally from TMDB metadata instead of being sourced from Radarr/Sonarr like Tier-1 is, so an "upgrade" download lands in a different folder than the original instead of replacing it — turning a redundant download into a genuine duplicate file. Candidate ranking also only considers resolution, ignoring language (VF/MULTI/VOSTFR), which matters more for this French IPTV catalog.

## What Changes

- `ProcessedLine` gains a persisted `language` field, detected from the M3U entry alongside the existing resolution detection.
- Candidate ordering (`m3u-quality-selection`) ranks primarily by language preference (`VF` > `MULTI` > unspecified > `VOSTFR`), then by the existing resolution preference within the same language tier.
- Tier-2 stream construction (`media-download-scheduling`) only builds a stream for a movie/series-season when its best untried candidate is strictly better — by language, then resolution — than the candidate already downloaded for it. A leftover candidate that is equal or worse no longer creates a Tier-2 stream.
- Tier-2 destination paths (`sonarr-series-path-routing`) are sourced from a live Radarr/Sonarr lookup (the same `GetMovieByTMDBID` / `GetSeriesByTVDBID` calls already used by Tier-1 and by the force-download path), instead of being recomputed from `cfg.Downloads.MoviesPath`/`TVShowsPath` + TMDB title/year. This makes an accepted upgrade land at the same destination path as the original download, so it replaces the file instead of creating a duplicate.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `m3u-quality-selection`: adds language detection/persistence on `ProcessedLine` and folds language preference into the candidate ordering used for download selection.
- `media-download-scheduling`: tier-2 stream construction requires the best untried candidate to be strictly better (language, then resolution) than the already-downloaded candidate, instead of merely existing.
- `sonarr-series-path-routing`: destination path resolution via a live Radarr/Sonarr lookup also applies to tier-2 upgrade downloads, not just tier-1 missing-content downloads.

## Impact

- **Backend**:
  - `internal/models/processed_line.go` (new `Language` field, migrated via `AutoMigrate`)
  - `internal/processor/processor.go` (language detection alongside existing resolution detection)
  - `internal/matcher/matcher.go` (`FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates` ordering: language then resolution)
  - `internal/scheduler/build.go` (`buildTier2MovieStreams`/`buildTier2SeriesStreams`: strictly-better gating against the already-downloaded candidate; live Radarr/Sonarr path lookup instead of the local fallback)
  - `internal/scheduler/build.go` (`BuildDeps`/`RadarrClient`/`SonarrClient` interfaces: expose `GetMovieByTMDBID` / `GetSeriesByTVDBID`, already implemented in `internal/external/radarr` and `internal/external/sonarr`)
- **Database**: migration adding the `language` column to `processed_lines`.
- **No frontend changes** required for this fix; the automatic download pipeline's behavior changes only.
