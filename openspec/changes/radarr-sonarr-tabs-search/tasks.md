## 1. Backend: search parameter

- [ ] 1.1 In `internal/api/radarr_sonarr_handlers.go`, add a `search` query param read (trimmed, lower-cased) in `listRadarrMonitoredMovies`; filter `monitored` by case-insensitive title substring match before computing `total`/slicing the page. Verify with a new table-driven test case in `radarr_sonarr_handlers_test.go` asserting `total` and returned titles for a `search` term matching a subset, matching nothing, and omitted.
- [ ] 1.2 Apply the same `search` filtering to `listSonarrMonitoredSeries`, filtering `allSeries` before `total`/pagination. Verify with the equivalent test cases for the Sonarr endpoint.
- [ ] 1.3 Run `go test ./internal/api/...` and confirm all pass, including the existing (unfiltered) test cases for both endpoints still pass unchanged.

## 2. Backend: monitoring stats endpoint

- [ ] 2.1 Add `RadarrSonarrStatsResponse` struct (`radarr_monitored`, `radarr_matched`, `sonarr_monitored` fields) and a `listRadarrSonarrStats` handler in `internal/api/radarr_sonarr_handlers.go`: fetch full monitored Radarr list, run `matcher.MatchMoviesBatch` over the entire list (not paginated) to get `radarr_matched`/`radarr_monitored`; fetch full monitored Sonarr list and set `sonarr_monitored` to its length only (no per-series episode fetch). Handle each service independently: if Radarr is unreachable/unconfigured, still return Sonarr's count (and vice versa), using per-field null/omission or a per-service error flag - decide the exact shape while implementing and keep it consistent with the "each upstream reported independently" pattern already used by the listing endpoints.
- [ ] 2.2 Register the route (e.g. `GET /api/v1/radarr-sonarr/stats`) in `internal/api/api.go` next to the existing `/radarr/movies` and `/sonarr/series` routes.
- [ ] 2.3 Add tests in `radarr_sonarr_handlers_test.go` covering: full-catalog matched/unmatched counts with a catalog larger than one page, Radarr unreachable with Sonarr still returned, Sonarr unreachable with Radarr still returned, and confirm no per-series Sonarr episode-list call is made when computing `sonarr_monitored` (e.g. assert on the mock/stub call count). Run `go test ./internal/api/...`.

## 3. Frontend: API client

- [ ] 3.1 In `frontend/src/services/api.ts`, add an optional `search` parameter to `listRadarrMovies` and `listSonarrSeries`, appended to the query string only when non-empty.
- [ ] 3.2 Add a `RadarrSonarrStats` type to `frontend/src/types.ts` matching the new backend response shape, and a `getRadarrSonarrStats(): Promise<RadarrSonarrStats>` method in `api.ts` wrapping the new endpoint. Verify with `npx tsc --noEmit`.

## 4. Frontend: hook state for search and stats

- [ ] 4.1 In `frontend/src/hooks/useRadarrSonarr.ts`, add `filmsSearch`/`setFilmsSearch` and `seriesSearch`/`setSeriesSearch` state; pass the current search term into `fetchFilms`/`fetchSeries`; resetting the respective page to 1 whenever its search term changes (mirroring the existing manual-fetch-on-trigger discipline - search changes are a trigger, not a poll).
- [ ] 4.2 Add `stats`/`statsLoading`/`statsError`/`fetchStats` state, fetched once when the tab becomes active (same `isActive` gating as the existing fetches), calling `api.getRadarrSonarrStats()`.
- [ ] 4.3 Verify with `npx tsc --noEmit` and existing hook tests (add/update a test file if one exists for this hook; otherwise cover via the component test in task 6).

## 5. Frontend: nested tabs, search inputs, résumé panel

- [ ] 5.1 In `frontend/src/components/RadarrSonarrTab.tsx`, wrap the Films and Séries `<section>` blocks in a nested `Tabs.Root`/`Tabs.List`/`Tabs.Trigger` (reusing the `segmented-tabs-list`/`segmented-tabs-trigger` classes from `App.tsx`) with three `Tabs.Content` values: `resume`, `radarr`, `sonarr`. Default to `radarr` (or persist last-selected sub-tab in local state, implementer's choice) on tab activation.
- [ ] 5.2 Move the existing Films section content into the `radarr` `Tabs.Content`, and Séries section content into the `sonarr` `Tabs.Content`, unchanged apart from the addition in 5.3.
- [ ] 5.3 Add a search `<input>` above each table (Films/Séries), bound to `filmsSearch`/`seriesSearch` from the hook, debounced (e.g. 300ms) before triggering `fetchFilms`/`fetchSeries`; clearing the input restores the unfiltered list on the next fetch.
- [ ] 5.4 Build the `resume` `Tabs.Content`: Radarr card showing total monitored + matched/unmatched breakdown from `stats`, Sonarr card showing total monitored only, each with its own loading/error state independent of the other (per design.md - "one summary source failing does not block the other").
- [ ] 5.5 Add an empty-results state (distinct from the existing "no monitored items" empty state) shown when a search term yields zero rows, for both Films and Séries tables.

## 6. i18n

- [ ] 6.1 Add strings to `frontend/src/locales/fr/radarrSonarr.json` and `frontend/src/locales/en/radarrSonarr.json` for: the three sub-tab labels (Résumé/Radarr/Sonarr), the search input placeholder and empty-results message for Films and Séries, and the Résumé panel's labels (total monitored, matched, unmatched, Sonarr total).

## 7. Verification

- [ ] 7.1 Run `npx tsc --noEmit`, `npm run lint`, and `npm test` in `frontend/`, and `go build ./...` and `go test ./...` at the repo root; confirm all green (excluding any pre-existing unrelated failures already present on `main`).
- [ ] 7.2 Manually launch the app: open the Radarr/Sonarr tab, confirm it opens on/switches between the three sub-tabs; in the Radarr sub-tab, search a known movie title and confirm results filter and pagination resets to page 1; clear the search and confirm the full list returns; repeat for the Sonarr sub-tab; open the Résumé sub-tab and confirm the Radarr matched/unmatched counts and the Sonarr total display; confirm the existing sidepanel/drawer still opens correctly from within the Radarr and Sonarr sub-tabs.
