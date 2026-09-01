## Why

The Radarr/Sonarr monitoring tab's active sub-tab (Résumé/Radarr/Sonarr) resets to Radarr on every page refresh, unlike the Playlist tab's Items/Grouped sub-view, which survives a refresh via a URL query parameter. Users switching between the Radarr and Sonarr sub-tabs lose their place every time they reload. Additionally, the Radarr and Sonarr sub-tabs are currently text-only, making them harder to visually scan than if they carried each solution's own logo.

## What Changes

- The Radarr/Sonarr monitoring tab's active sub-tab (`resume`/`radarr`/`sonarr`) is persisted across a page refresh via a URL query parameter, following the same `useURLState`-based pattern already used for the Playlist tab's Items/Grouped sub-view (`usePlaylistView`), and kept as a state independent from the tab's own data-fetching hook.
- The Radarr and Sonarr sub-tab triggers each display their respective solution's icon (from the provided `radarr.svg` / `sonarr.svg` files) alongside their existing text label. The Résumé sub-tab is unchanged (text label only).
- Establishes a `frontend/src/assets/icons/` convention for bundled SVG icon assets, imported as ES modules (no such asset directory exists in the frontend today).

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-view`: the "Radarr/Sonarr tab organizes content into three sub-tabs" requirement is extended so the selected sub-tab persists across a page refresh, and the Radarr/Sonarr sub-tab triggers display their solution's icon.

## Impact

- `frontend/src/components/RadarrSonarrTab.tsx`: replaces the local `useState` for `activeSubTab` with a URL-backed hook, and adds an icon image to the Radarr and Sonarr `Tabs.Trigger` elements.
- New hook (e.g. `frontend/src/hooks/useRadarrSonarrView.ts`), modeled on `frontend/src/hooks/usePlaylistView.ts`, built on the existing generic `useURLState` hook.
- New files `frontend/src/assets/icons/radarr.svg` and `frontend/src/assets/icons/sonarr.svg`, copied from the user-provided sources.
- No backend or API changes.
