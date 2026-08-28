## Context

See `proposal.md` - Why. `processed_lines` and `processing_logs` have no relationship today; `ProcessingLog.ItemCount` is a bare count with nothing behind it. Schema is managed via GORM `AutoMigrate` (`internal/database/database.go`, `runMigrations`), so additive nullable columns don't need hand-written migrations. `Processor.Process` (`internal/processor/processor.go`) creates the `ProcessingLog` row before parsing starts and knows its id for the rest of the run; `saveBatch` is the single place that creates/updates every `processed_lines` row.

## Goals / Non-Goals

**Goals:**
- Attribute every processed item to the run that last touched it, precisely and without ambiguity (see `specs/processing-run-items/spec.md` - Item Attribution to Processing Run).
- Let the frontend fetch a run's items via the existing items endpoint and existing table component, minimizing new UI surface.

**Non-Goals:**
- Backfilling `processing_log_id` for `processed_lines` rows written before this change (explicitly out of scope per proposal.md).
- Tracking full run history per item (only the *last* run that touched an item is recorded, not every run that ever saw it).
- Changing `GET /api/v1/items/grouped` or any grouped/aggregate view.

## Decisions

**FK column vs. time-window correlation.** Chose an explicit `ProcessingLogID *uint` column on `processed_lines`, set inside `saveBatch`, over correlating by `processed_at` falling between a run's `started_at`/`completed_at`. The FK is exact regardless of overlapping or unusually long-running imports, keeps the query a plain indexed equality filter, and gives the correct semantics for re-processing: an item touched by a later forced run should show up under that later run, not its original one. The cost is that historical logs have no items to show, which the proposal already accepts.

**Reuse `GET /api/v1/items` rather than a new endpoint.** `processing_log_id` is added as one more optional filter alongside the existing `movie_id`/`tmdb_id` filters in `listItems` (`internal/api/handlers.go`), following the same pattern already used for "Item Listing Supports Filtering by Movie or Show Identity". This keeps pagination, sorting, and response shape (`ItemResponse`) identical to every other item listing, and the run's item dialog gets those for free.

**Reuse the existing `PlaylistItemsTable` + dialog pattern rather than a bespoke read-only list.** `frontend/src/components/PlaylistItemsTable.tsx` already renders items with state badges and per-item actions (association/correction via `ManualOverrideDialog`, pipeline reset), and already supports `onRowClick`/pagination from its use in the grouped view's row expansion. A new `RunItemsDialog` (Radix `Dialog`, same shape as `ManualOverrideDialog.tsx`) wraps it and fetches `/api/v1/items?processing_log_id=<id>`, wired at the `App.tsx` level next to the already-mounted `ManualOverrideDialog` so the existing `onOpenOverride`/`onResetPipeline` handlers can be passed straight through without duplicating them.

**No new `ProcessingLog` state or `item_count` change.** `item_count` keeps meaning "count of items processed by this run" (`Statistics.Processed`), which now equals `COUNT(processed_lines WHERE processing_log_id = <run>)` by construction — no separate consistency mechanism is needed.

## Risks / Trade-offs

- [Historical logs show no items] → Explicitly accepted; the dialog renders the same empty state as a run with zero items, per the spec's pre-migration scenario. Users can re-run an import to attribute items going forward.
- [Resetting an item's pipeline from inside the run dialog doesn't refresh that dialog's own list] → The existing `handleResetPipeline` in `App.tsx` only refetches the Playlist tab's list (`fetchPlaylist`). `RunItemsDialog` must trigger its own refetch after a successful reset/override so the open dialog stays accurate; this is a small addition to the dialog's own success/callback handling, not a shared-state change.
- [`processing_log_id` filter on an unindexed column could scan the whole table as data grows] → Add a database index on the new column (mirrors the existing indexed FKs like `movie_id`, `download_info_id` on the same table).

## Migration Plan

Additive only: new nullable, indexed column picked up by `AutoMigrate` on next boot; no data backfill, no rollback beyond dropping the column (not required since it's additive and unused by old code paths).
