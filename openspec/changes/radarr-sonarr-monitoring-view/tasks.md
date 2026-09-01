## 1. Radarr/Sonarr client: expose full monitored lists

- [x] 1.1 Add `radarr.Client.GetAllMovies` (or equivalent) hitting `/api/v3/movie`, returning lightweight metadata for every movie regardless of missing/hasFile status; add a unit test in `internal/external/radarr/radarr_test.go` verifying parsing and that `GetMissingMovies` behavior is unchanged.
- [x] 1.2 Add a Sonarr client method returning all monitored series (reusing the existing full-series fetch already inside `GetMissingSeries`, without the missing-episodes filter); add a unit test in `internal/external/sonarr/sonarr_test.go` verifying it returns monitored series with and without missing episodes, and that `GetMissingSeries` behavior is unchanged.
- [x] 1.3 Run `go test ./internal/external/radarr/... ./internal/external/sonarr/...` and verify all existing and new tests pass.

## 2. Matcher: state-agnostic, batched lookups

- [x] 2.1 Add a batched movie-match helper in `internal/matcher` that, given a set of Radarr movies (TVDB/TMDB ids, title, year), issues one batched DB query to find existing local `Movie` matches (TVDB/TMDB id first, fuzzy title+year fallback for the remainder) - no `ProcessedLine.state` filter. Add unit tests in `internal/matcher/matcher_test.go` covering TVDB hit, TMDB-only hit, fuzzy-only hit, and no match.
- [x] 2.2 Add a state-agnostic playlist-occurrence lookup (all `ProcessedLine` states, not just `processed`/`failed`) for a given movie ID and for a given TV show ID. Add a unit test proving a `downloaded`-state occurrence is included, contrasting with `FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates` which must remain unchanged and keep excluding it.
- [x] 2.3 Add a per-series aggregate helper: given a series' TVDB id and its list of monitored (season, episode) pairs, issue one batched local `TVShow` query (not one per episode) and return matched-count/monitored-count. Add unit tests for zero monitored episodes, partial match, and full match.
- [x] 2.4 Run `go test ./internal/matcher/...` and verify all existing and new tests pass.

## 3. Backend API endpoints

- [x] 3.1 Add `GET /api/radarr/movies?page&limit` handler: fetch the full Radarr movie list, sort by title, slice to the requested page (default and max page size enforced), run the batched movie-match helper (2.1) only on that page, and return paginated results with per-movie match status. Register the route alongside existing routes in `internal/api`.
- [x] 3.2 Add `GET /api/sonarr/series?page&limit` handler: fetch the full Sonarr series list, slice to the requested page, then for that page's series concurrently fetch each series' monitored episodes from Sonarr (bounded by page size) and compute the aggregate via helper 2.3. Register the route.
- [x] 3.3 Add `GET /api/radarr/movies/:id/matches` detail handler returning the matched local `Movie` metadata (if any) and its full occurrence list via helper 2.2, distinguishing "no local match" from "matched but zero occurrences".
- [x] 3.4 Add `GET /api/sonarr/series/:id/episodes` detail handler returning, per monitored episode, whether a matching local `TVShow`/occurrence exists (feeding the sidepanel's per-episode breakdown), using the same state-agnostic lookup as 2.2.
- [x] 3.5 In all four handlers, distinguish "service not configured" (no URL/API key) from "upstream unreachable/error" in the response, and verify with a handler test that a Radarr failure does not affect the Sonarr endpoints (independent goroutines/requests, no shared failure state).
- [x] 3.6 Add a handler test asserting that requesting a page from a catalog larger than the page size only triggers matching/DB work bounded by `limit` (e.g. via a call-counting fake Radarr/Sonarr client or DB query count assertion), proving the pagination-before-matching guarantee from the spec.
- [x] 3.7 Run `go test ./internal/api/...` and verify all tests pass.

## 4. Frontend: types and API client

- [x] 4.1 Add TypeScript types in `frontend/src/types.ts` for the Radarr movie monitoring list item, Sonarr series monitoring list item (with matched/monitored counts), and the two detail-response shapes.
- [x] 4.2 Add corresponding methods to `frontend/src/services/api.ts` (`listRadarrMovies`, `listSonarrSeries`, `getRadarrMovieMatches`, `getSonarrSeriesEpisodes`) following the existing method conventions (e.g. `searchTMDB`, `forceDownload`).

## 5. Frontend: new tab

- [x] 5.1 Register a new tab trigger/value (e.g. `radarr-sonarr`) in `frontend/src/App.tsx` next to the existing `playlist`/`filters`/`logs`/`downloads` triggers, and create `frontend/src/components/RadarrSonarrTab.tsx` as the panel component.
- [x] 5.2 Implement the Films section: paginated list from `listRadarrMovies`, a match/no-match badge per row, its own loading/error state, and an explicit "Actualiser" action (also triggered on first mount) - no polling.
- [x] 5.3 Implement the Séries section with the same independent loading/error/refresh pattern, showing the matched/monitored ratio per series.
- [x] 5.4 Verify independence: simulate the Films section's request failing (e.g. temporarily point Radarr config at an invalid URL) and confirm the Séries section still loads normally, and vice versa.
- [x] 5.5 Add the sidepanel (Radix `Dialog`, reusing the visual pattern from `PlaylistTab.tsx`'s drawer) for a selected movie: matched local metadata when matched, full occurrence list (resolution + state, including `downloaded`) when matched, and a clear "no match" state when unmatched.
- [x] 5.6 Extend the sidepanel for a selected series: per-episode match breakdown backed by `getSonarrSeriesEpisodes`.
- [x] 5.7 Add `radarrSonarr.json` to `frontend/src/locales/fr/` and `frontend/src/locales/en/` with all strings used by the new tab and sidepanel, and register the namespace wherever `playlist.json` etc. are wired into i18n.

## 6. Verification

- [x] 6.1 Run the frontend test suite/lint and the Go test suite end to end and confirm both are green.
- [ ] 6.2 Launch the app (with Radarr/Sonarr configured against a real or test instance) and manually walk through: opening the new tab, paginating both sections, refreshing each independently, triggering an error in one section without affecting the other, and opening the sidepanel for a matched movie, an unmatched movie, and a series.
