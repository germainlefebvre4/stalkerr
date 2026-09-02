## Why

Opening a processing run's items dialog (Logs/Processing page) already reuses the exact same table and `PlaylistItem` data as the Playlist page, but clicking a row inside that dialog currently does nothing — the detail sidepanel available from the Playlist page is not wired up there, forcing users back to the Playlist page (and its own filtering) just to inspect one item's TMDB metadata, pipeline state, or raw ingestion details.

## What Changes

- Wire `onRowClick` on the items table rendered inside the run's items dialog (`RunItemsDialog`), the same way the Playlist page wires it, so clicking a row selects that item.
- Reuse the existing, page-agnostic `MediaOccurrenceDrawer` sidepanel to display the selected item's details, driven by local `selectedItem` state inside `RunItemsDialog` (mirroring `PlaylistTab`'s pattern) — no new drawer component or content.
- Since the run's items dialog is itself a Radix `Dialog.Root`, disable the sidepanel's own overlay and focus-trap (`withOverlay={false}`, `modal={false}`) so it nests inside the parent dialog without a double overlay or competing focus traps, rather than stacking two fully modal dialogs.
- No changes to the sidepanel's content, the item data shape, or any other page (Playlist, Radarr/Sonarr) that already consumes `MediaOccurrenceDrawer`.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `processing-run-items`: the "Inspecting a Processing Run's Items From the Logs Table" requirement gains a scenario where clicking an item row within the run's items dialog opens the same detail sidepanel used on the Playlist page, scoped to that single item.

## Impact

- `frontend/src/components/RunItemsDialog.tsx`: add `selectedItem` state, wire `onRowClick` on `PlaylistItemsTable`, mount `MediaOccurrenceDrawer` with `withOverlay={false}` and `modal={false}`.
- No backend/API changes — the run's items dialog already fetches full `PlaylistItem` records via `api.getRunItems`, the same shape `MediaOccurrenceDrawer` already expects.
- No changes to `MediaOccurrenceDrawer.tsx`, `PlaylistItemsTable.tsx`, or `PlaylistTab.tsx` — their existing props already support this usage.
