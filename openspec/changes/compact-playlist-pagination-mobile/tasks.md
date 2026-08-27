## 1. Compact mobile pagination bar

- [ ] 1.1 In `frontend/src/components/PlaylistTab.tsx`, add an `isMobile`-conditional render branch for the pagination buttons block (lines ~427-508): when `isMobile` is true, render only the first/previous/next/last controls (`<<`, `<`, `>`, `>>`) plus a `Page X / Y` text indicator (using `t('pagination.page', { current: playlistPage, total: totalPages })` or equivalent i18n key), and omit the numbered page-button list (`getPaginationRange` map) and the "go to page" input/button group. Verify by reading the rendered branch: on `isMobile`, no `getPaginationRange` buttons or "go to page" block are present in the JSX output.
- [ ] 1.2 Add the new translation key(s) used by the mobile page indicator (e.g. `pagination.page`) to every locale file under `frontend/src/i18n/` (or wherever existing `pagination.*` keys live), matching the interpolation placeholders used in 1.1. Verify by grepping all locale files for the new key and confirming each has a value.
- [ ] 1.3 Confirm the desktop branch (`!isMobile`) is untouched: numbered page buttons with ellipsis, first/last jump buttons, and the "go to page" input still render exactly as before. Verify by diffing the desktop-branch JSX against the pre-change code.

## 2. Verification

- [ ] 2.1 Run `npm run build` and `npm run lint` in `frontend/` and verify both succeed with no errors.
- [ ] 2.2 Start the dev server (`npm run dev`), open the Playlist tab with a viewport narrower than `768px` and a result set spanning multiple pages, and verify visually that the pagination bar shows `<<  <  Page X / Y  >  >>` with no numbered buttons, no "go to page" input, and no horizontal scrollbar on the page.
- [ ] 2.3 In the same narrow viewport, tap `>` and `<` and verify the current page changes, the fetched items update, and the `Page X / Y` indicator reflects the new page.
- [ ] 2.4 Resize/switch to a viewport at or above `768px` and verify the pagination bar reverts to the full desktop layout (numbered buttons, ellipsis, first/last jump buttons, "go to page" input) with no visual regression.
