## Context

`StatsKPICards` (`frontend/src/components/StatsKPICards.tsx`) is a stateless presentational component rendered in `App.tsx` above `Tabs.Root`, so it shows on every tab. Mobile/desktop layout switches purely via CSS media queries (`max-width: 767.98px` in `index.css`) — there is no JS-level viewport branching today. `compact-downloads-mobile-cards` (separate, in-progress change) introduces the same collapsed/tap-to-expand pattern for individual Downloads cards; this change reuses that precedent for the KPI header. See `proposal.md` for motivation and `specs/frontend-responsive-layout/spec.md` for the target behavior.

## Goals / Non-Goals

**Goals:**
- Collapse the KPI grid to a single toggle row on mobile, expandable/collapsible by tap.
- Keep the toggle state local and simple — no persistence, no coupling to other app state.
- Reuse the CSS-visibility-toggle pattern already established for Downloads cards, for consistency.

**Non-Goals:**
- No change to stats fetching, polling, or the `/api/v1/stats` data shape.
- No persistence of collapsed/expanded state across reloads or navigation (resets to collapsed on mobile every time, matching the "no new shared/global expand-state" precedent from the sibling change).
- No `useMediaQuery`-style hook — stays consistent with the codebase's existing CSS-only breakpoint approach.

## Decisions

**Expand state: single boolean `useState` local to `StatsKPICards`, not lifted to `App.tsx`.**
Unlike the Downloads list (many cards needing independent per-id state), there's exactly one KPI header instance, so a single local boolean is sufficient. `App.tsx` has no other reason to know about this state, so keeping it local avoids unnecessary prop plumbing.

**Collapse/expand is CSS-visibility only, not conditional unmount.**
The 4 KPI cards render in the DOM at all times; a class toggle (e.g. `.kpi-grid--collapsed`) controls visibility via CSS, matching the approach used for Downloads card accordions. This keeps desktop rendering (always expanded) a no-op case of the same markup, and avoids remount cost on toggle.

**Mobile-only via CSS, no JS viewport check.**
The toggle row (label + chevron) renders unconditionally but is hidden on desktop via the existing `@media (max-width: 767.98px)` block; the KPI grid is force-visible on desktop via the same media query overriding the collapsed class. Default collapsed state is also CSS-driven (the collapsed class is present by default; JS only removes/re-adds it on tap), so desktop never needs JS to override anything.

**Data still fetches regardless of collapsed state.**
`stats` fetching is unrelated to this component's local UI state (per `frontend-ihm-dashboard`'s existing "Display global stats on dashboard load" scenario) — collapsing only hides the rendered cards, it does not skip or defer the fetch. This keeps expand instantaneous (no loading state to handle on first tap).

## Risks / Trade-offs

- [Toggling by tapping the row could be visually inconsistent with the Downloads card's whole-card-tap-minus-button pattern] → Low risk: the KPI toggle row has no other interactive elements inside it (unlike a download card with a "Déplacer" button), so the whole row can safely be the tap target with no `stopPropagation` concerns.
- [A user who wants stats visible on every visit has to re-tap after every reload] → Accepted trade-off per explicit user preference (collapsed-by-default takes priority over persistence); no persistence mechanism introduced.
