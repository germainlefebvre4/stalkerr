## Context

See proposal.md - Why. Relevant current-state constraints:
- `PlaylistTab.tsx` already calls `useIsMobile()` (from `frontend/src/hooks/useMediaQuery.ts`) to branch between the `<table>` and the stacked mobile card list; this change reuses that same hook rather than introducing a new breakpoint mechanism.
- The advanced filter grid (media name, group, TMDB, pipeline state) and the content-type pills are two separate `<div>` blocks already, both inside the same wrapper `<div style={{ display: 'flex', flexDirection: 'column', ... }}>` at the top of `PlaylistTab.tsx` — no existing shared/generic collapsible component exists anywhere in `frontend/src`.
- Filter values themselves (search, searchName, tmdbFilter, stateFilter) are owned by `usePlaylist.ts` and persisted via the shared `useURLState` hook; this change only wraps the advanced grid's container in mobile markup and does not touch how those values are read, written, or persisted.
- No Radix `Collapsible` package is installed (only `react-tabs`, `react-dialog`, `react-progress`, and one more per `package.json`).

## Goals / Non-Goals

**Goals:**
- Collapse only the advanced filter grid on mobile, behind a local (non-persisted) `useState`, defaulting to collapsed on every mount.
- Show an active-filter count on the disclosure header, computed from the same filter props the component already receives.
- Leave desktop rendering, the content-type pills, and all filter-value state/persistence untouched.

**Non-Goals:**
- No new shared/generic "collapsible section" component — this is a single call site (`PlaylistTab.tsx`); a reusable abstraction would add indirection for one consumer, consistent with the precedent set in `mobile-responsive-ui`'s design.md for the table/card split.
- No new dependency (no Radix `Collapsible`) — a plain `useState` plus conditional rendering and CSS covers the collapse/expand and matches how `selectedItem`/`gotoPageInput` are already managed locally in this file.
- No change to what counts as an "active" filter value in the URL/localStorage layer — the count is a derived, presentational value only.

## Decisions

**Plain local `useState<boolean>` for expanded/collapsed, not a hook or persisted state.**
The proposal is explicit that this state must reset to collapsed on every mount and must not touch the URL or `localStorage`. A local `const [advancedFiltersOpen, setAdvancedFiltersOpen] = React.useState(false)` inside `PlaylistTab.tsx`, alongside the existing `selectedItem`/`copiedText`/`gotoPageInput` local state, satisfies this directly with no extra module.
- Alternative considered: reuse `useURLState` for the open/closed flag for consistency with other UI state. Rejected — the proposal explicitly calls for no persistence, and wiring a throwaway flag through the URL-state hook would contradict that and add a query param for no benefit.

**Active-filter count computed inline from existing props, not a new prop or context.**
`PlaylistTab` already receives `playlistSearchName`, `playlistSearch`, `playlistTMDBFilter`, `playlistStateFilter` as props. The count is `[searchName, search, tmdbFilter !== 'all', stateFilter !== 'all'].filter(Boolean).length` computed at render time — no new state, no new prop threading from `App.tsx`.

**Disclosure is a plain header `<div>`/`<button>` toggling a conditional render, not a CSS-only `<details>` element.**
Using semantic `<details>/<summary>` would work for pure show/hide but this codebase already renders custom chevron/label markup patterns (e.g. `renderSortableHeader`'s `▲`/`▼` indicator) and keeps every interactive control as `<button>` + `onClick` + inline styles for consistent styling with the rest of the file (buttons, badges). A `<button>` toggling `advancedFiltersOpen` matches that existing convention and keeps the active-filter-count badge easy to render as a sibling span, which is more awkward inside `<summary>`.
- Alternative considered: `<details>/<summary>`. Rejected for markup/styling consistency with the rest of the file, not for a technical limitation.

**Gating is `isMobile` only, reusing the existing branch — no new breakpoint or media query.**
The advanced filter grid's collapsible wrapper renders only when `isMobile` is true; when false, the grid renders exactly as it does today (inline, always visible), by keeping the existing JSX for that branch unchanged and only adding the new wrapper in the mobile branch.

## Risks / Trade-offs

- [A filter restored from the URL on mobile is active but visually hidden behind the collapsed disclosure] → Mitigated by the active-filter count on the header (see spec scenario "Active filter count is visible while collapsed"); no auto-expand is added because the proposal explicitly chose collapsed-always-by-default over auto-expand-if-active.
- [Two rendering branches (mobile collapsible vs desktop inline) for the same advanced filter grid increase the surface to keep in sync if a filter field is added later] → Same mitigation already adopted in `mobile-responsive-ui`'s design.md for the table/card split: keep both branches in the same file, near each other, and treat "add a filter field" as "update both branches."

## Migration Plan

Purely additive/visual frontend change behind the existing mobile breakpoint hook — no data migration, no API change, no feature flag. Ship as a normal frontend deploy; rollback is a plain revert of the frontend build if a regression appears.
