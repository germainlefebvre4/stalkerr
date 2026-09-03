## 1. Shared root-reconciliation logic

- [x] 1.1 Add a package function (e.g. in `internal/downloader` or a new small file alongside `destpath.go`) that, given a completed download's stored path and its content type (movie/TV), extracts the current root directory using the existing `filepath.Dir` (movie) / `detectTVSeasonPath` (TV, exported or reused from `internal/api`) logic, and returns it alongside the sub-path to preserve. Verify with unit tests covering a movie path, a TV episode path under `Season NN`, and a path that doesn't match either shape.
- [x] 1.2 Add a function that, given the extracted root and a target root (basename + optional different parent) from Radarr/Sonarr's current `Path`, computes whether they differ and, if so, the full corrected destination path (root replaced, sub-path preserved). Verify with unit tests: same root (no-op), renamed root under the same parent, root moved to a different parent entirely.
- [x] 1.3 Add a function that performs the actual correction: move the single file via the existing `moveSingleFile`/collision-check logic (reused from the rename handler) and update `download_info.download_path` in a DB transaction, returning a typed outcome (`corrected`, `already_up_to_date`, `rename_target_exists`, `rename_failed`). Verify with a test that moves a file and updates the DB row, and a test that a pre-existing destination file blocks the move without modifying the DB.

## 2. Scheduled reconciliation (auto)

- [x] 2.1 In the `download` command's run (`internal/scheduler` / `cmd/download.go`), add a new per-run `GetAllMovies`/`GetAllMonitoredSeries` fetch (distinct from the existing `GetMissingMovies`/`GetMissingEpisodes` calls, which only cover not-yet-downloaded content) and build `map[tmdbID]moviePath` and `map[tvdbID]seriesPath` lookups from the responses.
- [x] 2.2 Iterate completed `download_info` rows whose associated `Movie.TMDBID`/`TVShow.TVDBID` is present in the corresponding map, and apply the correction from 1.2/1.3 when the stored root differs from the map's `Path`. Rows with no match in the map SHALL be left untouched (population-2 downloads, per proposal.md). Verify with a test seeding a completed download whose stored root doesn't match a fake Radarr/Sonarr library fixture, asserting `download_path` is corrected after the run, and a second download with no library match asserting it is untouched.
- [x] 2.3 Verify the run tolerates a Radarr/Sonarr fetch failure by skipping reconciliation for that run without failing the run's primary scheduling work (per design.md's risk mitigation).

## 3. On-demand resync endpoint

- [x] 3.1 Add `POST /api/v1/downloads/:id/resync-path` in `internal/api/handlers_frontend.go` and register the route in `internal/api/api.go`. It SHALL: reject a download with no `download_path` (mirroring `renameDownload`'s validation), perform a live `GetMovieByTMDBID`/`GetSeriesByTVDBID` lookup for the download's associated `Movie`/`TVShow`, and return `"not_managed_by_radarr_sonarr"` when the lookup finds nothing.
- [x] 3.2 Wire the lookup result into the correction function from 1.2/1.3, returning the three success outcomes (`corrected` with old/new path, `already_up_to_date`, `not_managed_by_radarr_sonarr`) plus the existing `rename_target_exists`/`rename_failed` error codes. Verify with handler tests covering: path corrected, path already correct, item not found in Radarr/Sonarr, incomplete download rejected, and destination collision blocked.
- [x] 3.3 Add the endpoint call to `frontend/src/services/api.ts`.

## 4. Sidepanel UI

- [x] 4.1 Add a "Resynchroniser" button to the Downloads tab's details sidepanel (`frontend/src/components`), visible only when the download's `status` is `completed`, alongside the existing "Déplacer"/"Renommer" actions.
- [x] 4.2 On click, call the endpoint from 3.3 and show a translatable outcome message (corrected / already up to date / not managed by Radarr/Sonarr), updating the displayed folder/path in place on success without a full list reload. Add locale keys to `frontend/src/locales/{en,fr}/downloads.json`.
- [x] 4.3 Verify via component test that the action is absent for a non-completed download, and confirm manually (or via component test) that it is not rendered in the Erreurs tab's dedicated sidepanel.

## 5. End-to-end verification

- [x] 5.1 Run the full backend test suite (`go test ./...`) and frontend test suite, confirming no regressions in existing rename/move/enrichment tests.
- [x] 5.2 Manually reproduce the original scenario: complete a download, rename its parent folder on disk and update the matching Radarr/Sonarr entry's path to match, trigger the on-demand resync from the sidepanel, and confirm the Erreurs tab's `missing_year`/etc. status reflects the corrected file immediately afterward.
