## Why

On the grouped playlist view ("Films & Séries"), a TV-show group's displayed season range (e.g. `S01-S14`) is aggregated over every episode ever processed for that show, while expanding the same group only reveals items from its most recent processing run (e.g. `S13-S14`). The badge and the expanded content visibly disagree, making the range look wrong or stale.

## What Changes

- `GET /api/v1/items/grouped` computes a TV-show group's `season_start`/`season_end` from only the episodes belonging to the group's most recent processing run (the same run reported by its run-attribution field and used to scope the expanded item list), instead of the show's entire history.
- When a group predates run attribution (no run-attribution value), the season range continues to be computed from all of the group's underlying items, matching the expanded list's existing fallback behavior in that case.
- The existing per-filter scoping of the season range (only matching items contribute) is preserved; this change narrows that scope further to the latest run.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `playlist-grouped-media-view`: the "Grouped Row Content" requirement's season-range aggregation is scoped to the group's most-recent-processing-run items (falling back to all items when no run is attributed), so it always matches what "Expanding a Group Lists Its Underlying Items" reveals.

## Impact

- `internal/api/handlers_grouped.go`: `listItemGroups`'s `tvshowsArm` query, currently `MIN(t.season)`/`MAX(t.season)` over all `processed_lines` rows for a `tmdb_id`, needs to restrict that aggregation to the rows sharing the group's latest `processing_log_id` (falling back to the unrestricted aggregation when no run is attributed).
- No frontend changes: `PlaylistGroupedView.tsx`'s `seasonLabel` already renders whatever `season_start`/`season_end` the API returns.
- Existing backend tests asserting season ranges from multi-run fixtures will need updated expectations.
