# Tasks

## 1. Thread live movie/series path + title into BuildDeps (design D2)

- [x] 1.1 In `cmd/download.go`'s `reconcileDownloadPaths`, capture each fetched movie's/series' `Title` alongside the `Path` it already captures (from the same single `GetAllMovies`/`GetAllSeries` call), keyed the same way as the existing `movieMonitored`/`seriesMonitored` maps (`Movie.TMDBID` / `TVShow.TVDBID`). Verify with a unit test asserting the returned lookup contains both path and title for a fetched movie/series.
- [x] 1.2 Add the new combined lookup (path + title, by TMDBID/TVDBID) to `scheduler.BuildDeps` alongside the existing monitored maps, and wire it through in `cmd/download.go`. Verify `go build ./...` succeeds and existing `BuildDeps` call sites compile unchanged (nil-safe when omitted, matching the existing monitored-maps convention).

## 2. Fix `mergeIncompleteDownloads` to respect the already-downloaded rule (design D1)

- [x] 2.1 Before resuming or synthesizing a stream for an incomplete `DownloadInfo` (movie branch), call the existing `findDownloadedLine(deps.DB, "movie_id", line.Movie.ID)` helper; when it returns a downloaded line, skip this `DownloadInfo` for this run instead of resuming/synthesizing. Verify with a scheduler unit test: a movie with one `downloaded` `ProcessedLine` and a separate incomplete `DownloadInfo` for a different `ProcessedLine` of the same movie produces no stream/candidate for that incomplete download.
- [x] 2.2 Apply the same check for the TV branch, keyed by the specific episode's `tv_show_id` (`findDownloadedLine(deps.DB, "tv_show_id", show.ID)`). Verify with an equivalent series/episode unit test.
- [x] 2.3 When a `DownloadInfo` is skipped by 2.1/2.2, transition it to `DownloadStatusCancelled` (reusing the existing cancel semantics) instead of leaving it in its current incomplete status. Verify with a unit test that a skipped `DownloadInfo`'s status is updated and that it no longer appears in a subsequent `GetIncompleteDownloads` call.
- [x] 2.4 When synthesizing a new stream (no existing tier-1/tier-2 stream this run), use the combined path+title lookup from Task 1 for the destination directory and filename title, falling back to the existing local-title/empty-path behavior only when the lookup has no entry for that movie/series. Verify with a unit test: synthesis with a lookup hit produces the same destination a tier-1 stream for that movie/series would produce; synthesis with no lookup hit falls back unchanged from today's behavior.

## 3. Fix tier-2 filename title source (design D4)

- [x] 3.1 In `buildTier2MovieStreams`, pass the already-fetched `radarrMovie.Title` (not `movie.TMDBTitle`) into `downloader.BuildRadarrDestPath`. Verify with a unit test: a movie whose local `TMDBTitle` differs from the live Radarr title produces a tier-2 destination filename matching the live Radarr title.
- [x] 3.2 In `buildTier2SeriesStreams`, pass the already-fetched `series.Title` (not `ep.TMDBTitle`) into `downloader.BuildSonarrDestPath`. Verify with an equivalent series unit test.

## 4. Add the `file_missing` reconciliation outcome (design D3)

- [x] 4.1 Add `OutcomeFileMissing ReconcileOutcome = "file_missing"` in `internal/api/resync_path.go`.
- [x] 4.2 In `reconcileDownloadPath`, when `computeReconciledPath` reports `changed=false` (root already matches), `os.Stat` the stored `download_path` before returning `OutcomeAlreadyUpToDate`; return `OutcomeFileMissing` instead when the file does not exist, making no database change either way. Verify with a unit test covering both branches (file present vs. absent) on a root-matches case.
- [x] 4.3 In `ReconcileScheduledDownloadPaths`, log `OutcomeFileMissing` as a WARN distinguishable from the existing `rename_failed` log line (distinct message text, same structured fields). Verify with a test asserting the log output/fields for a `file_missing` case differ from a `rename_failed` case.
- [x] 4.4 In `resyncDownloadPath` (the `POST /api/v1/downloads/:id/resync-path` handler), return `200 OK` with `{"status": "file_missing"}` for this outcome, alongside the existing `corrected`/`already_up_to_date`/`not_managed_by_radarr_sonarr` responses. Verify with an API handler test asserting the response body and status code for a download whose root matches but whose file is missing.

## 5. Full verification

- [x] 5.1 Run `go test ./...` and confirm all existing and new tests pass.
- [x] 5.2 Run `openspec validate fix-duplicate-media-downloads --strict` and confirm it reports the change as valid.
- [x] 5.3 Manually exercise the fixed `download` command against a test/staging Radarr+Sonarr+M3U setup with a deliberately duplicated playlist entry for one movie already marked downloaded, and confirm no second download is attempted, matching the `media-download-scheduling` tier-1-precedence requirement this change restores.
