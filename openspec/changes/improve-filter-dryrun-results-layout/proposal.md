## Why

The filter dry-run panel (`FilterDryRunPanel`, shared by `CreateFilterDialog` and the `FiltersSection` filter cards) stacks every result block vertically with no height limit: up to 20 "top matched" values, up to 20 "top excluded" values, and up to 100 content-search result lines. Inside `CreateFilterDialog`, a fixed-position, centered `Dialog.Content` with no `max-height`/`overflow-y` either, a full test-and-search cycle can push ~140 rows into the modal, growing it past the viewport with no scrollbar, making the dialog's own action buttons unreachable and the results unreadable.

## What Changes

- Rework `FilterDryRunPanel`'s result rendering:
  - "Top matched" and "top excluded" values render as two side-by-side columns instead of two stacked lists, halving their combined vertical footprint.
  - Content-search results render as a real two-column table (value / matched-or-excluded status) inside a height-capped, internally scrollable container, instead of an unbounded stack of rows.
- Add a generic safety net on `.dialog-content` (`max-height` + `overflow-y: auto`) in `index.css` so any dialog with variable-length content - not just the dry-run panel - stays within the viewport and scrolls internally instead of growing past it.
- No change to the dry-run request/response contract, the backend caps (20 top values, 100 search lines), or the underlying test/search behavior - this is a presentation-only change to how existing results are laid out.

## Capabilities

### Modified Capabilities
- `frontend-filters-management`: the dry-run summary and content-search requirements currently say results are rendered "in the same place the test was triggered from" without constraining layout; this change adds an explicit requirement that the results display stays height-bounded (columns for top values, a scrollable table for search results) rather than growing the containing dialog indefinitely.

## Impact

- `frontend/src/components/FilterDryRunPanel.tsx`: layout rework of the summary and search-result sections (shared by `CreateFilterDialog` and `FiltersSection`'s `FilterCardTester`).
- `frontend/src/index.css`: new/adjusted rules for `.dialog-content` (max-height + overflow) and for the new result-table/columns containers.
- No backend, API, or type changes.
