## 1. Wire row-click through the grouped view

- [x] 1.1 Add an `onRowClick?: (item: PlaylistItem) => void` prop to `PlaylistGroupedViewProps` (`frontend/src/components/PlaylistGroupedView.tsx`) and pass it to the `PlaylistItemsTable` rendered in `renderExpandedSection()`, and verify `frontend` type-checks (`npm run build` or `tsc --noEmit`).
- [x] 1.2 In `PlaylistTab.tsx`, pass `onRowClick={setSelectedItem}` to `<PlaylistGroupedView />`, reusing the existing `selectedItem`/`MediaOccurrenceDrawer` already mounted for the Items view.

## 2. Tests

- [x] 2.1 Add `frontend/src/components/PlaylistGroupedView.test.tsx` covering: clicking an item row inside an expanded group calls `onRowClick` with that item and does not collapse the group (matches spec scenario "Clicking an underlying item opens its details sidepanel"); clicking the group's own row still toggles expand/collapse; clicking an item row's "Associate"/"Correct" or "Reset" button calls `onOpenOverride`/`onResetPipeline` without calling `onRowClick` (matches spec scenario "Row-level action buttons do not open the sidepanel").
- [x] 2.2 Run the frontend test suite (`npm test` in `frontend/`) and verify it passes, including the new test file.

## 3. Manual verification

- [x] 3.1 Run the app, open the Playlist page's "Films & Séries" (grouped) tab, expand a group, click one of its underlying item rows, and verify the details sidepanel opens showing that item's TMDB metadata, pipeline badges, and provenance — matching what opening the same item from the "Items" tab shows.
- [x] 3.2 Verify clicking the "Associate"/"Correct" or "Reset" button on an item row inside the expanded list still performs that action and does not open the sidepanel.
