## Why

On mobile, opening the Playlist details sidepanel (drawer) triggers an unwanted horizontal scrollbar inside the panel. The cause is the "Hash unique" row: `line_hash` is rendered inside a plain `<code>` element with no `word-break`/`overflow-wrap`, so the unbroken hash string forces the drawer wider than the viewport. This field has no practical use on screen (it exists for internal deduplication), so the fix is to remove its display entirely rather than just wrap it.

## What Changes

- Remove the "Hash unique" row (label, `<code>{line_hash}</code>`, and its copy button) from the Playlist details sidepanel, on both mobile and desktop.
- Remove the now-unused `drawer.uniqueHash` translation key (fr/en) and the associated `handleCopy(..., 'hash')` call site in `PlaylistTab.tsx`.
- `line_hash` remains untouched in the API response and `PlaylistItem` type — only its on-screen display in the sidepanel is removed.
- **BREAKING**: none (UI-only removal; no API/data contract change).

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `m3u-playlist-details-sidepanel`: The "Playlist Item Details Sidepanel" requirement's blanket "toutes les métadonnées brutes de provenance et d'ingestion" scope is narrowed to explicitly exclude the unique hash — it is no longer part of the panel's displayed content.
- `frontend-responsive-layout`: The "Sidepanel Mobile Ergonomics" requirement gains an explicit guarantee that the sidepanel's own content never requires horizontal scrolling on mobile, closing the gap that let the hash field overflow unnoticed.

## Impact

- `frontend/src/components/PlaylistTab.tsx`: remove the "Hash unique" JSX block and its `handleCopy` call.
- `frontend/locales/fr/playlist.json`, `frontend/locales/en/playlist.json`: remove the orphaned `drawer.uniqueHash` key.
- No backend/API changes.
