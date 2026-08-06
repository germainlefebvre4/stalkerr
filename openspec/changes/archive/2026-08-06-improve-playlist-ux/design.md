## Context

The playlist page state (`playlist`, `usePlaylist.ts`) currently lives only in React state: `playlistPage`, `playlistLimit` (mirrored to `localStorage['stalkeer_playlist_limit']`), and four filter fields. Nothing is read from or written to the URL. The app has no routing library (`react-router` is not a dependency), and `activeTab` itself is only persisted via `localStorage`, not the URL. The pipeline state badge color is chosen by an inline ternary duplicated at two call sites in `PlaylistTab.tsx` (table row and sidepanel), and it has no branch for the `processed` state, so it falls through to the `pending` (amber) style.

## Goals / Non-Goals

**Goals:**
- Make the playlist page, page size, and filters survive a browser refresh via the URL query string.
- Keep the fix minimal and dependency-free (no router library) since only one page needs URL sync today.
- Centralize the state→badge-class mapping so the `processed` bug fix lives in one place, not two.
- Centralize date formatting so table and sidepanel can't drift apart again.

**Non-Goals:**
- Making the browser Back/Forward buttons step through pagination/filter history. Refresh-safety is the goal, not a navigation stack.
- Migrating `activeTab` or any other tab (filters, logs, downloads) to the URL. Out of scope for this change.
- Introducing `react-router` or any URL/state management library.

## Decisions

- **URL sync via native `URLSearchParams` + `history.replaceState`, not `pushState`.** Every page/filter change replaces the current history entry instead of pushing a new one. Rationale: pushing an entry per keystroke/page-click would flood the browser history stack and make Back effectively unusable; `replaceState` gives refresh-safety without that side effect. Alternative considered: `pushState` per page change only (not per filter change) — rejected for inconsistency (some state changes would be undoable, others not) and added complexity for no clear user benefit.
- **URL is the source of truth on load; `localStorage['stalkeer_playlist_limit']` remains the fallback default only.** On mount, `usePlaylist` reads page/limit/filters from `window.location.search` if present; otherwise it falls back to today's behavior (limit from `localStorage`, page 1, empty filters). Every change still writes the chosen limit to `localStorage` (so a fresh tab with no query string still remembers the user's preferred page size) and also writes the full state to the URL. Alternative considered: dropping `localStorage` entirely in favor of the URL — rejected because a brand-new tab/link with no query string should still respect the user's last-chosen page size instead of resetting to the default 50.
- **Sync logic lives inside `usePlaylist.ts`**, not a new generic "URL state" abstraction. Only one page needs this today; a reusable hook would be speculative generality for a single call site.
- **Extract a pure `getPipelineStateBadgeClass(state: string): string` helper** (co-located in `PlaylistTab.tsx` or a small shared module) used by both the table row and the sidepanel badge, replacing the two duplicated inline ternaries. This is where the `processed → success` fix is made, once.
- **Extract a pure `formatDate(value: string): string` helper** implementing manual zero-padding (`String(day).padStart(2, '0')` etc.), rather than relying solely on `toLocaleDateString('fr-FR', ...)`. Rationale: locale-based formatting depends on the ICU data available in the runtime/browser and isn't guaranteed to zero-pad consistently across environments, whereas manual formatting is deterministic. Used by both the table's "Créé le" column and the sidepanel's "Date d'import"/"Forcé le" fields.
- **Go-to-page input keeps its own local (uncontrolled-ish) text state**, clamped to `[1, totalPages]` and committed on Enter or a "Go" button click, rather than driving `playlistPage` directly on every keystroke (which would trigger a fetch per keystroke while the user is still typing).

## Risks / Trade-offs

- [Longer, less clean URLs once several filters are active] → Acceptable; this is a utility dashboard, not a marketing page, and refresh-safety matters more than URL aesthetics.
- [Writing to the URL on every keystroke in the free-text search fields] → `history.replaceState` does not reload or navigate the page, so the cost is negligible; no debouncing of the URL write is needed (the existing debounce/reset-to-page-1 behavior on filter change is unaffected).
- [Two capabilities' spec files touched for one change (`frontend-ihm-dashboard`, `m3u-playlist-details-sidepanel`)] → Intentional: the date-format fix genuinely spans both the table and the sidepanel, which are documented as separate capabilities.
