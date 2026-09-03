## 1. Backend: Multi-Value Problem Filter

- [x] 1.1 Update `matchesProblem` (`internal/api/handlers_frontend.go`) to parse `problem` as a comma-separated list and match a download if it satisfies any listed value, keeping single-value requests behaving exactly as today. Verify with a new/updated Go test covering a comma-separated `problem` value.
- [x] 1.2 Verify pagination (`total`/`total_pages`) is computed from the multi-value-filtered set, not the pre-filter set, by adding a test case mirroring the existing single-value pagination test in `internal/api/handlers_frontend_test.go`.
- [x] 1.3 Run `go test ./internal/api/...` and confirm all tests pass, including the new multi-value cases.

## 2. Frontend: Data Layer

- [x] 2.1 Add/confirm a service function (in `frontend/src/services`) that calls `GET /api/v1/downloads` with a `problem` value that may be a comma-separated string, reusing the existing downloads-fetching function if it already accepts an arbitrary `problem` string unchanged.
- [x] 2.2 Verify no changes are needed to `DownloadEnriched`/`ContentInfo`/`FileInfo` TS types in `frontend/src/types.ts` (the new tab reuses the existing shape) — confirm by reading the current type definitions against what the new components need.

## 3. Frontend: Tab Navigation and Desktop-Only Gating

- [x] 3.1 Add `'errors'` to `VALID_TABS` in `frontend/src/App.tsx` and register the "Erreurs" tab in the desktop segmented tab list only.
- [x] 3.2 Ensure the mobile bottom tab bar's tab list (used below the `768px` breakpoint) does NOT include the new tab, and add a fallback so that if `errors` is the active tab and the viewport narrows below the breakpoint, the active tab switches to another available tab. Verify manually by starting on desktop with Erreurs active, then resizing the browser below 768px.

## 4. Frontend: Errors Table

- [x] 4.1 Create a new `ErrorsTab` component (or equivalent) that fetches `GET /api/v1/downloads?problem=missing_year,year_mismatch,unknown_format` by default and renders one row per result with: type icon, title+year (falling back to file/folder name), reason badge(s), detected vs. expected year, detected extension, folder location, and completion date.
- [x] 4.2 Compute each row's reason badges from `file_info.has_year_in_path`, `file_info.year_mismatch`, and `file_info.is_valid_format` (a row can show more than one badge). Verify with a component/unit test asserting an item with two problems renders two badges.
- [x] 4.3 Add pagination controls and an items-per-page selector wired to `limit`/`offset`/`total`/`total_pages`, consistent with the existing Downloads tab pattern. Verify by testing page navigation re-fetches with the expected `offset`.

## 5. Frontend: Reason Filter

- [x] 5.1 Add a reason filter control with options "Tous", "Extension invalide", "Année manquante", "Année incohérente", mapped to `problem` values `missing_year,year_mismatch,unknown_format` (Tous), `unknown_format`, `missing_year`, `year_mismatch` respectively. Changing the filter SHALL reset pagination to page 1. Verify with a test asserting the correct `problem` query value is sent per option and that page resets to 1.

## 6. Frontend: Dedicated Diagnostic Sidepanel

- [x] 6.1 Create a new sidepanel component dedicated to the Erreurs tab (not reusing `MediaOccurrenceDrawer` or the Downloads tab's details sidepanel), opened on row click.
- [x] 6.2 Order the sidepanel's sections as: diagnostic (which reason(s) apply and their values, e.g. detected vs. expected year, detected extension) first, then file location, then matched content detail (title, year, genres, season/episode), then status/dates, then source `url`. Verify by asserting the diagnostic section renders before the others in the component's test.
- [x] 6.3 Confirm the sidepanel renders no action controls ("Déplacer", "Renommer", "Associer", "Forcer le téléchargement") for any download status. Verify with a test asserting none of these controls are present regardless of the selected download's status.

## 7. Verification

- [x] 7.1 Run `npm test` (vitest) in `frontend/` and confirm all tests pass, including the new Erreurs tab tests.
- [x] 7.2 Manually run the app, open the Erreurs tab on a desktop-width window, confirm the combined default view, the reason filter, pagination, and the dedicated sidepanel all work against real data, and confirm the tab disappears below `768px`.
- [x] 7.3 Run `openspec validate downloads-errors-tab --strict` and resolve any reported issues.
