## 1. Shared breakpoint foundation

- [ ] 1.1 Add `frontend/src/hooks/useMediaQuery.ts` exporting `useIsMobile()`, backed by `window.matchMedia('(max-width: 767.98px)')` with a `change` listener (cleaned up on unmount).
- [ ] 1.2 Add the matching `767.98px` breakpoint as the single `@media (max-width: 767.98px)` cut line used by every mobile-only CSS rule added in this change, so JS and CSS agree.

## 2. Mobile bottom tab navigation

- [ ] 2.1 In `App.tsx`, add a bottom tab bar (plain buttons for Playlist/Filters/Logs/Downloads, same icons/labels as the segmented pills) that calls the existing `setActiveTab`, rendered only when `useIsMobile()` is true.
- [ ] 2.2 Hide the existing `.segmented-tabs-list` below the breakpoint via CSS, and add `position: fixed; bottom: 0` styling (plus `env(safe-area-inset-bottom)` padding) for the new bar in `index.css`.
- [ ] 2.3 Add bottom padding to the page container below the breakpoint equal to the bar's height, so scrolled content never sits underneath it.
- [ ] 2.4 Verify tab selection still persists to `localStorage` under the same key and restores on refresh when using the bottom bar.

## 3. Playlist table → mobile cards, full-width desktop

- [ ] 3.1 In `PlaylistTab.tsx`, branch on `useIsMobile()`: keep the existing `<table>` for desktop, add a stacked card list (media name + pipeline state badge) for mobile over the same `playlist` array.
- [ ] 3.2 Wire the mobile card's tap handler to the same `setSelectedItem` used by the desktop row click, so the details sidepanel opens with identical content.
- [ ] 3.3 Add a `.table-flush` wrapper class (negative margins matching the card/tab-content padding) applied to the table wrapper at `≥768px`, so the table spans the full card width; keep the filter controls block at normal padding.

## 4. Logs table → mobile cards, full-width desktop

- [ ] 4.1 In `LogsTab.tsx`, branch on `useIsMobile()`: keep the existing `<table>` for desktop, add a stacked card list (action + status badge) for mobile over the same `logs` array.
- [ ] 4.2 Apply the same `.table-flush` wrapper class to the Logs table at `≥768px`.

## 5. Playlist default page size

- [ ] 5.1 In `frontend/src/hooks/usePlaylist.ts`, change `DEFAULT_LIMIT` from `50` to `10`.
- [ ] 5.2 Confirm the change only affects users with no `stalkeer_playlist_limit` in `localStorage` and no `limit` in the URL query string (existing stored/URL values still win).

## 6. Card density and touch targets

- [ ] 6.1 Add a `@media (max-width: 767.98px)` block in `index.css` reducing padding/font-size on `.kpi-card`, `.filter-card`, and `.download-card`.
- [ ] 6.2 In the same media query block, set `min-height: 44px` (and `min-width: 44px` for icon-only controls) on the shared button classes (`.btn-primary`, `.btn-secondary`, `.btn-danger`) and on the drawer close control.

## 7. Sidepanel mobile ergonomics

- [ ] 7.1 Add mobile-scoped overrides for `.drawer-content` padding and internal typography sizing in `index.css`.
- [ ] 7.2 Confirm the close control (`Dialog.Close` in `PlaylistTab.tsx`) meets the 44x44px touch target added in 6.2.

## 8. Manual verification

- [ ] 8.1 At `≤767px`: confirm all four tabs are reachable via the bottom bar with no horizontal scroll, Playlist/Logs show cards with no horizontal scroll, tapping a card opens the sidepanel, and Downloads/Filters cards remain fully readable.
- [ ] 8.2 At `≥768px`: confirm the segmented pills are unchanged, Playlist/Logs tables render full-width with no layout regression, and the Playlist page size defaults to 10 for a fresh (no `localStorage`) session.
- [ ] 8.3 Confirm switching the browser viewport across the `768px` boundary while a session is active (resize, not reload) transitions cleanly between layouts without console errors.
