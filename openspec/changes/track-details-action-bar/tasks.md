## 1. Drawer: props and action bar layout

- [ ] 1.1 Add `onOpenOverride?: (item: PlaylistItem) => void` to `MediaOccurrenceDrawerBodyProps` and `MediaOccurrenceDrawerProps` in `frontend/src/components/MediaOccurrenceDrawer.tsx`, forwarding it from `MediaOccurrenceDrawer` down to `MediaOccurrenceDrawerBody`; verify the project type-checks (`npm run build` or `tsc --noEmit` in `frontend/`).
- [ ] 1.2 In `MediaOccurrenceDrawerBody`, add a new action bar `<div>` (flex row, gap) directly below the panel header, rendering an "Associate" button (`btn-primary`, calling `onOpenOverride?.(item)`, hidden/absent when `onOpenOverride` is not provided) alongside the existing "Force Download" button; verify by inspecting the rendered drawer in the browser with both buttons visible side by side.
- [ ] 1.3 Move the Force Download eligibility hint, queued badge, and error badge out of the old standalone "Force Download" section into a block directly beneath the new action bar; delete the now-empty old section; verify the ineligibility hint, queued state, and error state each still render in their new location by exercising each state manually.
- [ ] 1.4 Add i18n keys for the "Associate" button label/title under the `playlist:drawer.*` namespace in `frontend/src/locales/en/playlist.json` and `frontend/src/locales/fr/playlist.json` (mirroring the existing `table.actions.associate` wording), and reference them from the new button; verify both locales render the new label by switching language in the UI.

## 2. Playlist tab wiring

- [ ] 2.1 Pass `onOpenOverride={onOpenOverride}` into the `MediaOccurrenceDrawer` call in `frontend/src/components/PlaylistTab.tsx:329`; verify clicking "Associate" in the drawer opened from a Playlist row opens `ManualOverrideDialog` targeting that same item.

## 3. Radarr/Sonarr tab wiring

- [ ] 3.1 Add an `onOpenOverride?: (item: PlaylistItem) => void` prop to `RadarrSonarrTab`'s props interface in `frontend/src/components/RadarrSonarrTab.tsx`, and pass `handleOpenOverride` into it from `frontend/src/App.tsx`; verify the prop reaches the component (type-check plus a manual smoke check).
- [ ] 3.2 Forward `onOpenOverride` to both drawer usages in `RadarrSonarrTab.tsx`: the mobile `MediaOccurrenceDrawerBody` (~line 485) and the desktop `MediaOccurrenceDrawer` (~line 647); verify "Associate" opens `ManualOverrideDialog` for an occurrence selected from both the Films and Séries sidepanels, on both a desktop-width and a mobile-width viewport.

## 4. End-to-end verification

- [ ] 4.1 Manually verify the full flow: open Track Details from the Playlist tab for an unmatched item, click "Associate", complete an override, and confirm the drawer's TMDB metadata section reflects the new match without a page reload.
- [ ] 4.2 Manually verify "Associate" stays enabled while "Force Download" is disabled, for an occurrence that is already downloaded and for one that is currently downloading.
- [ ] 4.3 Manually verify the action bar and moved status badges render correctly (no overlap, no horizontal scrolling) at the drawer's mobile width.
