## Context

See proposal.md - Why. `FilterDryRunPanel` (`frontend/src/components/FilterDryRunPanel.tsx`) is shared, unchanged, between `CreateFilterDialog` (a Radix `Dialog.Content` with `position: fixed`, `top/left: 50%`, and no `max-height`/`overflow-y` in `.dialog-content`, `frontend/src/index.css:538`) and `FiltersSection`'s `FilterCardTester` (a card embedded in normal page flow, which already scrolls with the page). The backend already caps result volume (`filterDryRunTopValues = 20`, `filterDryRunSearchCap = 100`, `internal/api/filter_dryrun_handlers.go`) but nothing on the frontend caps the *rendered height* of that volume.

The codebase already has a precedent for a height-capped, internally scrollable block: `.technical-block` (`max-height: 150px; overflow-y: auto`, `frontend/src/index.css`), used for the copiable technical config block.

## Goals / Non-Goals

**Goals:**
- Keep the dry-run summary and search results readable and navigable without the containing dialog growing past the viewport.
- Apply the same rendering in both places `FilterDryRunPanel` is used (`CreateFilterDialog`, `FiltersSection` cards), since it's one shared component.
- Add a generic `.dialog-content` height/overflow safety net so other dialogs with variable-length content get the same protection, not just this one.

**Non-Goals:**
- No change to the dry-run API contract, request/response shapes, or backend caps (20 top values, 100 search lines) - purely a rendering change.
- No change to `FiltersSection`'s own scroll behavior (page-level scroll already handles it there; the fix is about the panel's internal layout, which benefits that context too but isn't required by it).
- No redesign of unrelated dialogs beyond giving `.dialog-content` a shared max-height/overflow rule.

## Decisions

**Top matched / top excluded as a two-column grid.** Replace the two stacked `<div><strong/><ul/></div>` blocks with a two-column CSS grid (`display: grid; grid-template-columns: 1fr 1fr; gap: ...`), each column keeping its own heading and list. At 20 items max per side this halves the worst-case height without needing its own scroll container - unlike the search results, this section's volume is already bounded, so a scrollbar here would add complexity for little benefit. On narrow viewports (mobile), collapse to a single column (`grid-template-columns: 1fr` under a small-viewport media query or a `flex-wrap`-based fallback, matching how other two-column layouts in this codebase already respond to width, e.g. `frontend-responsive-layout`) rather than truncating.

**Search results as a height-capped scrollable table.** Replace the per-line `<div style={{display:'flex', justifyContent:'space-between'}}>` rows with an actual `<table>` (value column, matched/excluded status column), wrapped in a container styled like `.technical-block`'s scroll pattern: a fixed `max-height` (around 220-260px - enough to show ~6-8 rows before scrolling, small enough to never dominate the 500px-wide dialog) and `overflow-y: auto`. The `truncated` flag/message (`dryRun.truncated`) renders below the table, outside the scroll container, so it's always visible without scrolling.

**Generic `.dialog-content` safety net.** Add `max-height: 85vh; overflow-y: auto;` to the existing `.dialog-content` rule in `index.css`. This is a small, low-risk addition: dialogs shorter than 85vh (the common case today) are visually unaffected, and any dialog whose content would otherwise overflow the viewport now scrolls internally instead of clipping or pushing action buttons off-screen. This is a defense-in-depth measure on top of (not instead of) capping the dry-run panel's own sections - the panel-level caps keep the *common* case compact and readable; the dialog-level cap is the backstop for whatever still doesn't fit.

**No new CSS classes duplicated per-usage.** Because `FilterDryRunPanel` is shared, the new grid/table/scroll styles live once (either as new utility classes in `index.css` alongside `.technical-block`, or as inline styles in the component consistent with its current inline-style convention - implementation detail left to the tasks/coding step) and apply identically in both `CreateFilterDialog` and `FiltersSection`.

## Risks / Trade-offs

- [A ~250px scroll container inside a ~500px-wide dialog could feel cramped on very small screens] → Mitigation: reuse the existing `.technical-block` max-height precedent (150px) as a lower bound; the responsive layout capability (`frontend-responsive-layout`) already handles narrow-viewport dialog sizing, and this change doesn't need to alter that.
- [Adding `overflow-y: auto` to `.dialog-content` globally could interact with any dialog that relies on content overflowing visibly, e.g. a dropdown menu rendered inside a dialog] → Mitigation: `max-height: 85vh` is generous (most dialogs are well under that), and Radix `Dialog`/`Select` portals typically render popovers outside the `Dialog.Content` DOM subtree already, so this is expected to be low-risk; worth a quick manual check of the other dialogs (`CreateM3uSourceDialog`-equivalent, settings dialogs) during implementation.
