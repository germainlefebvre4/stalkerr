## Why

On the Playlist page, the "Items / Films & Séries" sub-tab selector sits on its own row above the "All / Movies / TV Shows" content-type filter row, and both are rendered as full pill buttons. This costs an extra row of vertical space and, on a narrow viewport, forces two separate rows of wrapping buttons. Merging the two controls onto a single row, and replacing the two-button sub-tab selector with a compact icon-only toggle switch pinned to the right, tightens the layout without losing functionality.

## What Changes

- Replace the "Items" / "Films & Séries" pill-button pair with a single compact toggle switch (custom-styled checkbox, no library dependency), flanked by two unicode icons (no text label) and a native `title` tooltip describing the current/target view, consistent with the project's existing icon-as-emoji and `title`-attribute-tooltip conventions.
- Move this switch onto the same row as the "All / Movies / TV Shows" content-type filter buttons: the three filter buttons stay left-aligned in their existing order, and the switch is pushed to the far right of that row (`margin-left: auto`), so it stays right-aligned even if the row wraps.
- On mobile viewports (below the existing `768px` breakpoint), reduce the content-type filter buttons' padding and inter-button gap so the merged row (3 buttons + switch) fits on one line at common mobile widths instead of wrapping, using the same `useIsMobile()` conditional pattern already used elsewhere in `PlaylistTab.tsx`.
- No change to the underlying `playlistView`/`playlistFilter` state, data fetching, or filter semantics — this is a presentation-only change to the existing controls.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `playlist-grouped-media-view`: The "Grouped Media Sub-Tab" requirement's UI is changing from a two-button sub-tab pair on its own row to an icon-only toggle switch sharing a row with the content-type filter buttons, right-aligned. The switching behavior and effect (replacing the table, preserving active filters) are unchanged.
- `frontend-responsive-layout`: Adds a new requirement for compacting the merged Playlist selector row (content-type filter buttons + view toggle switch) below the mobile breakpoint, so it fits on one line without wrapping.

## Impact

- `frontend/src/components/PlaylistTab.tsx`: merge the two selector rows into one; replace the two sub-tab buttons with a new toggle switch element; add mobile-conditional padding/gap.
- `frontend/src/index.css`: new CSS for the toggle switch (track/knob, checked state) and any mobile-specific compact-row rules.
- `frontend/src/locales/{en,fr}/playlist.json`: replace or repurpose the `view.items` / `view.grouped` button labels with tooltip text (and any new keys needed for the toggle's `title`).
- No backend/API changes.
