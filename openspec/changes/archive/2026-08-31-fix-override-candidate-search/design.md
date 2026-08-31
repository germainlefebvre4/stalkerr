## Context

See proposal.md - Why. `ManualOverrideDialog` (`frontend/src/components/ManualOverrideDialog.tsx:106-130`) currently builds its bulk-associate candidate list by filtering the `playlist` prop, which is `App.tsx`'s `usePlaylist()` state - the array backing the main Playlist tab's own paginated table. That state exists regardless of which tab is active and is wired into `ManualOverrideDialog` unconditionally (`App.tsx:269-279`), so opening the modal from the Processing page's run-items dialog (`RunItemsDialog`) inherits whatever page/filter the Playlist tab last fetched, unrelated to the run or the item being corrected.

The backend's item-listing endpoint (`GET /api/v1/items`, `internal/api/handlers.go`) already supports `content_type` and `tvg_name` (case-insensitive `LIKE`) filters, used today by `getPlaylist`/`getGroupItems` in `frontend/src/services/api.ts`.

## Goals / Non-Goals

**Goals:**
- Candidate list is correct and complete for the opened item's show, independent of any other view's loaded page/filter/sort.
- No backend changes; reuse the existing `GET /api/v1/items` filters.
- Preserve existing candidate UX exactly (pre-check heuristic, per-row season/episode preview, manual check/uncheck) - only the source of the candidate pool changes.

**Non-Goals:**
- Changing the clean-title matching heuristic itself (`cleanRawTitle`, `extractSeasonEpisode`).
- Introducing a new, purpose-built backend endpoint for "siblings" - the existing filter is sufficient given real `tvg_name` data is space-delimited, not dot-delimited (verified against test fixtures), so a substring `LIKE` on the cleaned title reliably matches sibling episodes.
- Changing how the batch of override requests is issued or reported (steps 10-13 of the requirement are untouched).

## Decisions

**Fetch scoped by `tvg_name` search, not a full unpaginated fetch (Piste B over Piste A).** When `overrideMediaType === 'tvshow'` and a TMDB result is selected, `ManualOverrideDialog` calls `api.getPlaylist(1, <limit>, 'tvshows', undefined, undefined, cleanedTitle)` (or an equivalent typed helper) instead of filtering a `playlist` prop. `cleanedTitle` is the same `cleanRawTitle(overrideItemData.tvg_name)` string already computed for the search box. The server-side `LIKE %cleanedTitle%` on `tvg_name` narrows the result set before it reaches the client, avoiding a full-table client-side scan.
- Alternative considered: fetch all `content_type=tvshows` items unpaginated (Piste A) and keep filtering client-side. Rejected because it re-fetches the same unbounded pool that's part of today's problem (just a differently-scoped unbounded pool), and does unnecessary work on large playlists when a targeted search is just as effective here.
- Alternative considered: a new backend endpoint that performs the clean-title matching itself (Piste C). Rejected for this fix - it centralizes matching logic in Go (a real long-term win) but is a larger change than the bug warrants; can be revisited later if the clean-title heuristic needs to grow more backend-side smarts.

**`playlist` prop is removed from `ManualOverrideDialog`.** Nothing else in the component depends on it once the candidate fetch is self-contained. `App.tsx` drops the prop from its `<ManualOverrideDialog ... />` usage.

**Candidate fetch limit.** Use a fixed generous limit (e.g. 100) rather than any tab's configured page size, since the candidate list's own scroll area (`ManualOverrideDialog.tsx:400`, `maxHeight: 160px`) already handles overflow, and a show is very unlikely to have more unassociated episodes than that in one batch.

**Fetch trigger and staleness.** The candidate fetch runs once when `(overrideItemData, overrideMediaType === 'tvshow', selectedResult)` become true, mirroring the existing `useEffect` at `ManualOverrideDialog.tsx:106-130` - just replacing its data source, not its trigger conditions. It is not re-issued on every keystroke of the search box; changing media type or reselecting a TMDB result for the same item does not need a fresh candidate fetch since the opened item's cleaned title doesn't change.

## Risks / Trade-offs

- [Risk] The `LIKE %cleanedTitle%` search could still miss a sibling whose raw title, after cleaning, differs enough from the opened item's cleaned title (e.g. inconsistent tagging across episodes) → Mitigation: this is the same limitation the current exact-equality pre-check already has; the "manually add a candidate the heuristic did not suggest" scenario already covers the user working around a missed match. Not a regression.
- [Risk] A very short or generic cleaned title (e.g. a single common word) could match unrelated shows via substring `LIKE` → Mitigation: unchanged from today's risk surface - the candidate list already requires the user to review and can pre-check false positives, which the user can uncheck (existing "deselects a false-positive candidate" scenario).
- [Trade-off] One extra network round-trip when a TMDB result is selected, versus the current zero-cost (but incorrect) client-side filter. Acceptable given the correctness gain and that this only fires once per override attempt.

## Migration Plan

Frontend-only change, no data migration. Deploy as a normal frontend release; no feature flag needed since the new behavior is a strict correctness fix over the old one (empty/wrong candidate lists were already effectively unusable).
