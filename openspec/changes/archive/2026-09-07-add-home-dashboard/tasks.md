## 1. Backend: Persist Per-Run Statistics

- [x] 1.1 Add a `StringList` type (or equivalent) implementing `sql.Scanner`/`driver.Valuer` that marshals `[]string` to/from a JSON array stored in a `TEXT` column, in `internal/models` (per design.md Decision 2), and verify it round-trips via a unit test (encode a slice, `Value()`, `Scan()` it back, compare).
- [x] 1.2 Add `MoviesCount`, `TVShowsCount`, `NewItemsCount`, `TMDBMatchedCount`, `TMDBUnmatchedCount *int` and `GroupTitles StringList` (nilable/JSON-null-capable) fields to `models.ProcessingLog` (`internal/models/log.go`), and verify `go build ./...` succeeds and `AutoMigrate` (via the existing test suite's SQLite bootstrap in `internal/testutil/helpers.go`) creates the table without error.
- [x] 1.3 Extend `processor.Statistics` (`internal/processor/processor.go`) with `NewItems int` and `GroupTitles map[string]struct{}`, and verify `go build ./...` succeeds.
- [x] 1.4 In `saveBatch` (`processor.go:643-689`), increment `stats.NewItems` in the create branch (alongside the existing `err == gorm.ErrRecordNotFound` case) and add each line's `GroupTitle` to `stats.GroupTitles`, and verify with a processor test asserting `NewItems` counts only newly-created lines when a batch mixes new and force-updated lines.
- [x] 1.5 In `updateProcessingLog` (`processor.go:692-701`), populate the new `ProcessingLog` fields from `stats` (movies, TV shows, new items, TMDB matched/unmatched, and the group title set converted to a sorted `[]string`), and verify with a processor test asserting a completed run's `processing_logs` row has all new fields populated matching the run's actual content.
- [x] 1.6 Verify a run that fails partway (e.g. parse error after some batches saved) still persists the statistics accumulated up to that point, per the `processing-run-statistics` spec's "run that fails partway" scenario — add/extend a processor test for this.
- [x] 1.7 Verify a run that processes zero items (e.g. every line is a duplicate and skipped) persists an empty group title list (`[]`, not `null`) and zero counts, per the spec's "no processed items" scenario.

## 2. Backend: Expose Statistics via the API

- [x] 2.1 Confirm `listProcessingLogs` (`internal/api/handlers_frontend.go:27-58`) needs no handler changes since it serializes `models.ProcessingLog` directly (per design.md Context) — verify by running `GET /api/v1/processing-logs` against a test DB seeded with a run that has the new fields set, and asserting the JSON response includes them.
- [x] 2.2 Verify a `processing_logs` row created before this migration (new columns left as their zero/null DB value) serializes as `null`/absent for the new fields in the API response, not `0` or `[]`, per the `api-processing-logs` delta spec.

## 3. Frontend: Navigation and Default Tab

- [x] 3.1 Add `'home'` to `VALID_TABS`, change `TAB_URL_SCHEMA.tab.default` and `readInitialActiveTab`'s final fallback from `'playlist'` to `'home'`, and change `MOBILE_FALLBACK_TAB` from `'playlist'` to `'home'` in `frontend/src/App.tsx`, and verify by checking `App.test.tsx` (extended in task 8.1) confirms Home is active with no URL param and no `localStorage` value.
- [x] 3.2 Extend the mobile-narrowing fallback `useEffect` (`App.tsx:149-153`, currently `activeTab === 'errors'`) to also cover `activeTab === 'filters'`, and verify with a test that narrowing the viewport while "Filtres" is active switches `activeTab` to `'home'`.
- [x] 3.3 Remove the `filters` entry from the mobile `tabs[]` array (`App.tsx:235-241`) and add a `home` entry first; add a `home` `Tabs.Trigger` first in the desktop `Tabs.List` (`App.tsx:263-270`), keeping the existing `filters` trigger there unchanged (desktop-only-in-bottom-bar, still selectable from desktop tabs), and verify by rendering `App` at a mobile viewport width and asserting the bottom tab bar shows Home/Playlist/Traitements/Téléchargements/Arr-Suite and not Filtres.
- [x] 3.4 Add `'tabs.home'` translation keys to `frontend/src/locales/{fr,en}/common.json`, and verify both locale files remain valid JSON and the key renders in both the desktop trigger and the mobile tab bar label.

## 4. Frontend: Home Tab Data Layer

- [x] 4.1 Add a `useHomeDashboard` hook (or equivalent) gated on `activeTab === 'home'` (mirroring `useDownloads`/`useErrorsTab`'s activation pattern) that fetches: the latest processing log (`GET /api/v1/processing-logs?limit=1&offset=0`), the downloads total (`GET /api/v1/downloads?limit=1`), and the errors total (`GET /api/v1/downloads?problem=missing_year,year_mismatch,unknown_format&limit=1`), reusing `api.ts` methods (adding thin wrappers there if none exist for a bare processing-logs/downloads count fetch), and verify with a hook test asserting each endpoint is called only when `activeTab === 'home'`.
- [x] 4.2 Verify `useHealthAndStats` (`/api/v1/stats`) and `useRadarrSonarr`'s stats fetch (`/api/v1/radarr-sonarr/stats`) are reachable from `App.tsx` for the Home tab without duplicating their poll loops — wire the Home tab to consume `stats`/`radarrSonarrStats` already held in `App` state, requesting `useRadarrSonarr`'s stats fetch on Home-tab activation the same way it's requested on `radarr-sonarr`-tab activation today.
- [x] 4.3 Compute elapsed run duration from `started_at`/`completed_at` (or "in progress" when `completed_at` is null) in a small utility (e.g. extend `frontend/src/utils/date.ts`), and verify with a unit test covering a completed run, an in-progress run, and formatting stability (matching the existing `DD/MM/YYYY` fixed-format convention used elsewhere per `frontend-ihm-dashboard`).

## 5. Frontend: Home Tab UI

- [x] 5.1 Create `frontend/src/components/HomeTab.tsx` rendering the four sections (last-run summary, catalog overview, Radarr/Sonarr summary, downloads/errors summary) per the `frontend-home-dashboard` spec, and mount it as the first `Tabs.Content` in `App.tsx`, and verify by rendering the Home tab in a component test and asserting all four sections appear.
- [x] 5.2 Render the last-run summary's empty state when `GET /api/v1/processing-logs` returns no entries, and a distinct "statistics unavailable" placeholder (not `0`/`[]`) when the latest entry's new fields are `null`, and verify both with component tests.
- [x] 5.3 Render the in-progress state (no fixed duration) when the latest run's `status` is `in_progress`, and verify with a component test.
- [x] 5.4 Render the Radarr/Sonarr summary's per-service error state (e.g. `radarr_error: "radarr_unreachable"`) independently of the other service still rendering its count, and verify with a component test mirroring `RadarrSonarrTab.test.tsx`'s existing error-state coverage.
- [x] 5.5 Apply mobile responsive card density (stacked cards, `44px` touch targets, reduced padding/font below the mobile breakpoint) to the Home tab's sections in `frontend/src/index.css`/`variables.css`, and verify by rendering at a sub-768px width and asserting no horizontal scroll container is needed and cards stack vertically.
- [x] 5.6 Add `frontend/src/locales/{fr,en}/home.json` (or extend `common.json`) with all Home tab copy (section titles, empty states, error states), register the namespace in `frontend/src/i18n/index.ts` if a new file is used, and verify both locale files are valid JSON and every string used by `HomeTab.tsx` resolves via `t(...)`.

## 6. Frontend: Remove the KPI Banner

- [x] 6.1 Remove the `<StatsKPICards ... />` render from `App.tsx` and delete `frontend/src/components/StatsKPICards.tsx` (and its now-unused `kpi.*` locale keys, after confirming `HomeTab.tsx` uses its own keys per task 5.6), and verify `go`/`npm` builds are unaffected (`npm run build` in `frontend/`) and no remaining import references `StatsKPICards`.
- [x] 6.2 Remove the page-level `kpi-grid`/`kpi-toggle-row` CSS rules and their mobile collapse/expand styling from `frontend/src/index.css`, keeping any class names still used by `HomeTab.tsx`'s own cards renamed/scoped separately, and verify visually (via `run` skill or manual dev-server check) that no orphaned collapsed toggle row renders on any non-Home tab.
- [x] 6.3 Update/remove any test asserting the old KPI banner's presence across tabs (search `StatsKPICards`, `kpi-grid`, `kpi-toggle-row` in `frontend/src/**/*.test.tsx`), and verify the full frontend test suite passes (`npm test` in `frontend/`).

## 7. Verification

- [x] 7.1 Run the full backend test suite (`go test ./...`) and confirm it passes, including the new/updated processor and API tests from sections 1-2.
- [x] 7.2 Run the full frontend test suite (`npm test` in `frontend/`) and confirm it passes, including the new/updated tests from sections 3-6.
- [x] 7.3 Using the `run` skill, launch the app, trigger an M3U processing run, and confirm the Home tab shows that run's actual movies/TV shows/new-items/TMDB matched-unmatched counts and group title list, matching what was just processed.
- [x] 7.4 In the running app, confirm Home is the default tab on a fresh load (no `tab` param, cleared `localStorage`), confirm "Filtres" is absent from the mobile bottom tab bar but still reachable from the desktop tab list, and confirm narrowing the viewport while "Filtres" is active falls back to Home.
