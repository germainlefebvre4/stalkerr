## Why

On the Playlist (M3U) view, the pagination bar's page-number buttons, first/last jump buttons, and "go to page" input are all rendered in a single non-wrapping row identical to desktop. On mobile viewports (below the `768px` breakpoint) this row can exceed 10+ elements and overflows the viewport width, forcing a horizontal scrollbar on the whole page — which contradicts the no-horizontal-scroll guarantee the mobile layout otherwise provides for the Playlist view.

## What Changes

- Below the mobile breakpoint, replace the numbered page-button list in the Playlist pagination bar with a compact `Page X / Y` text indicator, keeping the first/previous/next/last (`<<`, `<`, `>`, `>>`) controls.
- Below the mobile breakpoint, hide the "go to page" input/button group from the pagination bar.
- Desktop behavior (`>= 768px`) is unchanged: numbered page buttons with ellipsis, first/last jump buttons, and the "go to page" input all continue to render exactly as today.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `frontend-responsive-layout`: adds a requirement that the Playlist pagination bar renders in a compact form (page indicator instead of numbered buttons, "go to page" hidden) below the mobile breakpoint, with no horizontal overflow.

## Impact

- `frontend/src/components/PlaylistTab.tsx` — the inline pagination bar (page-number rendering via `getPaginationRange`, first/last/prev/next buttons, "go to page" block) gains a mobile-specific render branch using the existing `useIsMobile()` hook.
- No API, backend, or data model changes.
