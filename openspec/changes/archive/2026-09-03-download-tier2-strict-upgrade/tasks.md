## 1. Persist language on ProcessedLine

- [x] 1.1 Add `Language *string` to `models.ProcessedLine` (`internal/models/processed_line.go`) and verify `go build ./...` succeeds and the app starts with `AutoMigrate` adding the `language` column without error
- [x] 1.2 In `internal/processor/processor.go`, capture the `VF`/`MULTI`/`VOSTFR` token already matched by `qualitySuffixRe` (or the equivalent detection pass) and set it on the `ProcessedLine` being built, leaving it `nil` when no token matches; verify with unit tests covering a `VF` entry, a `MULTI` entry, a `VOSTFR` entry, and an entry with no language marker (`language = NULL`)

## 2. Language-then-resolution candidate ranking

- [x] 2.1 Extend `resolutionOrderSQL` in `internal/matcher/matcher.go` into a compound `ORDER BY` (language `CASE`: `VF`=1, `MULTI`=2, `NULL`=3, `VOSTFR`=4, then the existing resolution `CASE`, then `created_at DESC`) and update `FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates` to use it
- [x] 2.2 Update/extend `internal/matcher/matcher_test.go` to verify: a `VF` candidate is returned before a `MULTI` candidate regardless of resolution; an unmarked-language candidate is returned before a `VOSTFR` candidate; resolution still breaks ties within the same language (mirrors `m3u-quality-selection` spec scenarios)

## 3. Tier-2 gating: only strictly-better candidates

- [x] 3.1 Add a small ranking helper (e.g. `candidateRank(language, resolution *string) int` or a comparable struct) in `internal/scheduler` reusing the same preference order as task 2.1, so tier-2 gating and SQL ordering agree
- [x] 3.2 Change `countDownloaded` (or add a sibling helper) in `internal/scheduler/build.go` to also return the already-downloaded `ProcessedLine`'s language/resolution (not just a count), for both `buildTier2MovieStreams` and `buildTier2SeriesStreams`
- [x] 3.3 In `buildTier2MovieStreams`/`buildTier2SeriesStreams`, after fetching the SQL-ordered candidates (task 2.1 makes `candidates[0]` the best untried one), only build the tier-2 stream when `candidateRank(candidates[0]) < candidateRank(downloadedLine)`; verify with unit tests for: worse-language untried candidate → no stream, worse-resolution-same-language untried candidate → no stream, better-language untried candidate → stream built, better-resolution-same-language untried candidate → stream built (mirrors the `media-download-scheduling` spec's new "Tier-2 stream requires a strictly better candidate" scenarios)

## 4. Tier-2 destination path sourced from Radarr/Sonarr

- [x] 4.1 Add `GetMovieByTMDBID(ctx, tmdbID int) (*radarr.Movie, error)` to the `RadarrClient` interface and `GetSeriesByTVDBID(ctx, tvdbID int) (*sonarr.Series, error)` to the `SonarrClient` interface in `internal/scheduler/build.go`; update `fakeRadarr`/`fakeSonarr` in `internal/scheduler/build_test.go` (and any other test fakes implementing these interfaces) accordingly, and verify `go build ./...` and `go vet ./...` succeed
- [x] 4.2 In `buildTier2MovieStreams`, after a movie passes the strictly-better gate (task 3.3), call `deps.Radarr.GetMovieByTMDBID` and use the returned `movie.Path` in `BuildRadarrDestPath` instead of `""`; on a lookup error, log and skip that movie for this run (do not abort `BuildStreams`); on "no match found", fall back to the existing `cfg.Downloads.MoviesPath`-based path
- [x] 4.3 Apply the same pattern to `buildTier2SeriesStreams` using `deps.Sonarr.GetSeriesByTVDBID` and `series.Path`
- [x] 4.4 Add/extend `internal/scheduler/build_test.go` cases verifying: a tier-2 movie/series stream resolves to the same `BaseDestPath` as its original tier-1 download when both use the same live `Path`; a lookup error skips that candidate without failing the whole `BuildStreams` call; a "not found" result falls back to the config-based path (mirrors the `sonarr-series-path-routing` spec's new tier-2 scenarios)

## 5. End-to-end verification

- [ ] 5.1 Run the full test suite (`go test ./...`) and confirm all scheduler, matcher, and processor tests pass
- [x] 5.2 Run `openspec validate download-tier2-strict-upgrade --strict` and confirm it reports the change as valid
- [ ] 5.3 Manually exercise `download --dry-run` (or equivalent) against a fixture/local Radarr+Sonarr setup: (a) a movie already downloaded as `VF 720p` with a leftover untried `VF 4K` candidate should plan no tier-2 stream (720p already outranks 4K); (b) a movie already downloaded as `MULTI 1080p` with a leftover untried `VF 720p` candidate should plan a tier-2 stream targeting that `VF` candidate (better language outranks resolution) — use this to sanity-check the ranking direction end-to-end, not just at the unit level
