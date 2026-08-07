## Why

The frontend has zero `@media` queries anywhere in the codebase — the entire layout (fixed 1600px container, inline-flex tab pills, tables wrapped in `overflowX: auto`) is desktop-only. On a phone, the 4-tab navigation and the 6-column Playlist/Logs tables force horizontal scrolling just to reach a tab or read a row, which breaks the core navigation and browsing flows on mobile.

## What Changes

- Introduce a mobile breakpoint and, below it, replace the segmented tab pills with a fixed bottom tab bar (Playlist / Filters / Logs / Downloads) so switching sections never requires horizontal scrolling. Desktop keeps the current segmented pills.
- On mobile, the Playlist and Logs tables render as stacked cards (one card per row, key info + status badge) instead of a horizontally-scrollable `<table>`; tapping a card opens the existing details sidepanel, same as clicking a row today.
- On desktop, the Playlist and Logs tables render full-width by removing the extra padding layers (page container + `.card` wrapper) around the table specifically, to show more content per row without introducing horizontal scroll.
- Reduce the Playlist default page size from 50 to 10 entries per page (applies on both desktop and mobile); the existing 10/50/100 selector and `localStorage` persistence are unchanged, only the un-set default changes. Logs and Downloads are not paginated today and remain out of scope for this change.
- Tighten padding, font sizes, and touch target sizes on KPI cards, filter cards, and download cards for mobile viewports (no structural change — they already stack via CSS grid).
- Keep the Playlist details sidepanel as a right-side drawer on all viewports, but resize its padding/typography for mobile and enlarge the close button's touch target.

## Capabilities

### New Capabilities
- `frontend-responsive-layout`: mobile breakpoint definition, bottom tab bar navigation on mobile, table-to-card transformation for Playlist/Logs on mobile, full-width table rendering on desktop, and mobile ergonomics for cards and the sidepanel.

### Modified Capabilities
- `frontend-ihm-dashboard`: the Playlist page-size default changes from 50 to 10 entries per page.

## Impact

- Frontend-only change: no API/backend contract changes.
- Affected files: `frontend/src/index.css`, `frontend/src/variables.css`, `frontend/src/App.tsx`, `frontend/src/components/PlaylistTab.tsx`, `frontend/src/components/LogsTab.tsx`, `frontend/src/components/DownloadsTab.tsx`, `frontend/src/components/FiltersTab.tsx`, `frontend/src/components/FloatingHeader.tsx`, `frontend/src/components/StatsKPICards.tsx`, `frontend/src/hooks/usePlaylist.ts` (`DEFAULT_LIMIT`).
