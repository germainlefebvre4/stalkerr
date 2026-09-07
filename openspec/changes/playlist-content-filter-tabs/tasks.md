## 1. Component Markup

- [ ] 1.1 In `frontend/src/components/PlaylistTab.tsx` (~lines 99-122), replace the three `<button>` elements for the content-type filter with `Tabs.Root value={playlistFilter} onValueChange={...} className="segmented-tabs-list playlist-content-tabs"` wrapping three `Tabs.Trigger` elements (`all`/`movies`/`tvshows`) with `className="segmented-tabs-trigger"`, keeping the existing `t('contentFilter.all'|'movies'|'tvshows')` labels. `Tabs` is already imported in this file (`import * as Tabs from '@radix-ui/react-tabs'`). Each `onValueChange`/`onClick` path SHALL still call `setPlaylistFilter(value)` followed by `setPlaylistPage(1)`, matching current behavior. Verify by reading the diff: no `<button>` remains for this filter, and `playlistFilter`/`setPlaylistFilter`/`setPlaylistPage` props are still wired the same way.
- [ ] 1.2 Verify no `Tabs.Content` panel is introduced for `all`/`movies`/`tvshows` — the existing table/grouped view below continues to render unconditionally from the single fetched `playlist` list, unchanged. Verify by confirming `PlaylistItemsTable`/`PlaylistGroupedView` render calls are untouched.

## 2. Styling

- [ ] 2.1 In `frontend/src/index.css`, add a `.playlist-content-tabs` modifier analogous to `.radarr-sonarr-subtabs` (~line 908) inside the `@media (max-width: 767.98px)` block (~line 903), so `.segmented-tabs-list.playlist-content-tabs` uses `display: flex; overflow-x: auto; -webkit-overflow-scrolling: touch;` (overriding the base `.segmented-tabs-list { display: none; }` rule) and `.playlist-content-tabs .segmented-tabs-trigger { flex-shrink: 0; }`, so the content-type filter stays visible and horizontally scrollable on mobile like today. Verify by inspecting the added CSS block for syntax correctness.
- [ ] 2.2 Remove the now-unused inline `style={{ padding: ... }}` sizing from the old buttons if superseded by `segmented-tabs-trigger` padding, or keep an inline override only if needed to preserve current mobile sizing. Verify by comparing rendered padding on mobile vs desktop against the current buttons' `0.35rem 0.6rem` / `0.45rem 1rem` values.

## 3. Manual Verification

- [ ] 3.1 Run the frontend dev server, open the Playlist tab, and confirm clicking each of "Tous"/"Films"/"Séries" (or their i18n labels) updates the `?type=` URL query param, resets to page 1, and filters the table exactly as before.
- [ ] 3.2 Confirm the active tab is visually indicated via the `segmented-tabs-trigger[data-state="active"]` style (same look as the Radarr/Sonarr sub-tabs), and that keyboard arrow-key navigation moves focus between the three tabs when one is focused.
- [ ] 3.3 Resize to a mobile viewport and confirm the content-type filter remains visible above the (still collapsed-by-default) advanced filters disclosure, matching the existing `frontend-ihm-dashboard` "Mobile Collapsible Advanced Filters" requirement.
- [ ] 3.4 Run the frontend lint/typecheck/build (e.g. `npm run build` or the project's configured script in `frontend/`) and confirm it passes with no new errors.
