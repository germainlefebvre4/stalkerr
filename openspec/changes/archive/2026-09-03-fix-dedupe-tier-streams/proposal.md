## Why

`BuildStreams` (introduced in the radarr/sonarr unification, `internal/scheduler/build.go`) builds tier-1 ("still missing per Radarr/Sonarr") and tier-2 ("already downloaded, upgrade-eligible") movie/season streams independently, from two different data sources (the remote API's missing list vs. local `ProcessedLine` state), and never cross-checks them by `SourceKey`. When a work unit satisfies both conditions at once — e.g. Radarr still reports a movie as missing while the local DB already has a `downloaded` `ProcessedLine` for it plus another eligible candidate — two separate streams are built for the same `SourceKey` (`movie:<id>` or `series:tvdb:<id>:season:<n>`) and both become independently claimable. Since both streams resolve to the same `BaseDestPath`, and the downloader renames its temp file into that path on success, the second stream to finish silently overwrites the first's file — observed in production as a movie ("Dune - Première partie (2021)") downloaded and logged twice in the same run, with the second download clobbering the first on disk. It also wastes bandwidth/time re-downloading a multi-GB file for no benefit.

## What Changes

- `BuildStreams` SHALL ensure each work unit (identified by `SourceKey`) is represented by at most one claimable stream per run: when a movie or series-season qualifies for both tier-1 and tier-2 classification, tier-1 (still missing) SHALL take precedence and the tier-2 stream for that same `SourceKey` SHALL be dropped.
- Applies to both `buildTier2MovieStreams` and `buildTier2SeriesStreams`, checked against the tier-1 streams already built earlier in `BuildStreams` (movie streams, series-season streams, and any streams synthesized by `mergeIncompleteDownloads`).

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `media-download-scheduling`: adds a requirement that a work unit is represented by exactly one claimable stream — never simultaneously as both tier-1 and tier-2 — so the same destination path is never targeted by two independently-scheduled downloads in one run.

## Impact

- `internal/scheduler/build.go`: `BuildStreams`, `buildTier2MovieStreams`, `buildTier2SeriesStreams` (or the call site combining them) need a dedup check against tier-1 `SourceKey`s.
- No API or config surface changes. No breaking change to existing callers of `BuildStreams`/`ApplyLimit`.
