## 1. Sub-tab persistence

- [x] 1.1 Create `frontend/src/hooks/useRadarrSonarrView.ts`, modeled on `usePlaylistView.ts`: its own `useURLState` schema with a `subtab` field (`'resume' | 'radarr' | 'sonarr'`, default `'resume'`, validated against those three values), exposing `activeSubTab` and `setActiveSubTab`. Verify it compiles and existing hooks/tests are unaffected.
- [x] 1.2 In `frontend/src/components/RadarrSonarrTab.tsx`, replace the local `React.useState<'resume' | 'radarr' | 'sonarr'>('radarr')` (line 65) with `useRadarrSonarrView()`, wiring `activeSubTab`/`setActiveSubTab` into the existing `Tabs.Root value`/`onValueChange` props. Verify `npm run build` (or `tsc --noEmit`) passes in `frontend/`.
- [x] 1.3 Verify manually (or via test) that selecting the Radarr or Sonarr sub-tab, then reloading the page, keeps that same sub-tab selected, matching the Items/Grouped behavior on the Playlist tab.

## 2. Solution icons

- [x] 2.1 Create `frontend/src/assets/icons/` and copy `radarr.svg` and `sonarr.svg` into it from the provided source files, unmodified. Verify the files exist at `frontend/src/assets/icons/radarr.svg` and `frontend/src/assets/icons/sonarr.svg`.
- [x] 2.2 In `frontend/src/components/RadarrSonarrTab.tsx`, import both SVGs as ES modules (`import radarrIcon from '../assets/icons/radarr.svg'`, `import sonarrIcon from '../assets/icons/sonarr.svg'`) and render each as an `<img>` before the label text in the Radarr and Sonarr `Tabs.Trigger` elements (lines 164-165); leave the Résumé trigger (line 163) unchanged. Verify `npm run build` passes.
- [x] 2.3 Add minimal CSS for the new icon (e.g. a `tab-icon` class sized to align with the existing text label) alongside the other `segmented-tabs-trigger` styles. Verify visually in the browser that both icons render at a consistent size next to their labels, in both the Radarr and Sonarr sub-tabs.

## 3. Test coverage

- [x] 3.1 Update `frontend/src/components/RadarrSonarrTab.test.tsx` to cover: sub-tab selection persists in the URL query string, and the Radarr/Sonarr triggers render their icon `<img>` while the Résumé trigger does not. Verify `npm test` passes in `frontend/`.
