## 1. Radarr/Sonarr existence lookups

- [x] 1.1 Add `radarr.Client.GetMovieByTMDBID(ctx, tmdbID int) (*Movie, error)` calling `GET /api/v3/movie?tmdbId=`, returning a distinguishable not-found result (empty array) versus a transport/HTTP error; verify with a unit test covering found, not-found, and error-response cases.
- [x] 1.2 Add `sonarr.Client.GetSeriesByTVDBID(ctx, tvdbID int) (*Series, error)` calling `GET /api/v3/series?tvdbId=`, with the same found/not-found/error distinction; verify with a unit test.
- [x] 1.3 Add `sonarr.Client.GetEpisodesBySeriesID(ctx, seriesID int) ([]Episode, error)` calling `GET /api/v3/episode?seriesId=`; verify with a unit test using a fixture with multiple seasons/episodes.
- [x] 1.4 Add a helper that combines 1.2+1.3 to answer "does series X have an episode SxxEyy" (filtering client-side on `SeasonNumber`/`EpisodeNumber`), returning found/not-found/error; verify with a unit test for a matching episode, a missing episode number, and a missing series.

## 2. Shared destination naming (extract, then extend)

- [x] 2.1 Move `sanitizeFilename`, `buildRadarrDestPath`, and `buildSonarrDestPath` out of `cmd/format.go` into an internal package reachable from both `cmd` and `internal/api` (e.g. `internal/downloader`); update `cmd/format.go` call sites and move/adjust `cmd/pathutil_test.go` accordingly; verify existing tests still pass unchanged in behavior.
- [x] 2.2 Add a resolution-suffixed variant of the movie/TV base-path builders (suffix like `[1080p]`, with a unique fallback marker when resolution is `NULL`, per design Decision 3) used only by the force-download path; verify with unit tests that: two different resolutions of the same title produce different paths, and a `NULL`-resolution occurrence still produces a path distinct from any sibling.
- [x] 2.3 Verify (unit test) that the existing automatic-pipeline call sites (`cmd/radarr.go`, `cmd/sonarr.go`) still produce byte-for-byte the same unsuffixed destination path as before this change.

## 3. Resume path-resolution fix

- [x] 3.1 Reorder `ResumeHelper.buildBaseDestPath` to prefer a non-empty `download.DownloadPath`-derived base over recomputing from `line.Movie`/`line.TVShow` metadata, falling back to recomputation only when no path was ever recorded (design Decision 4); verify with a unit test that a `DownloadInfo` with a pre-set suffixed `DownloadPath` resumes into that same path rather than a recomputed unsuffixed one.
- [x] 3.2 Verify (unit test) that a `DownloadInfo` with no `DownloadPath` yet (a download that never started) still falls back to recomputing from `line.Movie`/`line.TVShow` exactly as before.

## 4. Force-download trigger endpoint

- [x] 4.1 Add `POST /api/v1/items/:id/force-download` route and handler skeleton in `internal/api`, wired to construct Radarr/Sonarr clients from `config.Get()` (mirroring the existing CLI construction pattern); verify the route is reachable and returns a structured response.
- [x] 4.2 Implement the local eligibility checks in order (item exists → matched to a movie/TV show → not already `downloaded`/`downloading`), returning a 4xx with a clear reason and no side effect on failure, without contacting Radarr/Sonarr (spec: "no existence check outside an actual trigger"); verify with unit/integration tests for each refusal case.
- [x] 4.3 Implement the live existence check under a short timeout: movie → `GetMovieByTMDBID`; TV episode → the 1.4 helper, refusing (fail-closed) when the TVShow's `TVDBID` is `nil`, when not found, or when the call errors/times out; verify with tests for found, not-found, and simulated-timeout/error cases, asserting no `DownloadInfo` row is created on refusal.
- [x] 4.4 On success, compute the resolution-suffixed destination base path (using the just-fetched `movie.Path`/series path as root, per design Decision 2/3) and persist it onto a newly created `DownloadInfo` before returning, then start the transfer in a background goroutine via the existing `Downloader.Download()`; return `202 Accepted` immediately; verify with an integration test that the HTTP response returns before the (mocked/slow) transfer completes.
- [x] 4.5 Verify (integration test) that a second forced-download request for the same occurrence while the first is still in flight is rejected via the existing per-`DownloadInfo` lock, without starting a duplicate transfer.
- [x] 4.6 Verify (integration test) end-to-end that requesting a forced download for an occurrence whose sibling occurrence is already `downloaded` succeeds and does not alter the sibling's state or file.

## 5. Sidepanel "Forcer le téléchargement" action

- [x] 5.1 Add the API call for the new endpoint in `frontend/src/services/api.ts`; verify with a type-check/build.
- [x] 5.2 Add the action to the drawer in `frontend/src/components/PlaylistTab.tsx`, enabled only when `selectedItem` is matched to a movie/TV show and its state is neither `downloaded` nor `downloading`; verify manually that the button is absent/disabled for an unmatched or already-downloaded item and present for an eligible one.
- [x] 5.3 Wire the click handler to show a queued/in-progress acknowledgement on `202`, and a distinct, explicit error message on any refusal response, without blocking or closing the drawer either way; verify manually for both a success and a refusal (e.g. by pointing at a movie absent from Radarr).
- [x] 5.4 Add the new drawer strings to the fr/en i18n resources; verify both locales render without missing-key warnings.

## 6. End-to-end verification

- [x] 6.1 Run the full backend test suite (`go test ./...`) and confirm no regression in existing `radarr`/`sonarr`/`resume-downloads` command tests.
- [ ] 6.2 Manually reproduce the original scenario: a movie with two ingested occurrences where one is already downloaded — force-download the other from the sidepanel, confirm both files end up on disk under distinct resolution-suffixed names, and confirm the already-downloaded occurrence's state/file is untouched.
