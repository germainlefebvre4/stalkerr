# Tasks

## 1. Sonarr series counters (verify first)

- [x] 1.1 Fetch a real `/api/v3/series` payload from a Sonarr instance (user-provided or via a local run) and record where `episodeFileCount`, `episodeCount` and `totalEpisodeCount` live (top level vs `statistics`); verify by saving a redacted sample as a test fixture under `internal/external/sonarr/testdata/`
- [x] 1.2 Update `sonarr.Series` in `internal/external/sonarr/sonarr.go` to parse the counters from the real shape (`statistics` with top-level fallback) and expose the aired-episode count; verify with a unit test that decodes the fixture and one that decodes the legacy flat shape
- [x] 1.3 Adjust `GetMissingSeries` (and its tests) to the corrected counters if it relies on the old fields; verify `go test ./internal/external/sonarr/...` passes

## 2. Backend: file state and filters

- [x] 2.1 Add a file-state helper in `internal/api/radarr_sonarr_handlers.go` (movie/episode: downloaded|none; series: none|partial|downloaded over aired episodes, `aired == 0` gives none, `files > aired` gives downloaded); verify with a table-driven unit test covering all boundaries
- [x] 2.2 Replace `missing`/`has_file` with `file_state` on movie, series and episode response structs, and add `episode_file_count`/`episode_count` to series entries; verify with handler tests asserting the JSON for each of the three endpoints
- [x] 2.3 Narrow `status` to `monitored|unmonitored` with OR semantics (default `{monitored}`, unrecognized values including `missing` ignored) and add the repeatable `file` parameter (OR within, AND with `status`, no restriction when absent); verify with handler tests for single values, OR combinations, `status=monitored&file=none&file=partial`, `status=missing` ignored, and combination with `search` and match-status
- [x] 2.4 Confirm the file filter adds no per-item upstream call and is applied before pagination, including the Sonarr match-status-cache path; verify with a test asserting the upstream mock call counts and page totals
- [x] 2.5 Run `go test ./internal/api/... ./internal/external/...` and `go vet ./...`; verify they pass

## 3. Frontend: types, state, API client

- [x] 3.1 Update `frontend/src/types.ts`: add `FileState` and `FileFilter`, narrow `EtatStatus` to Suivi values, replace `missing` with `file_state` (and series counts) on `RadarrMovieListItem`, `SonarrSeriesListItem`, `SonarrSeriesEpisodeItem`; verify with `npm run typecheck` (or the project's tsc script)
- [x] 3.2 Update `frontend/src/hooks/useRadarrSonarr.ts`: narrow `filmsStatus`/`seriesStatus` validation (invalid values such as `missing` fall back to `monitored`) and add `filmsFile`/`seriesFile` URL state with setters; verify in `useRadarrSonarr.test.ts` (default, round-trip, obsolete `missing` URL, pagination reset)
- [x] 3.3 Update `frontend/src/services/api.ts` to send repeated `status` and `file` params for movies and series; verify with an api unit test on the built query string

## 4. Frontend: UI

- [x] 4.1 Generalize `EtatFilterDropdown` into a neutral multi-select dropdown (rename the file/component and its CSS classes accordingly) and update its test; verify the test passes
- [x] 4.2 In `RadarrSonarrTab.tsx` replace the État dropdown with Suivi and Fichier dropdowns on both sections (Films: Téléchargé/Absent; Séries: Téléchargé/Partiel/Absent) plus the "Manquants" preset button with active state; verify in `RadarrSonarrTab.test.tsx` (defaults, selection triggers a fetch from page 1, preset sets both filters, preset active state)
- [x] 4.3 Replace the État badge with a Suivi column and Fichier badge (Absent red when monitored, grey otherwise; Partiel orange with ratio for series) in both tables, the sidepanel header and episode rows, via shared render helpers; verify in component tests for each color/state combination
- [x] 4.4 Render both indicators as dots through `renderStatusIndicator` on mobile (cards, header, episode rows) and check the filter bar wraps without horizontal scroll; verify in mobile-mode component tests and by viewing the page at phone width
- [x] 4.5 Update `frontend/src/locales/{fr,en}/radarrSonarr.json`: remove `etat` keys, add Suivi/Fichier/Manquants labels (FR: Surveillé, Non surveillé, Téléchargé, Partiel, Absent, Manquants); verify with the frontend i18n key-parity check or a test and that no `etat` key is still referenced
- [x] 4.6 Run the frontend lint, typecheck and test suite; verify they pass

## 5. Integration checks and docs

- [x] 5.1 Run the app against a real or mocked Radarr/Sonarr and check Films and Séries end to end: default view unchanged, Téléchargé filter, Manquants preset, unmonitored + downloaded items, a continuing series reading Téléchargé when all aired episodes are present
- [x] 5.2 Add a `CHANGELOG.md` entry covering the Suivi/Fichier split, the removal of Missing/`missing`, and (if task 1.1 shows it) the series counter parsing fix; verify the entry follows the file's existing format
