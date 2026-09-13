## 1. Local guard: stop re-adding already-downloaded content to tier-1

- [x] 1.1 In `buildTier1MovieStreams` (`internal/scheduler/build.go`), call `findDownloadedLine(deps.DB, "movie_id", dbMovie.ID)` right after matching `dbMovie` and before `FindMovieDownloadCandidates`; `continue` (skip the movie) when it returns a non-nil line. Verify with a new case in `internal/scheduler/build_test.go`: a movie still reported missing by Radarr but with a `downloaded` `ProcessedLine` locally yields no tier-1 stream for it.
- [x] 1.2 In `buildTier1SeriesStreams`, call `findDownloadedLine(deps.DB, "tv_show_id", dbShow.ID)` per episode, placed before the existing `FindTVShowDownloadCandidates`/empty-candidates check, and `continue` when it returns a non-nil line — without creating an empty `seasonMap` entry. Verify with a new case in `internal/scheduler/build_test.go`: a season with one already-downloaded episode and one genuinely-missing episode (per Sonarr) yields a season stream containing only the missing episode.
- [x] 1.3 Run `go test ./internal/scheduler/...` and confirm `dedupTier2MovieStreams`/`dedupTier2SeriesStreams` and the existing tier-1/tier-2 precedence tests still pass unchanged now that tier-1 no longer claims already-downloaded work units.

## 2. Carry Radarr movie ID / Sonarr series ID on Stream

- [x] 2.1 Add `RadarrMovieID int` to `Stream` in `internal/scheduler/types.go` (0 = unknown/not applicable), documented like the existing `SeriesID` field.
- [x] 2.2 Populate `RadarrMovieID` in `buildTier1MovieStreams` from the `radarr.Movie` already fetched there.
- [x] 2.3 Populate `RadarrMovieID` in `buildTier2MovieStreams` from the `radarrMovie` already fetched via `deps.Radarr.GetMovieByTMDBID`.
- [x] 2.4 Populate `Stream.SeriesID` in `buildTier2SeriesStreams` from the `series` already fetched via `deps.Sonarr.GetSeriesByTVDBID` (currently left at 0 for tier-2).
- [x] 2.5 Add/extend `internal/scheduler/build_test.go` cases verifying `RadarrMovieID`/`SeriesID` are populated for tier-1 and tier-2 streams built from a live lookup, and left at 0 for streams synthesized by `mergeIncompleteDownloads` without one.

## 3. Radarr/Sonarr rescan client methods

- [x] 3.1 Confirm the exact `/api/v3/command` request body Radarr and Sonarr expect for `RescanMovie`/`RescanSeries` against the actually-deployed versions (or their API docs) — design.md's open question — before writing 3.2/3.3.
- [x] 3.2 Add `func (c *Client) RescanMovie(ctx context.Context, movieID int) error` to `internal/external/radarr/radarr.go`, POSTing to `/api/v3/command` via the client's existing `newRequest`/retry/breaker plumbing. Verify with a new test in `internal/external/radarr/radarr_test.go` using a test HTTP server asserting method, path, and request body.
- [x] 3.3 Add `func (c *Client) RescanSeries(ctx context.Context, seriesID int) error` to `internal/external/sonarr/sonarr.go`, same pattern. Verify with a new test in `internal/external/sonarr/sonarr_test.go`.

## 4. Wire per-item, run-deduplicated rescan notification into `download`

- [x] 4.1 Add `notifiedMovieIDs`/`notifiedSeriesIDs map[int]struct{}` to `downloadStats` (`cmd/download.go`), guarded by the existing `stats.mu`, with a method (e.g. `shouldNotify(kind, id)`) that returns true and records the id only the first time it's called for that id in the run.
- [x] 4.2 In `drainStream`/`downloadItem`, on a successful item, read the owning `Stream`'s `RadarrMovieID`/`SeriesID`; when non-zero and `shouldNotify` returns true, call `radarrFullClient.RescanMovie`/`sonarrFullClient.RescanSeries` immediately (bounded by a short timeout), logging any error without touching `stats.Failed` or the run's exit status.
- [x] 4.3 Add a test (e.g. `cmd/download_notify_test.go`) verifying a rescan call failure leaves `stats.Failed`/the run's reported statistics and exit status unaffected, and is logged.
- [x] 4.4 Add a test verifying Radarr and Sonarr rescans are attempted independently: a failing/unreachable Radarr rescan does not prevent a Sonarr rescan for a completed episode in the same run, and vice versa.
- [x] 4.5 Add a test verifying multiple episodes of the same series completing in one run trigger exactly one Sonarr `RescanSeries` call for that series (per-run dedup via `notifiedSeriesIDs`).
- [x] 4.6 Add a test verifying no rescan call is attempted for the service that isn't configured for the run (e.g. Sonarr absent, only Radarr configured).

## 5. Full verification

- [x] 5.1 Run `go test ./...` and fix any regressions.
- [x] 5.2 Run `stalkeer download --dry-run --verbose` (or review `BuildStreams`' output in a test) against data shaped like the Vaiana/Pat'Patrouille incident (movie still Radarr-missing, one `downloaded` `ProcessedLine` already present) and confirm no stream is produced for it.
- [x] 5.3 Run `openspec validate fix-duplicate-content-redownload --strict` and fix any reported issues before archiving the change.
