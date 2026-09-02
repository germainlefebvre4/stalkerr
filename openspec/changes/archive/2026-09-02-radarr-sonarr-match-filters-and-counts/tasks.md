## 1. Backend - Radarr full-catalog filtering

- [x] 1.1 Add an optional `filter` query parameter (`matched` / `no_match`) to `listRadarrMonitoredMovies`; when present, run `matcher.MatchMoviesBatch` over the full monitored (post-search) list instead of the page, filter by status, then paginate. Verify with a handler test covering `filter=matched`, `filter=no_match`, and no filter (unchanged behavior).
- [x] 1.2 Verify `total`/`TotalPages` reflect the filtered count, not the unfiltered catalog size, when `filter` is set. Add a test with a catalog spanning multiple pages pre-filter but fewer post-filter.
- [x] 1.3 Verify `filter` combines with `search` (both applied before pagination). Add a test exercising both together.

## 2. Backend - Radarr occurrence count

- [x] 2.1 Add a batched query (e.g. `processed_lines` grouped by `movie_id`) that returns occurrence counts for a given set of matched movie IDs. Verify with a unit test covering a movie with multiple occurrences and a movie with none.
- [x] 2.2 Add an `occurrence_count` field to `RadarrMovieListItem` (Go) and populate it in `listRadarrMonitoredMovies` for the page's matched movies (0 for unmatched). Verify via handler test asserting the field on a known fixture.
- [x] 2.3 Add `occurrence_count` to the corresponding frontend type in `frontend/src/types.ts`.

## 3. Backend - Sonarr match-status cache

- [x] 3.1 Implement an in-memory cache (`series_id -> matched bool`) scoped to the server process, with a lookup/populate function that fetches a series' episodes from Sonarr and computes `matched = (matched episodes > 0)` on a miss. Verify with a unit test covering miss-then-populate and hit-without-refetch.
- [x] 3.2 Wire a cache-invalidation call into the existing Séries manual-refresh path (frontend refresh trigger -> backend request that clears the relevant cache entries). Verify with a test that a refresh causes the next filtered request to recompute rather than reuse a stale cached value.
- [x] 3.3 Add an optional `filter` query parameter (`matched` / `no_match`) to `listSonarrMonitoredSeries`; when present, resolve each monitored series' matched status via the cache (populating on miss) instead of computing aggregates only for the page, filter by status, then paginate. Verify with a handler test covering `filter=matched`, `filter=no_match`, and no filter (existing per-page aggregate behavior unchanged).
- [x] 3.4 Verify `total`/`TotalPages` reflect the filtered count for Sonarr, and that `filter` combines correctly with `search`.

## 4. Backend - Sonarr occurrence count

- [x] 4.1 Switch the per-page series computation in `listSonarrMonitoredSeries` from `matcher.MatchSeriesEpisodesAggregate` to `matcher.MatchSeriesEpisodesDetail` for the page's series, and sum `len(Occurrences)` across each series' monitored episodes into an `occurrence_count`. Verify with a test asserting the summed count against a fixture with episodes having varying occurrence counts.
- [x] 4.2 Add `occurrence_count` to `SonarrSeriesListItem` (Go) and the corresponding frontend type.

## 5. Frontend - filter controls

- [x] 5.1 Add a match-status filter select (All / Matched / No match) to the Films section in `RadarrSonarrTab.tsx`, following the `DownloadsTab.tsx` dropdown pattern.
- [x] 5.2 Add the same filter control to the Séries section.
- [x] 5.3 Persist both filters in the URL via `useURLState`/`patchURLState` (extending `useRadarrSonarrView.ts` or a sibling hook, matching `useDownloads.ts`'s pattern), so a filter selection survives a page refresh.
- [x] 5.4 Wire the filter state into `useRadarrSonarr.ts`'s fetch calls (query param), re-fetching from the first page when the filter changes, mirroring the existing debounced-search re-fetch behavior.
- [x] 5.5 Manually verify in the browser: selecting "No match" / "Matched" on both tables filters correctly, combines with an active search term, and survives a page refresh.

## 6. Frontend - occurrence count column

- [x] 6.1 Add an "Occurrences" column to the Films table (desktop and mobile card layout) rendering `occurrence_count`.
- [x] 6.2 Add an "Occurrences" column to the Séries table (desktop and mobile card layout) rendering `occurrence_count`.
- [x] 6.3 Add i18n strings for the new column header and filter labels in all locales the app currently supports.
- [x] 6.4 Manually verify in the browser: occurrence counts render correctly for a matched item with duplicate-quality occurrences, an unmatched item shows 0, and mobile card layouts display the new value legibly.

## 7. Verification

- [x] 7.1 Run the full backend test suite and confirm no regression in existing Radarr/Sonarr handler and matcher tests.
- [x] 7.2 Run the frontend test suite (if present) and confirm no regression in existing Radarr/Sonarr component tests.
- [x] 7.3 Manually exercise a large monitored Sonarr catalog (or a mock with many series) to confirm the first filtered request's latency is acceptable and a cached second request is fast, per the design's documented latency trade-off.
