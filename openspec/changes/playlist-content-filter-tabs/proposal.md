## Why

The Playlist tab's content-type filter (`all` / `movies` / `tvshows`) is rendered as three plain `<button>` elements with a manual `btn-primary`/`btn-secondary` class ternary, while the Radarr/Sonarr suite's sub-tabs (`resume` / `radarr` / `sonarr`) use Radix UI `Tabs` with the shared `segmented-tabs-list`/`segmented-tabs-trigger` styling and `data-state="active"` semantics. The two controls serve the same purpose (a small set of mutually exclusive segments) but look and behave inconsistently, and the content-type filter lacks the tab-role accessibility semantics (roving tabindex, arrow-key navigation) that Radix Tabs provides. Aligning the content-type filter on the same Radix Tabs mechanism gives a consistent look, feel, and interaction pattern across the app.

## What Changes

- Replace the three `<button>` elements for the content-type filter in `PlaylistTab.tsx` with a Radix UI `Tabs.Root` / `Tabs.List` / `Tabs.Trigger` control, mirroring the pattern used by `RadarrSonarrTab.tsx`.
- Reuse the shared `segmented-tabs-list` / `segmented-tabs-trigger` CSS classes so the active segment is indicated via `data-state="active"` instead of the current manual `btn-primary`/`btn-secondary` ternary.
- Add a dedicated modifier CSS class (analogous to `radarr-sonarr-subtabs`) for the content-type filter's own mobile-responsive sizing, since it must remain visible outside the mobile advanced-filters disclosure.
- No `Tabs.Content` panels are introduced: the content-type filter has no per-segment panel, it only narrows a single fetched/rendered list, so switching segments continues to just change the `?type=` query param and re-fetch, exactly as today.
- Filter state management is unchanged: `?type=all|movies|tvshows` continues to be owned by `useURLState` in `usePlaylist.ts`, and selecting a segment still resets the current page to 1.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `frontend-ihm-dashboard`: the content-type filter under "Requirement: Playlist View with Filtering" changes its rendering mechanism from plain buttons to a Radix UI Tabs-based segmented control (ARIA tab semantics, arrow-key navigation, `data-state="active"` styling), while preserving all existing filtering, URL-sync, and mobile-visibility behavior.

## Impact

- `frontend/src/components/PlaylistTab.tsx`: content-type filter markup replaced with Radix `Tabs.Root`/`Tabs.List`/`Tabs.Trigger`.
- `frontend/src/index.css`: new modifier class for the content-type filter's mobile-responsive layout, reusing existing `segmented-tabs-list`/`segmented-tabs-trigger` rules.
- No backend, API, or state-management changes; `usePlaylist.ts` and `useURLState.ts` are unaffected.
