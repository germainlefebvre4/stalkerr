## Context

See proposal.md - Why. Relevant current-state constraints:
- No `@media` query exists anywhere in `frontend/src` today; all spacing is inline `style={}` or fixed values in `index.css`/`variables.css`.
- Navigation is a single Radix `Tabs.Root`/`Tabs.List` driven by `activeTab` state in `App.tsx`, persisted to `localStorage`.
- `PlaylistTab.tsx` and `LogsTab.tsx` each render their own `<table className="custom-table">` with their own column set and row-click handling; there is no shared table component today.
- `DownloadsTab.tsx` already renders a stacked card list (not a table), so it needs density/touch-target tuning only, not a structural rewrite.
- No automated frontend tests exist, so there is no fixture asserting the current `DEFAULT_LIMIT = 50`.

## Goals / Non-Goals

**Goals:**
- One shared, testable definition of "mobile" (`< 768px`) usable from both CSS and JS, so layout decisions stay in sync.
- Bottom tab bar and card-based tables activate purely based on that breakpoint, with no change to how tab/row-click state is managed.
- Keep the drawer sidepanel's content and open/close logic untouched; only its chrome (padding, close target) changes per breakpoint.

**Non-Goals:**
- No new shared/generic "responsive table" abstraction — only two tables (Playlist, Logs) need this, and their columns/actions differ enough that a generic abstraction would add indirection for two call sites.
- No dark mode work (out of scope; the app has no dark theme today).
- No change to Downloads/Logs pagination (explicitly excluded by proposal).
- No visual redesign beyond what's needed for the breakpoint (colors, radii, shadows unchanged).

## Decisions

**Single breakpoint via a shared `useIsMobile()` hook, not CSS-only tricks.**
Table→card swapping needs different JSX (a `<table>` vs a stacked `<div>` list), which pure CSS can't express cleanly for cells with links/buttons/badges. Add `frontend/src/hooks/useMediaQuery.ts` exporting `useIsMobile()`, backed by `window.matchMedia('(max-width: 767.98px)')` with a `change` event listener (no polling, no resize thrashing). The same `767.98px` value is mirrored as a CSS breakpoint for the purely-visual rules (card density, drawer padding, bottom tab bar) so both layers agree on the same cut line.
- Alternative considered: container queries per component. Rejected — one global viewport breakpoint is sufficient for this app's single-column dashboard and avoids introducing a newer, less-supported API for no real benefit here.

**Bottom tab bar is a second, purely-visual nav wired to the same `activeTab`/`setActiveTab` state.**
`App.tsx` keeps its single `Tabs.Root value={activeTab} onValueChange={setActiveTab}`. Below the breakpoint, the existing `Tabs.List` (segmented pills) is hidden via CSS and a fixed-position bottom bar (plain buttons calling `setActiveTab` directly) is rendered instead, reusing the same four labels/icons. This avoids duplicating tab-switch logic or persistence.
- Alternative considered: two separate `Tabs.List` instances (one Radix-driven for desktop, one for mobile) both inside `Tabs.Root`. Rejected in favor of plain buttons for the mobile bar — Radix's `Tabs.Trigger` roving-tabindex/keyboard-nav behavior isn't needed for a bottom bar with visible icons, and plain buttons keep the mobile markup simpler.

**Playlist/Logs render two branches from the same data, not a shared table component.**
Each of `PlaylistTab.tsx` and `LogsTab.tsx` calls `useIsMobile()` and renders either its existing `<table>` (unchanged) or a new inline stacked-card list over the same `playlist`/`logs` array, reusing the same row-click handler (`setSelectedItem`) and the same badge/date helpers already imported. No new shared component; each file keeps its own small card-rendering function next to its existing row-rendering function.

**Full-width desktop table via a flush wrapper, not restructuring the page container.**
`App.tsx`'s `maxWidth: 1600` container and the `.card` (`padding: 1.75rem`+ `Tabs.Content` `padding: 2rem`) wrapper stay as-is for every other tab section (filters bar, headings, KPI cards). Only the table's own wrapper gets a "flush" treatment at `≥768px`: negative margins equal to the accumulated horizontal padding (`calc(-1 * (2rem))`, matching `Tabs.Content`'s inline padding) so the table's outer border reaches the edges of the card, while the filter controls above it keep their normal inset. This is scoped with a new `.table-flush` class applied only to the Playlist/Logs table wrapper.

**Touch targets enforced with a mobile-scoped CSS rule, not per-button edits.**
Rather than editing every `btn-primary`/`btn-secondary`/`btn-danger` inline style, add a `@media (max-width: 767.98px)` block in `index.css` setting `min-height: 44px` (and matching `min-width` for icon-only controls like the drawer close button) on the shared button/badge classes and on the drawer's close control specifically. Buttons already use these shared classes everywhere, so this one rule covers Playlist actions, Downloads' "Déplacer" button, and the drawer close button without touching each call site.

## Risks / Trade-offs

- [Two render branches per table component (table vs cards) increase the surface area to keep in sync when a column is added later] → Mitigate by keeping both branches next to each other in the same file, iterating the same source array, and having code review treat "add a column" as "update both branches" for these two files specifically (documented as a comment at the top of each branch).
- [Fixed bottom tab bar can overlap page content or get obscured by mobile browser chrome/home-indicator] → Add bottom padding to the page container equal to the bar's height (plus `env(safe-area-inset-bottom)`) only below the breakpoint, so scrolled content never sits under the bar.
- [Changing the Playlist default page size from 50→10 changes the "Showing X-Y of Z" numbers users are used to] → This is an intentional, explicit product decision from the proposal; no mitigation needed beyond the spec's scenario coverage. Existing `localStorage`-persisted choices for returning users are unaffected since the default only applies when no preference is stored.
- [`window.matchMedia` is unavailable in non-browser test environments] → Not a current concern since there are no frontend automated tests today; if tests are added later, `useIsMobile()` should be mockable by exporting the underlying query string as a constant.

## Migration Plan

Purely additive/visual frontend change behind a CSS breakpoint and a hook — no data migration, no API version bump, no feature flag needed. Ship as a normal frontend deploy; rollback is a plain revert of the frontend build if a regression appears.
