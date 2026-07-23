## Why

Currently, the M3U playlist table pagination is hardcoded to 15 items per page and only offers "Précédent" and "Suivant" navigation buttons. This is completely inadequate for large-scale playlists with up to 100,000 items, as users cannot jump to specific pages and are forced to step through thousands of pages sequentially. Furthermore, querying and sorting 100,000 rows on a non-indexed column (`created_at`) would cause severe database performance degradation.

## What Changes

- **Database Performance**: Add an index to the `created_at` column in the `processed_lines` database table to make sorting high offsets fast and efficient under SQLite/PostgreSQL.
- **Configurable Page Size**: Provide a page-size selector in the UI (10, 50, 100 entries per page).
- **User Preference Persistence**: Store the selected page size in `localStorage` to persist across reloads.
- **Pleasant Pagination Navigation**: Replace the simple Next/Previous buttons with an interactive pagination bar featuring:
  - Direct page number buttons.
  - Smart ellipsis (`...`) representation for large page counts.
  - First-page (`<<`) and Last-page (`>>`) navigation shortcuts.
  - Current record range indicator (e.g., "Affichage de 51 à 100 sur 102 345 entrées").

## Capabilities

### New Capabilities
*None*

### Modified Capabilities
- `frontend-ihm-dashboard`: Extend M3U playlist UI requirements to cover page-size customization, interactive page number navigation, range indicators, and user preference persistence in `localStorage`.

## Impact

- **Backend**:
  - `internal/models/processed_line.go`: Add an index tag for the `CreatedAt` field.
- **Frontend**:
  - `frontend/src/hooks/usePlaylist.ts`: Expose `playlistLimit` and `setPlaylistLimit` states, and persist selection in `localStorage`.
  - `frontend/src/components/PlaylistTab.tsx`: Render a dynamic pagination bar with limit selector and smart page number buttons.
