## 1. Top matched / top excluded columns

- [x] 1.1 In `FilterDryRunPanel.tsx`, replace the stacked "top matched" and "top excluded" blocks with a two-column grid (side by side), collapsing to one column on narrow viewports; verify visually (or via an existing responsive breakpoint check) that both columns render at once above the collapse breakpoint
- [x] 1.2 Update/extend `FilterDryRunPanel.test.tsx` to assert both `top_matched` and `top_excluded` values are present in the rendered output when a summary with both is provided, and that the two lists are no longer nested one below the other in a single stacked container (e.g. assert on the new grid container structure)

## 2. Search results as a scrollable table

- [x] 2.1 In `FilterDryRunPanel.tsx`, replace the per-line flex `<div>` rows with a `<table>` (value column, matched/excluded status column) wrapped in a container with a fixed `max-height` and `overflow-y: auto`, following the `.technical-block` scroll pattern in `index.css`
- [x] 2.2 Keep the `truncated` message (`dryRun.truncated`) rendered outside/below the scrollable table container so it stays visible without scrolling
- [x] 2.3 Update `FilterDryRunPanel.test.tsx` to assert the search results render as a table (`role="table"` or equivalent) and that the truncated message still renders when `searchResult.truncated` is true

## 3. Dialog-level safety net

- [x] 3.1 Add `max-height: 85vh; overflow-y: auto;` to `.dialog-content` in `frontend/src/index.css`
- [x] 3.2 Manually verify (via `run`/dev server) that other dialogs using `.dialog-content` (e.g. the M3U source creation dialog, settings dialogs) still render correctly and that any nested dropdown/select popovers are not visually clipped by the new overflow

## 4. Integration verification

- [x] 4.1 Run the frontend test suite (`npm test` in `frontend/`) and verify `FilterDryRunPanel.test.tsx`, `CreateFilterDialog.test.tsx`, and `FiltersSection.test.tsx` all pass
- [x] 4.2 Using `run`, open the Configuration page's "Filtres" section, trigger "Créer un filtre", run a dry-run test against a source with many distinct group/channel names, and confirm: the top matched/excluded columns sit side by side, a content search with many matches scrolls within its own table instead of growing the dialog, and the dialog's Cancel/Save buttons stay reachable without page-level scrolling
- [x] 4.3 Repeat the manual check from 4.2 for a filter card's inline tester in `FiltersSection` (outside the dialog) to confirm the same layout applies there
