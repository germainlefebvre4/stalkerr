## 1. Thread monitored-status data into the scheduler

- [x] 1.1 Add nil-safe `MonitoredMovieTMDBIDs map[int]bool` and `MonitoredSeriesTVDBIDs map[int]bool` fields to `scheduler.BuildDeps` (`internal/scheduler/build.go`), documented as optional/nil-safe. Verify existing `internal/scheduler` tests still compile and pass unchanged (no updates required, since the new fields default to nil).
- [x] 1.2 In `mergeIncompleteDownloads`, before folding an incomplete movie download into this run's streams (whether reusing an existing `movieByID` entry or synthesizing a new one), skip it when `deps.MonitoredMovieTMDBIDs[line.Movie.TMDBID]` is present and `false`. Verify with unit tests: a confirmed-unmonitored movie's incomplete download is excluded from the run's stream set; a movie absent from the map (including a nil map) is still resumed exactly as before.
- [x] 1.3 Apply the equivalent check for TV episodes using `deps.MonitoredSeriesTVDBIDs[*show.TVDBID]`, guarding both the "new season stream" and "existing season stream, new episode" branches (the check is per-series, so all of a confirmed-unmonitored series' incomplete episodes are excluded together). Verify with unit tests: a confirmed-unmonitored series' incomplete episodes are excluded; a series absent from the map is still resumed as before.

## 2. Wire the existing full-library fetch through

- [x] 2.1 In `cmd/download.go`'s `reconcileDownloadPaths`, alongside the existing `movieTMDBPaths`/`seriesTVDBPaths` maps built from the already-fetched `[]radarr.Movie`/`[]sonarr.Series`, also build `movieMonitored map[int]bool`/`seriesMonitored map[int]bool` (`true` when `Monitored`, entry present only for a fetched item) from the same slices, and return them alongside the function's existing behavior. No new Radarr/Sonarr API calls. Verify with a unit test asserting the returned maps reflect a fake library's `monitored` fields. (Required adding `sonarr.GetAllSeries`, an unfiltered mirror of `GetAllMovies`, since `GetAllMonitoredSeries` pre-filters unmonitored series away — see design.md's correction.)
- [x] 2.2 Update `downloadCmd.Run`'s call site to capture the two new maps and pass them into `scheduler.BuildDeps{MonitoredMovieTMDBIDs: ..., MonitoredSeriesTVDBIDs: ...}` before calling `scheduler.BuildStreams`. Verify with a `cmd`-package test: a fake Radarr library marking a movie unmonitored results in its incomplete download not appearing in the built stream set, while a monitored (or Radarr-fetch-failed) case still resumes it.

## 3. End-to-end verification

- [x] 3.1 Run the full backend test suite (`go test ./...`), confirming no regressions in `internal/scheduler`, `internal/api`, or `cmd` tests.
- [x] 3.2 Manually reproduce the original scenario: with an incomplete/stuck download whose movie is currently unmonitored in Radarr, run `download` and confirm it is no longer attempted; confirm a different incomplete download whose movie is still monitored resumes normally in the same run.
