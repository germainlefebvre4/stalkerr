## Why

Files with naming problems (invalid extension, missing or inconsistent year) are only visible today by opening the Downloads tab, picking a single `problem` filter value at a time, and reading the detail sidepanel of each match one by one. There is no single place to see everything that needs a manual fix across the whole catalog, which makes cleanup slow and easy to miss items for.

## What Changes

- Add a new desktop-only "Erreurs" tab dedicated to surfacing downloaded items whose organized file has a naming problem: invalid/unrecognized extension, missing year, or year inconsistent with the matched TMDB year.
- The tab is hidden below the existing `768px` mobile breakpoint (no mobile card fallback, not added to the mobile bottom tab bar) — it stays a desktop-only surface.
- New dedicated table: one row per problematic download, with one or more reason badges per row (an item can have more than one problem at once), plus title/type, detected vs. expected year, detected extension, file location, and completion date.
- A reason filter narrows the table to a single problem type ("Extension invalide", "Année manquante", "Année incohérente") or shows all.
- Clicking a row opens a new, dedicated sidepanel built specifically for this tab, leading with the diagnostic detail (which rule failed and why) before the file/content context. This is a separate component from the existing Downloads tab sidepanel.
- No corrective actions (rename/move/associate/force-download) are exposed from this tab — it is a read-only triage view; corrections still happen from the existing Downloads/Playlist tabs.
- **Backend**: extend the `problem` query parameter on `GET /api/v1/downloads` so a request can match items having *any* of the known naming problems in one call (needed for the tab's default "all errors" view), in addition to the existing single-value behavior.

## Capabilities

### New Capabilities
- `downloads-errors-view`: the new desktop-only "Erreurs" tab — its table, reason badges, reason filter, and dedicated sidepanel.

### Modified Capabilities
- `downloads-enrichment-api`: the `problem` query parameter on `GET /api/v1/downloads` gains support for matching any of several problem types in a single request (not just one exact value), to back the new tab's combined "all errors" and per-reason views.

## Impact

- Backend: `internal/api` download listing/enrichment handler and its `problem` filter logic (`internal/api/handlers_frontend.go` and related); `openspec/specs/downloads-enrichment-api/spec.md`.
- Frontend: tab navigation (`VALID_TABS` in `frontend/src/App.tsx`) and mobile breakpoint gating (`useMediaQuery`); a new tab component, table, and sidepanel under `frontend/src/components`; `frontend/src/types.ts` if new fields are needed for reason reporting.
