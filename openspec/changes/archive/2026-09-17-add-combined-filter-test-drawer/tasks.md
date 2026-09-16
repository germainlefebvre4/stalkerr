## 1. Backend: generalize the dry-run request/response

- [x] 1.1 In `internal/api/dto.go`, replace `FilterDryRunRequest`'s single `Attribute`/`IncludePatterns`/`ExcludePatterns` fields with an `Attributes []string` list plus `GroupTitleInclude`/`GroupTitleExclude`/`TvgNameInclude`/`TvgNameExclude` and `SearchAttribute`, per design.md; verify `go build ./...` succeeds
- [x] 1.2 Add `FilterDryRunCombinedSummaryResponse` (kept count + excluded-by-cause counts + each attribute's own top-20 matched/excluded values) and extend `FilterDryRunResultLine` with an optional `Verdict` field, per design.md's shapes; verify `go build ./...` succeeds
- [x] 1.3 In `internal/api/filter_dryrun_handlers.go`, validate `Attributes` (1 or 2 entries, each `group_title`/`tvg_name`, no duplicates, at least one entry required) and `SearchAttribute` (required and must name a supplied attribute when `Attributes` has 2 entries and `Search` is set), rejecting invalid requests per the modified `filter-dry-run-test` Input Validation requirement; verify with new test cases in `internal/api/filter_dryrun_handlers_test.go` for each rejection scenario
- [x] 1.4 Implement combined evaluation: when 2 attributes are supplied, compile both attributes' patterns and, per line, compute kept / excluded-by-group_title-only / excluded-by-tvg_name-only / excluded-by-both, mirroring `filter.Manager.ShouldProcess`'s AND logic (`internal/filter/filter.go:150-163`); verify with a unit test asserting the four counts sum to `total_lines` against a hand-built fixture
- [x] 1.5 Extend the summary-building logic to also produce each attribute's own top-20 matched/excluded values in combined mode; verify with a test asserting each attribute's top values reflect only that attribute's own matched/excluded partition
- [x] 1.6 Extend content-search handling to accept `SearchAttribute`, and in combined mode tag each returned line's `Verdict` (kept / excluded_by_group_title / excluded_by_tvg_name / excluded_by_both) while leaving `Matched` populated and unchanged in single-attribute mode; verify with tests covering both modes plus truncation
- [x] 1.7 Update/extend `internal/api/filter_dryrun_handlers_test.go` to cover every scenario in the modified `filter-dry-run-test` spec (single-attribute request/response unchanged, combined summary, combined search, all Input Validation scenarios); verify `go test ./internal/api/...` passes

## 2. Frontend: shared test panel and drawer

- [x] 2.1 Update `frontend/src/types.ts`'s dry-run request/response types to match the new backend shapes (`attributes`, per-attribute pattern fields, `search_attribute`, `FilterDryRunCombinedSummaryResponse`, `verdict`); verify `tsc` (via `npm run build`) succeeds
- [x] 2.2 Create `frontend/src/components/FilterTestPanelBody.tsx`: the M3U source selector and "Tester" trigger, plus mode-dependent result rendering — single-attribute mode reuses the existing two-column top-values grid and height-capped scrollable search table; combined mode stacks each attribute's own two-column block, shows the four cause-count badges, and lets the user pick which attribute the search field targets. Accepts a `target: FilterTestTarget` and `sources: M3uSource[]`; verify with unit tests covering both modes' rendering
- [x] 2.3 Create `frontend/src/components/FilterTestDrawer.tsx`: a `Dialog.Root` + `.drawer-content` shell with a contextual title derived from `target` (attribute + origin/override/in-progress, or "Tester l'ensemble"), open if and only if `target !== null`, rendering `<FilterTestPanelBody>` inside; accepts `modal`/`withOverlay` passthrough props; verify with a unit test that it renders nothing when `target` is `null` and shows the expected title per mode otherwise
- [x] 2.4 Retire `frontend/src/components/FilterDryRunPanel.tsx` and its inline usage now that `FilterTestPanelBody` supersedes it; verify `npm test` has no remaining references to the old component

## 3. Frontend: CreateFilterDialog wiring

- [x] 3.1 In `CreateFilterDialog.tsx`, replace the inline `FilterDryRunPanel` usage with a local `testTarget: FilterTestTarget | null` state and a `<FilterTestDrawer modal={false} withOverlay={false} target={testTarget} ...>` mounted as a sibling of the dialog's own `Dialog.Content`, copying the `justClosedDrawerRef` outside-click suppression pattern from `RunItemsDialog.tsx:41-61`; verify manually that closing the drawer does not close the create dialog
- [x] 3.2 Ensure the existing `wasOpen` reset effect (`CreateFilterDialog.tsx:38-52`) also clears `testTarget` when the create dialog itself closes, and that editing a pattern after a test invalidates the current target/result exactly as today; verify manually
- [x] 3.3 Update `CreateFilterDialog.test.tsx` to assert clicking "Tester" opens the drawer with a `single` target built from the in-progress attribute/patterns; verify `npm test -- CreateFilterDialog` passes

## 4. Frontend: FiltersSection wiring + combined entry point

- [x] 4.1 In `FiltersSection.tsx`, remove `FilterCardTester`'s per-card state and replace it with a plain "Tester" trigger button per card that calls a lifted `setTestTarget({mode: 'single', ...})`; lift one `testTarget: FilterTestTarget | null` state for the whole section and mount exactly one `<FilterTestDrawer modal target={testTarget} ...>`; verify manually that testing one card then a different card updates the same open drawer instead of opening a second one
- [x] 4.2 Add a "Tester l'ensemble" button next to "Configurer un filtre" in the section header that resolves each attribute's effective patterns (the active override if present, else the origin `config.yml` patterns — reusing the resolution already inlined per card at `FiltersSection.tsx:176-177,205-206`) and opens the drawer with a `combined` target; verify with a unit test asserting the assembled patterns use the override when present and the origin patterns otherwise
- [x] 4.3 Update `FiltersSection.test.tsx` for the lifted single shared drawer and the new "Tester l'ensemble" entry point; verify `npm test -- FiltersSection` passes

## 5. i18n

- [x] 5.1 Add/update translation keys in `frontend/src/locales/{en,fr}/filters.json` (and `dialogs.json` if needed) for the drawer's contextual titles, the "Tester l'ensemble" button, the combined summary's cause labels (kept / excluded by Group Title only / excluded by TVG Name only / excluded by both), and the search-attribute picker; verify both locales render with no missing-key fallback text

## 6. Integration verification

- [x] 6.1 Run the full frontend test suite (`npm test` in `frontend/`) and the full backend test suite (`go test ./...`); verify both pass (2 pre-existing, unrelated pagination test failures in DownloadsTab/ErrorsTab confirmed present before this change)
- [x] 6.2 Using `run`, open the "Filtres" section and test a single card (both an origin-only attribute and one with an active override) and from the create dialog; confirm the shared drawer opens correctly each time with source selection inside the drawer
- [x] 6.3 Using `run`, click "Tester l'ensemble" against a source with both a Group Title and a TVG Name filter configured; confirm the four cause counts sum to the total lines scanned, and that switching the search-attribute picker changes which field is matched
- [x] 6.4 Using `run`, confirm the create dialog's drawer can be opened and closed without closing the dialog itself, and that the dialog's Cancel/Save remain reachable while the drawer is open
