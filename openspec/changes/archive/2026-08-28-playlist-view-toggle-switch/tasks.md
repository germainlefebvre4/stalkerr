## 1. Styles

- [x] 1.1 Add toggle switch CSS to `frontend/src/index.css` (hidden native checkbox, track, knob, checked-state transform/color) reusing existing CSS custom properties (`--accent-gradient`, `--border-color`, `--radius-sm`, etc.), and verify it renders as a track+knob switch with no visible checkbox in both light states.
- [x] 1.2 Verify the switch is keyboard-focusable and toggles on Space/Enter when focused (native checkbox behavior), by tabbing to it in the browser.

## 2. Translations

- [x] 2.1 In `frontend/src/locales/en/playlist.json` and `frontend/src/locales/fr/playlist.json`, add `view.switchToGrouped` and `view.switchToItems` tooltip strings, and remove the now-unused `view.items`/`view.grouped` button-label strings if nothing else references them (grep the codebase for `view.items`/`view.grouped` first to confirm), and verify `i18next` has no missing-key console warnings when the Playlist tab loads in either language.

## 3. Component

- [x] 3.1 In `frontend/src/components/PlaylistTab.tsx`, merge the "Items"/"Films & Séries" row and the "All"/"Movies"/"TV Shows" row into a single flex row: the three content-type buttons stay in their own inner wrapping flex group (unchanged internally) as the first flex child, and verify the buttons still work exactly as before (clicking still calls `setPlaylistFilter` and `setPlaylistPage(1)`).
- [x] 3.2 Replace the two "Items"/"Films & Séries" buttons with the new toggle switch element (checkbox bound to `playlistView === 'grouped'`, flanking `☰`/`🎬` icons marked `aria-hidden`, dynamic `title` from the new i18n keys), wired to `setPlaylistView`, and verify toggling it switches between the items table and the grouped view exactly as the old buttons did (existing filters stay applied).
- [x] 3.3 Give the switch `marginLeft: 'auto'` so it sits at the right edge of the row, and verify visually at a desktop width that the 3 filter buttons are left-aligned and the switch is flush right on the same line.
- [x] 3.4 Add `isMobile`-conditional padding/gap reduction to the content-type filter buttons and the row's gap (mirroring the existing `useIsMobile()` pattern already used a few lines below for advanced filters), and verify at a 360px viewport width that the row (3 buttons + switch) renders on one line without wrapping.

## 4. Verification

- [x] 4.1 Run the frontend locally, open the Playlist tab, and manually verify: toggling the switch between Items/Grouped preserves the active content-type/state/TMDB/search filters (per the existing `playlist-grouped-media-view` scenarios), at both a desktop width (e.g. 1280px) and a narrow mobile width (360-375px), with no horizontal scrolling introduced on the selector row.
- [x] 4.2 Run the frontend's existing lint/typecheck/test commands (per `package.json` scripts) and verify they pass with no new errors introduced by this change.
