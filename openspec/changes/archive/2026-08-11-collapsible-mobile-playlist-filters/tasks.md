## 1. Local state and derived count

- [x] 1.1 In `PlaylistTab.tsx`, add local state `const [advancedFiltersOpen, setAdvancedFiltersOpen] = React.useState(false)` alongside the existing `selectedItem`/`copiedText`/`gotoPageInput` state, so it always starts collapsed on mount and is never persisted.
- [x] 1.2 Compute an `activeAdvancedFilterCount` at render time from `playlistSearchName`, `playlistSearch`, `playlistTMDBFilter`, and `playlistStateFilter` (count how many are non-empty/non-`'all'`).

## 2. Mobile disclosure markup

- [x] 2.1 In the `isMobile` branch, keep the content-type pills block (`Tous`/`Films`/`Séries`) rendered exactly as today, outside and above the disclosure.
- [x] 2.2 Wrap the advanced filter grid (media name, group, TMDB enrichment, pipeline state) in a disclosure: a clickable header showing a label and `activeAdvancedFilterCount` (when > 0) plus a chevron indicator, toggling `advancedFiltersOpen` on click.
- [x] 2.3 Render the advanced filter grid's fields only when `advancedFiltersOpen` is true; when false, render nothing for that grid so the table/card list is visible without scrolling past it.
- [x] 2.4 Leave the non-mobile branch of the advanced filter grid unchanged (always inline, no disclosure, no toggle).

## 3. Styling

- [x] 3.1 Add CSS for the disclosure header (spacing, chevron rotation on open/close, active-filter count badge) in `frontend/src/index.css`, scoped so it only affects the new disclosure markup (no changes to existing desktop filter styles).
- [x] 3.2 Verify touch target size on the disclosure header meets the existing mobile `min-height: 44px` rule already established for buttons/badges in `mobile-responsive-ui`.

## 4. i18n

- [x] 4.1 Add new keys under a `advancedFilters` section in `frontend/src/locales/fr/playlist.json` and `frontend/src/locales/en/playlist.json` for the disclosure header label and the active-filter-count text (e.g. `advancedFilters.toggle`, `advancedFilters.activeCount`).
- [x] 4.2 Use `useTranslation('playlist')`'s existing `t()` call in `PlaylistTab.tsx` to render the new disclosure header text, following the same pattern as the existing `fields.*`/`contentFilter.*` keys.

## 5. Manual verification

- [x] 5.1 On a mobile viewport (or browser dev tools mobile emulation, `<768px`), confirm the Playlist tab loads with the advanced filters collapsed and the card list visible without scrolling past the filter grid.
- [x] 5.2 Confirm tapping the disclosure header expands the advanced filter grid, and tapping again collapses it, with no change to the content-type pills.
- [x] 5.3 Set a pipeline state filter and a TMDB enrichment filter, collapse the disclosure, and confirm the active-filter count shown is `2`.
- [x] 5.4 Load the Playlist tab on mobile with a filter already present in the URL query string (e.g. `?state=failed`) and confirm the disclosure still starts collapsed while the `failed` filter is applied to the fetched results and reflected in the count.
- [x] 5.5 On a desktop viewport (`>=768px`), confirm the advanced filter grid still renders inline and always visible with no disclosure control, matching current behavior.
