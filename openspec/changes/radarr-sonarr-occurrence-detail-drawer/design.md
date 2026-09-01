## Context

See `proposal.md` - Why/What Changes for motivation and scope.

Relevant existing building blocks this change builds on:
- `frontend/src/components/PlaylistTab.tsx`: contains an inline (not extracted) Radix `Dialog` drawer rendering a `PlaylistItem`'s TMDB metadata, pipeline state, force-download action, M3U provenance, and raw line/URL. It owns local `copiedText`/`forceDownloadStatus`/`forceDownloadError` state reset via a `React.useEffect` keyed on `selectedItem?.id`.
- `frontend/src/components/RadarrSonarrTab.tsx`: today renders one Radix `Dialog` sidepanel per selection (movie or series), using `.drawer-overlay`/`.drawer-content` (`position: fixed; right: 0; max-width: 550px`). Movie occurrences and series episode rows are plain, non-interactive `<tr>`s.
- `internal/api/handlers.go` `getItem` / `GET /api/v1/items/:id`: already returns the exact `ItemResponse` shape `PlaylistItem` needs (TMDB, pipeline state, raw line/URL, etc.), keyed by `ProcessedLine.ID` - which is exactly `OccurrenceResponse.ID` from the Radarr/Sonarr endpoints. No backend change needed.
- `frontend/src/hooks/useMediaQuery.ts` `useIsMobile()`: existing `(max-width: 767.98px)` breakpoint, already used by both components above for their responsive branching.

## Goals / Non-Goals

**Goals:**
- Reuse the Playlist tab's item detail drawer byte-for-byte in behavior (including force-download) from the Radarr/Sonarr tab, without duplicating its ~250 lines of JSX.
- Keep the existing occurrences sidepanel open and visible while the detail drawer is open on desktop, per the user's explicit direction (stack to the left, don't replace).
- Make series episode rows show their own occurrences inline before drilling further, mirroring the movie occurrence list exactly.

**Non-Goals:**
- No change to the Playlist tab's own behavior or appearance - the extraction must be behavior-preserving there.
- No backend changes - `GET /api/v1/items/:id` is reused as-is.
- No change to how the Films/Séries list pagination, refresh, or independent-error-handling work (already covered by the archived `radarr-sonarr-monitoring-view` change).

## Decisions

### 1. Extract the Playlist drawer into `MediaOccurrenceDrawer`, a self-contained component
Move the existing drawer JSX out of `PlaylistTab.tsx` into a new component taking `item: PlaylistItem | null` and `onOpenChange: (open: boolean) => void`. It owns its own `copiedText`/`forceDownloadStatus`/`forceDownloadError` state internally (reset via the existing `useEffect` pattern keyed on `item?.id`), so neither call site has to manage that state.

`PlaylistTab.tsx` keeps its own `selectedItem` state (a `PlaylistItem` it already has in memory from the list) and passes it straight through - zero behavior change, zero extra fetch.

`RadarrSonarrTab.tsx` does not have a `PlaylistItem` in memory (it only has `OccurrenceResponse` - id/resolution/state). Selecting an occurrence triggers `api.getItem(occurrence.id)` and feeds the resolved `PlaylistItem` into the same `MediaOccurrenceDrawer`.

Alternative considered: keep the drawer inline in `PlaylistTab.tsx` and duplicate a trimmed copy into `RadarrSonarrTab.tsx`. Rejected - the user explicitly asked to reuse "the one that already exists in the Playlist page", and duplicating ~250 lines of JSX is exactly the kind of drift this change is meant to avoid (the previous change already saw the cost of two near-identical matching code paths).

### 2. Second-level drawer is a non-modal `Dialog.Root` without its own `Dialog.Overlay`, offset left via CSS
Radix `Dialog` supports multiple simultaneous `Dialog.Root` instances, but two independent *modal* dialogs each render their own overlay (double-dimming) and independently trap focus (fighting each other for it). Since the occurrences sidepanel must stay visible and interactive-looking underneath, the detail drawer is rendered as a second `Dialog.Root` with `modal={false}` and no `Dialog.Overlay` of its own - it relies on the occurrences sidepanel's existing overlay for the backdrop.

Positioning: a new CSS class `.drawer-content--secondary` (a modifier on the existing `.drawer-content`) sets `right: 550px` instead of `right: 0`, so it sits immediately to the left of the (unchanged) primary drawer. Both share the same `max-width: 550px` and slide-in animation.

```
+----------------------+------------------+------------------+
|                       | MediaOccurrence  | occurrences      |
|   Films / Séries      | Drawer            | sidepanel        |
|   table               | (secondary,       | (primary,        |
|                       |  right: 550px)    |  right: 0)       |
+----------------------+------------------+------------------+
```

Alternative considered: a single `Dialog.Root` whose content swaps between "occurrences list" and "media detail" (with a back button), i.e. in-place navigation instead of stacking. Rejected for desktop because the user explicitly asked for both panels visible side by side; adopted anyway for **mobile** (below), where there is no room to stack.

### 3. Mobile: swap content in place instead of stacking
Below the existing `767.98px` breakpoint (`useIsMobile()`), `RadarrSonarrTab` tracks a small `drawerView: 'occurrences' | 'detail'` state instead of mounting a second `Dialog.Root`. Selecting an occurrence sets `drawerView` to `'detail'` and renders `MediaOccurrenceDrawer` in place of the occurrences content, inside the *same* `Dialog.Content`; a back action sets it back to `'occurrences'`. This reuses the single existing mobile-width `.drawer-content` (already `width: 100%` below its `max-width`) with no new CSS needed.

### 4. Series episode rows expand in place, not via a third overlay
Expanding an episode row is a plain in-panel accordion (local `expandedEpisode: SeasonEpisode | null` state in the sidepanel), not another dialog layer - it reuses the same occurrence-row rendering (resolution + state, clickable) already used for movies, just nested under the episode's `<tr>`. This keeps the "movies -> occurrence -> detail" and "series -> episode -> occurrence -> detail" flows structurally identical from the occurrence row onward, and avoids a three-deep dialog stack.

Season/episode label changes from `S{{season}}E{{episode}}` to zero-padded, space-separated `S{{season}} E{{episode}}`, computed by padding the numbers (`String(n).padStart(2, '0')`) before interpolation - the `i18n` template itself doesn't need padding logic.

### 5. Force-download is back in scope for this view
The archived `radarr-sonarr-monitoring-view` change's design recorded "no write actions from this view" as a non-goal, reasoning that force-download "already exists elsewhere (the Playlist sidepanel)". Reusing that exact drawer here reopens that door deliberately: per the user's direction, once a user can see the specific occurrence behind a "no match" or a stale entry, being unable to act on it from the same panel is the friction this change removes. This is a superseding decision, not a bug in the archived change - it isn't retroactively edited.

## Risks / Trade-offs

- **[Risk]** A non-modal secondary `Dialog.Root` means clicking outside both panels closes only the primary one (Radix's default outside-click behavior on the modal dialog), while the non-modal secondary one has no built-in outside-click-to-close → **Mitigation**: explicitly close the secondary drawer whenever the primary one closes (movie/series selection changes or is cleared), so it can never outlive its parent sidepanel.
- **[Risk]** Two 550px-wide panels plus the base container can exceed common laptop viewport widths, causing horizontal scroll → **Mitigation**: accepted for now (matches the user's explicit "side by side" direction); the existing `max-width: 1600px` container plus two 550px drawers still fits typical 1440px+ desktop viewports without the drawers themselves scrolling internally. Not addressed further in this change.
- **[Risk]** `MediaOccurrenceDrawer` extraction could subtly change Playlist tab behavior if some state/prop is missed → **Mitigation**: extraction is mechanical (move JSX + its local state as-is), verified by manually re-testing the Playlist tab's drawer after extraction (open matched/unmatched items, force-download, copy actions).

## Migration Plan

Purely additive/refactor: no data model, API, or config changes. `PlaylistTab.tsx`'s behavior is preserved by construction (same JSX, same state, just relocated). Deploy as a normal release; rollback is a normal revert.
