## Why

The M3U playlist items table lists every ingested item with its import date (`created_at`), but every row looks identical regardless of when it was imported. Users scanning the table for recent activity (today's or yesterday's import run) have no quick visual anchor and must read every date cell individually.

## What Changes

- Add lightweight date-group separators to the playlist items table (desktop table rows and mobile cards): a thin, low-emphasis header is inserted above the first item of each new calendar day encountered in the current page.
- Group headers use a relative label for the two most recent days (`Aujourd'hui` / `Hier`, localized) and the full date (same format as the existing date cell) for any earlier day.
- Grouping and separators are only shown when the active sort is `created_at` (either direction); for any other sort column, the table renders as a flat list with no separators, since rows are no longer contiguous by day.
- Each paginated page re-renders its leading group header, even when that page opens mid-group (i.e., continues a group started on the previous page), so a page is self-contained without requiring the previous page for context.
- Purely a frontend rendering concern: grouping is computed client-side from the items already returned for the current page; no API or backend changes.

## Capabilities

### New Capabilities
- `playlist-import-date-grouping`: Defines how the playlist items table (desktop and mobile) groups items by import day and renders relative/full-date separators, including its dependency on the active sort field and its behavior across pagination.

### Modified Capabilities
(none)

## Impact

- `frontend/src/components/PlaylistTab.tsx`: desktop table row rendering and mobile card list rendering.
- `frontend/src/utils/date.ts` (or a new sibling util): date-grouping/labeling helper.
- `frontend/src/locales/fr/playlist.json`, `frontend/src/locales/en/playlist.json`: new labels for "Today" / "Yesterday" group headers.
- `frontend/src/index.css`: minimal styling for the separator/group header.
- No backend, API, or database changes.
