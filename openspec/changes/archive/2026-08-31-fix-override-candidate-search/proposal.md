## Why

On the Processing page, opening the manual override modal for a TV show item and selecting a TMDB result is supposed to offer a "bulk associate" candidate list of sibling episodes. Today that list is built by filtering the `playlist` array held by the main Playlist tab (`usePlaylist()` in `App.tsx`), regardless of where the modal was opened from. When opened from the Processing page's run-items dialog, that array has nothing to do with the run being inspected: it is whichever page (default page 1) the Playlist tab last loaded, in memory, for an unrelated filter/sort context. The candidate list is therefore effectively random from the user's point of view — often empty (the show's other episodes aren't on that page) or containing unrelated shows that happen to share the loaded page.

## What Changes

- Replace the candidate source in `ManualOverrideDialog`: instead of filtering the `playlist` prop passed down from `App.tsx`, the modal fetches its own candidate set from the backend, scoped to the opened item's cleaned title, via the existing `GET /api/v1/items` endpoint (`content_type=tvshows&tvg_name=<cleaned title>`, using the endpoint's existing `tvg_name` LIKE filter).
- Remove the `playlist` prop from `ManualOverrideDialog` (and the now-unnecessary pass-through in `App.tsx`), since the modal no longer depends on whatever page another tab happens to have loaded.
- Candidate matching/pre-check logic (clean-title equality, exclude the opened item, per-candidate season/episode preview) is unchanged — only where the candidate pool comes from changes.
- Fetch is scoped and paginated with a generous limit (not tied to any tab's page size), issued once when the modal selects a TMDB result for a `tvshow`, and re-issued if the opened item changes.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `tmdb-manual-override`: the bulk-associate candidate list requirement changes from "every other currently loaded playlist entry (the page of results currently displayed in the playlist table)" to "entries returned by a dedicated backend search scoped to the opened item's cleaned title", independent of any other view's pagination or filters.

## Impact

- Frontend: `frontend/src/components/ManualOverrideDialog.tsx` (candidate fetching), `frontend/src/services/api.ts` (reuse/extend the existing `getPlaylist`-style call for a candidate search), `frontend/src/App.tsx` (drop the `playlist` prop wiring for this dialog).
- No backend changes: `GET /api/v1/items` already supports `content_type` and `tvg_name` filtering (`internal/api/handlers.go`).
- No database or API contract changes.
