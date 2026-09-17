## 1. Backend - Radarr movies listing

- [ ] 1.1 In `internal/api/radarr_sonarr_handlers.go`, replace the hard `if !m.Monitored { continue }` filter in `listRadarrMonitoredMovies` with a status predicate (`monitored`/`unmonitored`/`missing`, ANDed across all values from `c.QueryArray("status")`, defaulting to `["monitored"]` when empty) applied to the full `GetAllMovies` catalog before matching/pagination, and verify with a new backend test that omitting `status` returns exactly the same set as before this change.
- [ ] 1.2 Add `monitored bool` and `missing bool` to the `RadarrMovieListItem` Go struct and populate them from `radarr.Movie.Monitored`/`.HasFile`, and verify with a backend test asserting `missing == true` iff `monitored && !has_file`.
- [ ] 1.3 Add backend tests for `status=missing`, `status=unmonitored`, `status=monitored&status=missing` (non-empty, equivalent to `missing` alone), and `status=unmonitored&status=missing` (empty result, HTTP 200), and verify all pass.
- [ ] 1.4 Add a backend test asserting the status filter combines correctly with the existing `search` and match-status (`matched`/`no_match`) parameters (all constraints applied together).

## 2. Backend - Sonarr series listing

- [ ] 2.1 In `listSonarrMonitoredSeries`, replace the `client.GetAllMonitoredSeries` call with `client.GetAllSeries` (already unfiltered) and apply the same status predicate as 1.1 (reusing a shared helper), defaulting to `["monitored"]`, and verify with a backend test that omitting `status` returns exactly the same set as before this change.
- [ ] 2.2 Add `monitored bool` and `missing bool` to the `SonarrSeriesListItem` Go struct, computing `missing` as `Monitored && EpisodeFileCount < TotalEpisodeCount` (reusing the existing formula from `sonarr.GetMissingSeries`), and verify with a backend test covering a missing series, a complete monitored series, and an unmonitored series.
- [ ] 2.3 Add backend tests mirroring 1.3/1.4 for the Sonarr endpoint (single values, AND combination, contradictory combination, combined with `search` and match-status filter).
- [ ] 2.4 Add a backend test asserting the status filter requires no additional per-series Sonarr episode-list fetches (i.e. it only reads fields already present on `sonarr.Series` from the base `GetAllSeries` call).

## 3. Backend - Sonarr per-episode list

- [ ] 3.1 Add `monitored bool` and `missing bool` (computed as `Monitored && !HasFile`) to the response item backing `/api/v1/sonarr/series/{id}/episodes` (`SonarrSeriesEpisodeItem` or its Go source struct), and verify with a backend test covering a missing episode, a complete monitored episode, and an unmonitored episode.

## 4. Frontend - types and API client

- [ ] 4.1 In `frontend/src/types.ts`, add `monitored: boolean` and `missing: boolean` to `RadarrMovieListItem` and `SonarrSeriesListItem`, and to the Sonarr episode item type used by the drawer, and add an `EtatFilter` type (`Set<'monitored' | 'unmonitored' | 'missing'>` or equivalent) - verify with `tsc`/the frontend build succeeding.
- [ ] 4.2 In `frontend/src/services/api.ts`, extend the Radarr movies and Sonarr series list calls to send the selected État values as repeated `status` query parameters, and verify with a unit test asserting the constructed URL for a multi-value selection.

## 5. Frontend - État filter dropdown component

- [ ] 5.1 Add `@radix-ui/react-dropdown-menu` to `frontend/package.json` and install it, and verify the package installs and the frontend build still succeeds.
- [ ] 5.2 Create a small `EtatFilterDropdown` component (trigger button + `DropdownMenu.Content` with three `DropdownMenu.CheckboxItem`s for Monitored/Unmonitored/Missing) reusable by both the Films and Séries sections, and verify with a component test that toggling a checkbox updates the reported selection without closing the menu, and that the trigger reflects the current selection count/labels.

## 6. Frontend - wire the État filter into Films and Séries sections

- [ ] 6.1 In `frontend/src/hooks/useRadarrSonarr.ts`, add `filmsStatus`/`seriesStatus` URL-persisted state (comma-joined string of the selected values, default `"monitored"`), following the existing `filmsFilter`/`seriesFilter` pattern, and verify with a hook test that the default value is `monitored` when the URL param is absent and round-trips correctly when present.
- [ ] 6.2 In `RadarrSonarrTab.tsx`, render the `EtatFilterDropdown` next to the existing match-status `<select>` in both the Films and Séries filter bars, wired to the new state, and verify manually that changing the selection re-fetches the table from page 1.
- [ ] 6.3 Verify manually (or with a component test) that selecting a contradictory combination (e.g. Unmonitored + Missing) shows the table's existing empty-results state, not an error.

## 7. Frontend - État badges

- [ ] 7.1 Add a small `etatBadgeClass(monitored, missing)` helper (mirroring the existing `seriesBadgeClass` pattern) mapping to `badge-failed` (Missing), `badge-success` (Monitored, complete), `badge-neutral` (Unmonitored), and verify with a unit test covering all three cases.
- [ ] 7.2 Add the État badge as a new column in the Films table (desktop) and Séries table (desktop), alongside the existing Status column, and verify manually that all three États render distinctly.
- [ ] 7.3 Add the État badge to the sidepanel header for both the selected movie and the selected series, next to the title/year, and verify manually.
- [ ] 7.4 Add the État badge to each episode row in the Séries sidepanel's season/episode breakdown (desktop table and mobile card variants), alongside the existing matched/unmatched badge, and verify manually with a series containing at least one missing episode.

## 8. Frontend - mobile compact rendering

- [ ] 8.1 Extend `renderStatusIndicator` (or add an equivalent) so it can render the new État badge as a colored dot + short label on mobile for: Films/Séries mobile list cards, the sidepanel header, and episode rows, and verify manually at a mobile viewport width that no card/row layout overflows or wraps awkwardly.

## 9. i18n

- [ ] 9.1 Add French translation keys for the three État labels, the filter control's own label, and the "no État filter" default state, under the existing `films.*`/`series.*`/`filterStatus.*` namespaces, and verify the frontend build's i18n key-usage check (if any) passes with no missing-key warnings.

## 10. Verification

- [ ] 10.1 Run the full backend test suite (`go test ./...`) and confirm it passes.
- [ ] 10.2 Run the full frontend test suite and confirm it passes, including the extended `RadarrSonarrTab.test.tsx`.
- [ ] 10.3 Manually exercise the feature end-to-end against a real or stubbed Radarr/Sonarr instance: default view unchanged, filter to Missing on both tabs, filter to Unmonitored on both tabs, a contradictory combination showing empty, badges visible in both tables and both drawers (including per-episode), and the same checks repeated at a mobile viewport width.
