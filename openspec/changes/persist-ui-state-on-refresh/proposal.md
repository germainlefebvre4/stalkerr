## Why

Refreshing the dashboard loses the Downloads tab's filters (status/type/problem reset to "all") and never lets a URL carry which tab is active, while the Playlist tab already survives a refresh via its own ad-hoc URL-writing logic. Extending that same ad-hoc pattern to Downloads and to the active tab would make three independent `useEffect`s each rewrite the full query string from scratch, and each write would clobber the params the others just wrote (last effect to run wins). This change introduces one shared hook that all URL-backed state goes through, and migrates the existing Playlist logic onto it so there is a single, race-free mechanism instead of one working ad-hoc implementation plus two new ones.

## What Changes

- Add a shared `useURLState` hook that reads/writes URL query parameters by always re-reading the live `window.location.search` before merging a change, so concurrent state owners never overwrite each other's keys.
- Migrate `usePlaylist`'s existing URL read/write logic (page, limit, search, searchName, tmdb, type, state) onto the shared hook. No observable behavior change for the Playlist tab.
- Add the active tab (`playlist` | `filters` | `logs` | `downloads`) to the URL query string (`?tab=`) via the shared hook, in addition to the existing `localStorage` persistence, so a refresh AND a shared link both restore the correct tab.
- Add the Downloads tab's filters (status, type, problem) to the URL query string via the shared hook, using dedicated param names (`dlStatus`, `dlType`, `dlProblem`) so they cannot collide with Playlist's `type`/`state` params, and restore them on load/refresh.

## Capabilities

### New Capabilities
(none — no new user-facing capability; this extends existing dashboard/downloads behavior)

### Modified Capabilities
- `frontend-ihm-dashboard`: the active tab is now also reflected in and restored from the URL query string, not just `localStorage`.
- `downloads-display-ui`: the status/type/problem filters are now reflected in and restored from the URL query string on load/refresh.

## Impact

- Affected code: `frontend/src/hooks/usePlaylist.ts`, `frontend/src/hooks/useDownloads.ts`, `frontend/src/App.tsx` (active tab state), new `frontend/src/hooks/useURLState.ts`.
- No backend/API changes.
- No breaking changes; existing bookmarked Playlist URLs keep working since their param names are unchanged.
