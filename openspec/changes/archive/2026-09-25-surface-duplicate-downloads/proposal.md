# Proposal

## Why

`fix-duplicate-media-downloads` (archived 2026-09-23) stopped the automatic pipeline from re-downloading a movie/episode that already has a different completed occurrence, but it deliberately did not touch the ~25 movies / 43 episodes already sitting in the database as duplicate `completed` rows, nor does it give an operator any way to see or resolve a duplicate if the bug recurs (e.g. through a code path this fix didn't cover). Today nothing in the product surfaces "this movie/episode was downloaded more than once by the automated pipeline" — an operator can only discover it by inspecting the database directly, and once discovered has no supported way to make Stalkeer's records match reality after cleaning the extra file up manually.

## What Changes

- Detect groups of automated downloads that duplicate each other: two or more `completed` occurrences of the same movie/episode whose stored `download_path` values differ (a same-path pair is a legitimate tier-2 quality upgrade overwrite, not a duplicate). Detection excludes any occurrence created by the manual `force-download-media-occurrence` action, since that is a deliberate, expected second file.
- Run this detection both continuously (as part of the existing scheduled path-reconciliation pass in the `download` command) and on demand (from `GET /api/v1/downloads`, so the historical backlog is visible immediately without waiting for a scheduled run).
- Add a `duplicate_download` value to the Erreurs tab's problem filter and a corresponding badge, alongside the existing `missing_year`/`year_mismatch`/`unknown_format` reasons.
- Add a manual, per-group resolve action: for each occurrence in a flagged duplicate group, verify whether its file still exists on disk and, for any occurrence whose file is gone, transition it to the existing `cancelled` terminal status. This is database-only (never deletes a file) and only ever runs when an operator explicitly triggers it — detection alone never mutates data.
- Give the Erreurs tab's diagnostic sidepanel an exception to its current "no corrective action" rule, scoped specifically to the new resolve action on `duplicate_download` items; the other three reasons keep offering no action, unchanged.
- Mark every `download_info` row created by a forced download with an explicit, persisted indicator of its manual origin, so duplicate detection (now and in the future) can reliably exclude it — closing the gap that today's code has no way to distinguish a forced download from an automatic one.

**Out of scope**: automatically resolving a duplicate without an explicit operator action; deleting files from disk (this system never does that); backfilling/cleaning duplicates any way other than through this new resolve action.

## Capabilities

### New Capabilities
- `duplicate-download-detection`: defines what counts as an unexpected automated duplicate (grouping rule, same-path exclusion for tier-2 upgrades, exclusion of forced downloads), where it runs (scheduled pass and on-demand read), and the manual resolve action that reconciles database state with what's actually on disk.

### Modified Capabilities
- `downloads-enrichment-api`: adds `duplicate_download` to the `problem` query parameter's supported values, sourced from the new detection capability rather than a `file_info`-derived condition like the existing three values.
- `downloads-errors-view`: adds the `duplicate_download` badge and reason-filter option; carves out a scoped exception to the sidepanel's "no corrective action" requirement for the new resolve action.
- `force-download-media-occurrence`: the endpoint now marks the `download_info` record it creates or reuses with an explicit forced-trigger indicator.

## Impact

- `internal/models/download.go`: `DownloadInfo` gains a field recording whether it originated from a forced download (schema change via GORM `AutoMigrate`, no manual migration file needed in this project).
- `internal/api/force_download.go`: `persistForceDownloadPath` sets the new field.
- New detection query (grouping `processed_lines` by `movie_id`/`tvshow_id`, joined to `download_info` where `status = completed`, excluding forced rows, comparing `download_path` across the group) — likely a new file under `internal/scheduler` or `internal/api`, reused by both the scheduled pass and the enrichment endpoint.
- `cmd/download.go` / the scheduled reconciliation pass (`sonarr-series-path-routing`'s existing job): calls the detection query each run and logs findings (report-only, no mutation), alongside its existing per-item `file_missing` check.
- `internal/api/handlers_frontend.go`: `listDownloadsEnriched`/`matchesSingleProblem` gain the `duplicate_download` condition; `DownloadEnrichedResponse` gains a field describing the sibling occurrence(s) for a flagged row.
- New endpoint (naming decided in design.md) for the manual resolve action, following the `resync-path`/`cancel` handler pattern.
- Frontend: `useErrorsTab.ts`, `errorReasons.ts`, `ErrorsTable.tsx`, the Erreurs sidepanel component, and `locales/{en,fr}` gain the new reason, badge, diagnostic content, and resolve action.
- No breaking API change (new field/value are additive); no removal of existing behavior.
