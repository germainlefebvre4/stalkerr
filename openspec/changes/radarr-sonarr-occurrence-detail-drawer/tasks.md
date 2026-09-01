## 1. Frontend API client

- [ ] 1.1 Add `api.getItem(id: number): Promise<PlaylistItem>` in `frontend/src/services/api.ts` wrapping the existing `GET /api/v1/items/:id` endpoint, following the existing method conventions (e.g. `forceDownload`); verify with `npx tsc --noEmit`.

## 2. Extract shared media detail drawer

- [ ] 2.1 Create `frontend/src/components/MediaOccurrenceDrawer.tsx`: move the existing item-detail `Dialog` JSX out of `PlaylistTab.tsx` verbatim (TMDB section, pipeline section, force-download section, M3U provenance section, raw line/URL sections), taking `item: PlaylistItem | null` and `onOpenChange: (open: boolean) => void` as props, and owning its own `copiedText`/`forceDownloadStatus`/`forceDownloadError` state (reset on `item?.id` change, same as today). Update `PlaylistTab.tsx` to render `<MediaOccurrenceDrawer item={selectedItem} onOpenChange={...} />`; verify with `npx tsc --noEmit`.
- [ ] 2.2 Manually verify the Playlist tab's drawer is behaviorally unchanged after extraction: open a matched item and an unmatched item, force-download an eligible item, copy the raw M3U line and the stream URL - confirm identical behavior to before extraction.

## 3. Films: occurrence rows open the detail drawer

- [ ] 3.1 In `RadarrSonarrTab.tsx`, make each row of a matched movie's "Playlist occurrences" table clickable; on click, call `api.getItem(occurrence.id)` and, once resolved, open `MediaOccurrenceDrawer` with the result.
- [ ] 3.2 Desktop (`!useIsMobile()`): render the detail drawer as a second `Dialog.Root` with `modal={false}` and no `Dialog.Overlay` of its own, positioned via a new `.drawer-content--secondary` CSS class; verify both the occurrences sidepanel and the detail drawer are visible and interactive simultaneously, and that closing the primary sidepanel (movie/series selection cleared) also closes the secondary drawer.
- [ ] 3.3 Mobile (`useIsMobile()`): add a `drawerView: 'occurrences' | 'detail'` state; selecting an occurrence swaps the existing `Dialog.Content`'s body to the detail view in place instead of mounting a second dialog, with a back action returning to `'occurrences'`; verify no second `Dialog.Root`/overlay is ever mounted on mobile.

## 4. Séries: expandable episode rows

- [ ] 4.1 Add `expandedEpisode: { season: number; episode: number } | null` state to the series sidepanel; clicking an episode row toggles its expansion instead of being inert.
- [ ] 4.2 When expanded, render the episode's `occurrences` (already returned by `GET .../sonarr/series/:id/episodes`) as a nested occurrence table structurally identical to the movie occurrence table; when an episode has zero occurrences, show an inline "no occurrences" indicator instead of an empty expandable list.
- [ ] 4.3 Make each expanded occurrence row clickable, opening `MediaOccurrenceDrawer` via the same desktop-stack / mobile-swap logic built in 3.2-3.3.
- [ ] 4.4 Reformat the season/episode label from `S{{season}}E{{episode}}` to zero-padded, space-separated `S{{season}} E{{episode}}` (e.g. "S01 E01"), padding season/episode with `String(n).padStart(2, '0')` before interpolation, for both the episode row label and the existing `drawer.episode` i18n key; verify visually that a single-digit season/episode renders as e.g. "S01 E03".

## 5. Styling

- [ ] 5.1 Add `.drawer-content--secondary` to `frontend/src/index.css`: same base styling as `.drawer-content` but `right: 550px` instead of `right: 0` (desktop only - no mobile override needed since it is never mounted on mobile); verify visually that it sits flush against the left edge of the primary drawer with no gap or overlap on a desktop-width viewport.

## 6. i18n

- [ ] 6.1 Add/adjust strings in `frontend/src/locales/fr/radarrSonarr.json` and `frontend/src/locales/en/radarrSonarr.json` for: the "no occurrences" indicator on an expanded episode, and the mobile back-to-occurrences action.

## 7. Verification

- [ ] 7.1 Run `npx tsc --noEmit`, `npm run lint`, and `npm test` in `frontend/` and confirm all green, excluding the 3 pre-existing unrelated `react-hooks/set-state-in-effect` lint errors already present on `main` before this change (`DownloadsTab.tsx`, `ManualOverrideDialog.tsx`, and - moved by the extraction in 2.1 - wherever the same pre-existing pattern now lives).
- [ ] 7.2 Manually launch the app and walk through: click a movie occurrence row -> detail drawer opens stacked to the left of the sidepanel, force-download works from it; click an unmatched episode -> shows "no occurrences", does not open an empty list; click a matched episode -> expands to its occurrences -> click one -> detail drawer opens; resize to mobile width and repeat both drill-downs, confirming the detail view replaces the sidepanel content in place instead of stacking.
