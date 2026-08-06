## 1. Shared utilities

- [x] 1.1 Add a `formatDate(value: string | Date): string` helper (e.g. `frontend/src/utils/date.ts`) that returns a zero-padded `DD/MM/YYYY` string via manual `padStart` formatting (not locale-dependent `toLocaleDateString`).
- [x] 1.2 Add a `getPipelineStateBadgeClass(state: string): string` helper mapping `processed`/`downloaded` → `badge-success`, `downloading`/`organizing` → `badge-progress`, `failed` → `badge-failed`, `pending` (and any other/unknown state) → `badge-pending`.

## 2. Apply date formatting

- [x] 2.1 Replace `new Date(item.created_at).toLocaleDateString()` in the playlist table's "Créé le" column (`PlaylistTab.tsx`) with `formatDate(item.created_at)`.
- [x] 2.2 Replace `new Date(selectedItem.created_at).toLocaleString()` ("Date d'import") and `new Date(selectedItem.override_at).toLocaleString()` ("Forcé le") in the details sidepanel with `formatDate(...)`.

## 3. Fix pipeline state badge color

- [x] 3.1 Replace the two duplicated inline ternaries (table row badge and sidepanel badge in `PlaylistTab.tsx`) with calls to `getPipelineStateBadgeClass(item.state)` / `getPipelineStateBadgeClass(selectedItem.state)`.

## 4. Persist pagination and filters in the URL

- [x] 4.1 In `usePlaylist.ts`, on initial mount, read `page`, `limit`, and the filter fields (media name search, group search, content type, TMDB enrichment, pipeline state) from `window.location.search` when present, falling back to today's defaults (limit from `localStorage`, page 1, empty filters) when absent.
- [x] 4.2 Add an effect that serializes the current page, limit, and filters into `URLSearchParams` and writes them via `history.replaceState` whenever any of them changes, omitting params that are at their default value to keep the URL minimal.
- [x] 4.3 Verify the existing "reset to page 1 on filter change" and `localStorage` page-size persistence behavior still work unchanged alongside the new URL sync.

## 5. Direct "go to page" control

- [x] 5.1 Add a small numeric input next to the existing `<< < > >>` pagination buttons in `PlaylistTab.tsx`, with local text state.
- [x] 5.2 On Enter or a "Go" button click, parse the input, clamp it to `[1, totalPages]`, and call `setPlaylistPage` with the clamped value; ignore/reset non-numeric input.

## 6. Verification

- [x] 6.1 Manually verify in the browser: refreshing the playlist page on page 3 with active filters restores the same page and filters; jumping via the new input navigates to the correct page; a `processed` item shows a green badge and a `pending` item still shows an amber badge; dates render as `DD/MM/YYYY` in both the table and the sidepanel.
