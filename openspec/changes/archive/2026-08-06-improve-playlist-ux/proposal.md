## Why

The M3U playlist page has four small UX issues reported by the user: dates render in the browser's locale format instead of a consistent `DD/MM/YYYY`, the current page is lost on refresh because pagination state lives only in memory, there is no way to jump directly to an arbitrary page, and the `processed` pipeline state (a successful terminal state set by the backend processor for all content types) is visually rendered with the same amber/warning color as the `pending` (waiting) state, misleadingly suggesting a problem where there is none.

## What Changes

- Format the "Créé le" column and the details sidepanel's "Date d'import" / "Forcé le" timestamps using a fixed `DD/MM/YYYY` (zero-padded) format instead of locale-dependent `toLocaleDateString()` / `toLocaleString()`.
- Persist playlist pagination state (page, page size, and all active filters: media name search, group search, content type, TMDB enrichment, pipeline state) in the URL query string, so a browser refresh restores the exact same view.
- Add a "go to page" input next to the existing pagination controls, letting the user type a page number and jump directly to it.
- Fix the pipeline state badge color mapping so `processed` renders with the success (green) style instead of falling through to the pending (amber) style; `pending` keeps its amber styling since it is a genuine waiting state.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `frontend-ihm-dashboard`: playlist pagination is now reflected in the URL (deep-linkable/refresh-safe), pagination gains a direct page-jump control, and the pipeline state badge color mapping distinguishes `processed` (success) from `pending` (waiting).
- `m3u-playlist-details-sidepanel`: dates displayed in the sidepanel ("Date d'import", "Forcé le") use the same fixed `DD/MM/YYYY` format as the playlist table.

## Impact

- Frontend only: `frontend/src/components/PlaylistTab.tsx`, `frontend/src/hooks/usePlaylist.ts`.
- Likely a new small date-formatting utility shared between the table and the sidepanel.
- No backend or API changes: the `processed` state already exists and is returned as-is; only its front-end color mapping changes.
- No new dependencies expected (no router library needed; URL sync can use the native `URLSearchParams`/History API).
