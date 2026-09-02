## Context

The Track Details sidepanel is implemented as `MediaOccurrenceDrawerBody` (content) wrapped by `MediaOccurrenceDrawer` (Radix `Dialog` chrome) in `frontend/src/components/MediaOccurrenceDrawer.tsx`. It is rendered from two places with different prop shapes:
- `PlaylistTab.tsx:329` renders `MediaOccurrenceDrawer` with only `item` and `onOpenChange`. The `onOpenOverride` handler (used to open `ManualOverrideDialog`) already exists as a prop on `PlaylistTab` itself (forwarded from `App.tsx`), but is not currently passed into the drawer.
- `RadarrSonarrTab.tsx` renders `MediaOccurrenceDrawerBody` directly for mobile (line 485) and the full `MediaOccurrenceDrawer` for desktop (line 647). Neither call site has access to an override handler today - `RadarrSonarrTab` has no such prop at all.

`ManualOverrideDialog` itself lives in `App.tsx`, rendered as a sibling of both tabs, driven by `overrideItemData`/`isOverrideOpen` state and the `handleOpenOverride` function (`App.tsx:108,148`). Opening it only requires calling `handleOpenOverride(item)` with the target `PlaylistItem` - no other wiring exists between the drawer and the dialog today.

The existing row-level "Associate" button (`PlaylistItemsTable.tsx:181-200`) is the reference implementation for label, styling (`btn-primary`), and behavior: it just calls `onOpenOverride(item)`.

See proposal.md for motivation; see the modified specs for the target behavior.

## Goals / Non-Goals

**Goals:**
- Give the drawer a way to trigger the existing manual-override flow, without duplicating `ManualOverrideDialog` or its API contract.
- Make that capability reach both drawer entry points (Playlist tab, Radarr/Sonarr tab) through prop threading, not a new global/context mechanism.
- Restructure the drawer's internal layout to a top action bar without touching `ManualOverrideDialog`, the override API, or Force Download's eligibility/state logic.

**Non-Goals:**
- No change to the manual-override dialog's UI, validation, or the `POST /api/v1/items/:id/override` contract.
- No change to Force Download's eligibility rules, backend behavior, or error/queued-state semantics - only its position in the panel and where its status hints render.
- No new shared "action bar" or "button group" component elsewhere in the app; this stays local to the drawer.

## Decisions

**Thread `onOpenOverride` as a plain optional prop, not context.** `MediaOccurrenceDrawerBodyProps` and `MediaOccurrenceDrawerProps` gain `onOpenOverride?: (item: PlaylistItem) => void`. The drawer already receives `item` as a prop and both existing call sites already sit in a component tree that has (or can easily receive) the handler, so prop threading matches the codebase's existing pattern (no context/store is used anywhere else for this). Making it optional means a hypothetical future usage of the drawer without override support degrades to simply hiding the "Associate" button, rather than requiring every call site to supply a no-op.

**Reuse the existing handler end-to-end; do not fork it.** The drawer's "Associate" button calls the same `onOpenOverride(item)` signature already used by `PlaylistItemsTable`, targeting `ManualOverrideDialog` in `App.tsx`. No new dialog, no new API call.

**`RadarrSonarrTab` gains a new required-in-practice prop, threaded from `App.tsx`.** `App.tsx` already holds `handleOpenOverride` and already renders `ManualOverrideDialog` as a sibling; it passes the same function to `RadarrSonarrTab` that it passes to `PlaylistTab`. `RadarrSonarrTab` forwards it to both its mobile (`MediaOccurrenceDrawerBody`) and desktop (`MediaOccurrenceDrawer`) call sites unchanged.

**Action bar is a layout change inside `MediaOccurrenceDrawerBody`, not a new component.** Following the codebase's existing convention (inline flex `<div>` + `.btn-primary`/`.btn-secondary`, no shared `ButtonGroup`/`ActionBar` abstraction anywhere else), the new action bar is a `<div>` with `display: flex; flex-direction: row; gap` placed directly under the header, containing the "Associate" button and the existing "Force Download" button. The Force Download eligibility hint / queued badge / error badge move to a `<div>` immediately below this row, replacing the old standalone "Force Download" section further down the panel.

**"Associate" has no eligibility gating.** Unlike Force Download, "Associate" is always enabled regardless of match/download state, since its purpose (correcting or confirming a TMDB match) is orthogonal to download state - this mirrors the existing row-level Associate button, which is never disabled based on item state.

## Risks / Trade-offs

- [Two prop-drilling hops for `RadarrSonarrTab` (App -> RadarrSonarrTab -> drawer, twice for mobile/desktop)] -> Matches the existing pattern already used for `PlaylistTab`; introducing context/global state for a single handler would be a larger, unrelated refactor.
- [Moving Force Download's status badges out of their dedicated section could reduce visual grouping if the action bar becomes cluttered on narrow drawer widths] -> Keep the badges directly beneath the action bar (own row/wrap), so the visual link between button and status is preserved; verify manually at the drawer's mobile width during implementation.
- [`RadarrSonarrTab`'s two drawer call sites (mobile body-only, desktop full drawer) must both receive the same handler] -> Both are edited together in one pass; no separate behavior is intended between them.
