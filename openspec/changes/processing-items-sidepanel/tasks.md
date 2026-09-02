## 1. Wire item selection in RunItemsDialog

- [x] 1.1 Add local `selectedItem` state (`useState<PlaylistItem | null>(null)`) to `frontend/src/components/RunItemsDialog.tsx`, mirroring `PlaylistTab.tsx`, and verify the file compiles (`tsc`/build) with no unused-state warnings
- [x] 1.2 Pass `onRowClick={setSelectedItem}` to the `PlaylistItemsTable` rendered inside `RunItemsDialog` and verify clicking a row (manual check or component test) updates `selectedItem` to that row's item

## 2. Render the sidepanel nested in the run's items dialog

- [x] 2.1 Render `<MediaOccurrenceDrawer item={selectedItem} onOpenChange={(open) => !open && setSelectedItem(null)} withOverlay={false} modal={false} />` inside `RunItemsDialog` and verify it opens showing the clicked item's details (TMDB metadata, pipeline-state badges, raw ingestion details) matching what the Playlist page shows for the same item
- [x] 2.2 Manually verify there is no double dark overlay and no competing focus trap when the sidepanel is open on top of the run's items dialog (only one overlay visible, keyboard focus stays usable)
- [x] 2.3 Manually verify closing the sidepanel (its own close button, Escape, and outside click) each clear `selectedItem` and leave `RunItemsDialog` open and showing the run's item list

## 3. Regression check on existing row actions

- [x] 3.1 Manually verify the existing per-row actions (association/correction, pipeline reset) inside the run's items dialog still work without opening the sidepanel, confirming their `e.stopPropagation()` still prevents `onRowClick` from firing
