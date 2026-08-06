## 1. Shared URL state hook

- [ ] 1.1 Create `frontend/src/hooks/useURLState.ts`: accepts a schema of `{ [key]: { default, parse, serialize, isValid? } }`, returns `[state, patchState]`. `patchState` re-reads `window.location.search` fresh, merges only the schema's own keys, omits keys equal to their `default`, and writes back via `history.replaceState`. Reading falls back to `default` when a value is missing, unparsable, or fails `isValid`.
- [ ] 1.2 Add an initial-read helper (mirroring `usePlaylist`'s current `readInitialStateFromURL`) so a schema can be read once on mount via `useRef`.

## 2. Migrate Playlist onto the shared hook

- [ ] 2.1 Rewrite `frontend/src/hooks/usePlaylist.ts` to define its 6 keys (`page`, `limit`, `search`, `searchName`, `tmdb`, `type`, `state`) as a `useURLState` schema, using the same param names, defaults, and validity allow-lists (`VALID_LIMITS`, `VALID_TMDB_FILTERS`, `VALID_CONTENT_FILTERS`, `VALID_STATE_FILTERS`) it already has.
- [ ] 2.2 Port the "reset page to 1 only when a filter actually changed (not on initial mount)" guard so it still holds with the new hook.
- [ ] 2.3 Keep the `limit` mirror into `localStorage` under `stalkeer_playlist_limit` (unchanged key/behavior).
- [ ] 2.4 Manually verify: reload with `?page=3&type=movies&state=failed` restores that exact filtered page (per the existing "Restore pagination and filters from the URL after a refresh" scenario); changing a filter still resets to page 1; changing the page size still writes `stalkeer_playlist_limit`.

## 3. Active tab in the URL

- [ ] 3.1 In `frontend/src/App.tsx`, add a 1-key `useURLState` schema for `tab` (default `'playlist'`, valid values `playlist|filters|logs|downloads`).
- [ ] 3.2 On mount, prefer the `tab` URL parameter when present; fall back to the existing `localStorage.getItem('stalkeer_active_tab')` value when it is absent.
- [ ] 3.3 On tab change, keep writing `localStorage` (unchanged) and additionally patch the `tab` URL parameter.
- [ ] 3.4 Manually verify: switching tabs and refreshing restores the same tab; opening a URL with `?tab=downloads` activates Downloads even if `localStorage` has a different stored tab.

## 4. Downloads filters in the URL

- [ ] 4.1 In `frontend/src/hooks/useDownloads.ts`, add a 3-key `useURLState` schema for `dlStatus`, `dlType`, `dlProblem` (defaults `''`, matching today's "all" behavior), replacing the bare `useState` calls for `statusFilter`/`typeFilter`/`problemFilter`.
- [ ] 4.2 Manually verify: selecting a status/type/problem filter on Downloads updates the URL under `dlStatus`/`dlType`/`dlProblem`; refreshing restores the same filters; setting a Playlist filter (`type`, `state`) and a Downloads filter at the same time leaves both present in the URL simultaneously (no clobbering).

## 5. Final checks

- [ ] 5.1 Run `npm run lint` and `npm run build` in `frontend/` and fix any resulting errors.
- [ ] 5.2 Manually exercise all four tabs in the browser per tasks 2.4, 3.4, and 4.2 in a single session (switch tabs, set filters on Playlist and Downloads, refresh, confirm nothing resets and nothing cross-applies between tabs).
