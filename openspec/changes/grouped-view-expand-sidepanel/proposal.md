## Why

In the "Films & Séries" (grouped) view, expanding a group shows its underlying items using the same `PlaylistItemsTable` component as the "Items" view, but clicking a row inside that expanded list does nothing. In the "Items" view, clicking a row opens the details sidepanel (`MediaOccurrenceDrawer`). Users expect the same click-to-open behavior in both places, since the expanded list is presented as the same table with the same columns and actions.

## What Changes

- Clicking an item row inside a group's expanded item list (grouped view) opens the same details sidepanel currently opened from the "Items" view's table, for that item.
- No change to the group row's own click behavior (still toggles expand/collapse).
- No new component: `PlaylistItemsTable` already supports an `onRowClick` prop and `MediaOccurrenceDrawer` already accepts a `PlaylistItem`; this wires the existing prop through `PlaylistGroupedView`.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `playlist-grouped-media-view`: expanding a group's item list gains row-click behavior that opens the same details sidepanel as the Items view.

## Impact

- Frontend: `frontend/src/components/PlaylistGroupedView.tsx` (accept and forward an `onRowClick` prop to the expanded `PlaylistItemsTable`), `frontend/src/components/PlaylistTab.tsx` (pass `setSelectedItem` through to `PlaylistGroupedView`, reusing the drawer already mounted there).
- No backend changes, no data shape changes (expanded items are already `PlaylistItem[]`, the same shape the drawer already consumes).
