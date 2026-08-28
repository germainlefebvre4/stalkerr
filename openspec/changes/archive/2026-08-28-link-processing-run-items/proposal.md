## Why

The "Latest Task Executions" table (Logs/Processing page) shows an `item_count` per import run, but there is no way to see *which* items that count refers to — `processing_logs` and `processed_lines` are not linked in the database. Users have to guess what an import actually added or touched.

## What Changes

- Add a `processing_log_id` foreign key on `processed_lines`, set to the current run's `ProcessingLog.ID` whenever a line is created or updated during `Processor.saveBatch` (both newly created lines and re-touched existing lines get attributed to the run that touched them last).
- Extend `GET /api/v1/items` to accept an optional `processing_log_id` query parameter, restricting results to items linked to that run — following the same pattern as the existing `movie_id`/`tmdb_id` filters.
- Add a click interaction on rows of the processing-logs table: clicking a row opens a dialog listing that run's items, reusing the existing playlist items table (same columns, state badges, and per-item actions: association/correction, pipeline reset) and its existing pagination.
- Processing log entries created before this change have no `processing_log_id` data on their items (no backfill); their item dialog will show an empty state until re-processed.

## Capabilities

### New Capabilities
- `processing-run-items`: Links each processed playlist item to the import run that last touched it, and lets users inspect a run's items by clicking its row in the processing-logs table.

### Modified Capabilities
(none — no existing capability's requirements change)

## Impact

- Backend: `internal/models/processed_line.go` (new column), `internal/processor/processor.go` (`saveBatch` sets the FK), `internal/api/handlers.go` (`listItems` filter), `internal/api/dto.go` (`ItemResponse` field), `internal/database/database.go` (auto-migration already covers `ProcessedLine`, no new manual migration needed).
- Frontend: `frontend/src/types.ts` (`PlaylistItem`/`ProcessingLog` types), `frontend/src/services/api.ts` (new fetch for run-scoped items), `frontend/src/components/LogsTab.tsx` (clickable rows), new `RunItemsDialog` component wired in `App.tsx` alongside the existing `ManualOverrideDialog`, translation keys in `frontend/src/locales/{en,fr}/logs.json`.
