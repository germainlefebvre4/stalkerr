## 1. Backend Optimization

- [x] 1.1 Add database index tag `index:idx_processed_lines_created_at` to the `CreatedAt` field in `internal/models/processed_line.go` to optimize high-offset pagination queries.

## 2. Frontend Hook Refactoring

- [x] 2.1 Refactor `frontend/src/hooks/usePlaylist.ts` to manage `playlistLimit` state, initializing it from `localStorage` under `stalkeer_playlist_limit` with a default fallback of 50.
- [x] 2.2 Update `api.getPlaylist` invocation in `usePlaylist.ts` to utilize the dynamic `playlistLimit` instead of the hardcoded value of 15, ensuring it is properly added to the `useCallback` dependency array.

## 3. UI Component Implementation

- [x] 3.1 Expose `playlistLimit` and `setPlaylistLimit` in the `PlaylistTabProps` of `frontend/src/components/PlaylistTab.tsx` and wire them correctly in `frontend/src/App.tsx`.
- [x] 3.2 Implement a page size selector dropdown (10, 50, 100) in the `PlaylistTab.tsx` header/footer that updates `playlistLimit` and resets `playlistPage` to 1.
- [x] 3.3 Replace the static "Précédent" / "Suivant" buttons in `PlaylistTab.tsx` with a full-featured pagination bar rendering a dynamic page list (with ellipses `...`), first page `<<`, and last page `>>` buttons.
- [x] 3.4 Display a clean record range descriptor in French, e.g., "Affichage de X à Y sur Z entrées".

## 4. Verification

- [x] 4.1 Verify correct query responses on the network tab when changing limits and page indexes.
- [x] 4.2 Verify persistence of page-size preferences in `localStorage` across page reloads.
