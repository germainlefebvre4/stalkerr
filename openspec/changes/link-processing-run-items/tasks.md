## 1. Data model

- [ ] 1.1 Add `ProcessingLogID *uint` (nullable, indexed) to `ProcessedLine` in `internal/models/processed_line.go`, and verify `go build ./...` succeeds and the app boots locally with the column auto-created (`\d processed_lines` / equivalent shows `processing_log_id`).

## 2. Backend: attribute items to their run

- [ ] 2.1 Thread the current `ProcessingLog.ID` from `Processor.Process` into `saveBatch` (`internal/processor/processor.go`), setting `line.ProcessingLogID` on every created or updated line, and verify with a processor test (extend `internal/processor/processor_test.go`) asserting a freshly processed line's `ProcessingLogID` matches the run's log id.
- [ ] 2.2 Verify a forced re-process of an existing line updates its `ProcessingLogID` to the newer run's id (test covering spec scenario "A re-processed item is re-attributed to the newer run").
- [ ] 2.3 Verify a skipped duplicate (non-forced, existing hash) leaves `ProcessingLogID` unchanged (test covering spec scenario "A skipped duplicate keeps its original attribution").

## 3. Backend: expose the filter

- [ ] 3.1 Add `processing_log_id` query param handling to `listItems` in `internal/api/handlers.go`, alongside the existing `movie_id`/`tmdb_id` handling, and add `ProcessingLogID` to `ItemResponse` in `internal/api/dto.go` (or reuse an existing field if more appropriate) so the frontend can display run attribution if needed.
- [ ] 3.2 Add a test in `internal/api/handlers_frontend_test.go` (alongside `TestListItemsFiltering_MovieIDAndTMDBID`) asserting `GET /api/v1/items?processing_log_id=<id>` returns only items with that attribution, and run `go test ./internal/api/...` to confirm it passes.
- [ ] 3.3 Add a test asserting `GET /api/v1/items?processing_log_id=<unknown-or-unattributed-id>` returns `200 OK` with an empty list (spec scenario "An unknown or not-yet-attributed processing_log_id returns an empty list").

## 4. Frontend: types and API client

- [ ] 4.1 Add `processing_log_id?: number` to the `PlaylistItem` type in `frontend/src/types.ts`.
- [ ] 4.2 Add a `getRunItems(processingLogId, page, limit)` method to `frontend/src/services/api.ts` following the shape of the existing `getGroupItems`, calling `/api/v1/items?processing_log_id=<id>&...`, and verify with `npx tsc --noEmit` (or the project's existing frontend build/typecheck command) that it type-checks.

## 5. Frontend: run items dialog

- [ ] 5.1 Create `RunItemsDialog` (e.g. `frontend/src/components/RunItemsDialog.tsx`) as a Radix `Dialog` following the shape of `ManualOverrideDialog.tsx`, fetching and paginating a run's items via `api.getRunItems` and rendering them with `PlaylistItemsTable` (same columns/badges/actions), including its own empty state for a run with zero attributed items.
- [ ] 5.2 Wire the dialog's item actions (association/correction, pipeline reset) to refetch the dialog's own item list on success, in addition to the existing `App.tsx` handlers' side effects (design.md - Risks: dialog staleness after in-place reset/override).
- [ ] 5.3 Make processing-log rows in `LogsTab.tsx` clickable (row click opens `RunItemsDialog` for that row's log id), matching the `clickable-row` styling already used by `PlaylistItemsTable`.
- [ ] 5.4 Mount `RunItemsDialog` and its open/selected-log state in `App.tsx` next to the existing `ManualOverrideDialog`, passing through the existing `onOpenOverride`/`onResetPipeline` handlers so the association dialog keeps working from inside the run dialog.

## 6. i18n

- [ ] 6.1 Add translation keys for the run items dialog (title, empty state, pagination) to `frontend/src/locales/en/logs.json` and `frontend/src/locales/fr/logs.json`, reusing existing `playlist`/`dialogs` namespace strings via `useTranslation` where the same copy already exists instead of duplicating it.

## 7. Verification

- [ ] 7.1 Run the full backend test suite (`go test ./...`) and confirm it passes.
- [ ] 7.2 Start the dev server, run an M3U import, click the resulting row in the Logs/Processing table, and confirm the dialog lists exactly that run's items (matching `item_count`), that an in-progress run's row shows items saved so far, and that a pre-existing (pre-migration) log row shows the empty state — per design.md - Migration Plan.
- [ ] 7.3 From within the run items dialog, exercise the association/correction and pipeline-reset actions and confirm the dialog's own list reflects the change without requiring a manual close/reopen.
