## Why

Expanding a group in the "Films & Séries" grouped view (`GET /api/v1/items?movie_id=X` or `&tmdb_id=X`) returns every item ever linked to that movie/show, across every processing run since it was first seen — not just the items from the run that produced the group's `latest_activity` timestamp shown in the table. Users expanding a group they see under "Aujourd'hui" expect to see only what was just processed, not the full history mixed in.

## What Changes

- `GET /api/v1/items/grouped` additionally reports, per group, the `processing_logs` id of the run that produced its `latest_activity` (i.e. the most recent run attribution among the group's contributing items).
- The "Films & Séries" view's expand action passes that id as the existing `processing_log_id` filter (capability `processing-run-items`) alongside its current `movie_id`/`tmdb_id`/pseudo-group filter, so the expanded item list is scoped to the same run that earned the group its place in the table.
- No change to how groups are listed, ordered, or dated — only to which underlying items an expanded group reveals.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `playlist-grouped-media-view`: group entries gain a run-attribution field, and expanding a group scopes its item list to that group's most recent processing run instead of the group's full history.

## Impact

- Backend: `internal/api/handlers_grouped.go` (`listItemGroups`, `itemGroupRow`), `internal/api/dto.go` (`ItemGroupResponse`) — add `MAX(pl.processing_log_id)` per UNION arm.
- Frontend: `frontend/src/types.ts` (`MediaGroupItem`), `frontend/src/services/api.ts` (`getGroupItems`) — thread the new field through as a `processing_log_id` query param.
- No database migration: reuses the existing `processed_lines.processing_log_id` column and the existing `GET /api/v1/items?processing_log_id=` filter from capability `processing-run-items`.
