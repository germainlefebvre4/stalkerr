## Why

Filter dry-run testing today is scattered and single-dimensional: each filter card and the create dialog run their own isolated tester, results are cramped inline (in a dialog or under a card), and there is no way to see the *combined* real-world effect of the Group Title and TVG Name filters together — even though a playlist line is only kept in production when it passes both. A user tuning one attribute has no visibility into how the other attribute's active filter would still reject the same lines.

## What Changes

- Introduce a single shared test panel (`FilterTestPanelBody`) and a reusable side drawer (`FilterTestDrawer`, reusing the existing `.drawer-content` pattern) that replaces both the create dialog's inline results and each filter card's own inline tester.
- `CreateFilterDialog` opens this drawer as a non-modal sibling (the same `modal={false}` / `withOverlay={false}` + outside-click-suppression pattern already used by `RunItemsDialog` + `MediaOccurrenceDrawer`), so the in-progress form stays visible and editable while results are shown.
- `FiltersSection` lifts a single shared drawer instance for the whole "Filtres" section (replacing the per-card `FilterCardTester` state); the source selector moves from each card into the drawer itself, so a card's "Tester" action is a plain trigger.
- Add a new "Tester l'ensemble" entry point that dry-run tests the currently *effective* patterns (active override if present, else origin `config.yml`) for **both** attributes together, mirroring the production AND-logic (`filter.Manager.ShouldProcess`) instead of testing one attribute in isolation.
- The combined test's summary reports a cause-ventilated breakdown (kept, excluded by Group Title only, excluded by TVG Name only, excluded by both) instead of a single matched/excluded count.
- The content-search field, when testing combined mode, lets the user choose which attribute (Group Title or TVG Name) to search against.
- Generalize the dry-run backend endpoint to accept pattern sets for both attributes in the same request (either can be empty, which preserves today's single-attribute behavior as the case where the other attribute's patterns are empty), and to return the new cause-ventilated summary shape when both are supplied.

## Capabilities

### New Capabilities
_None — this extends two existing capabilities rather than introducing a new one._

### Modified Capabilities
- `frontend-filters-management`: dry-run testing and content-search move from inline-per-trigger-site rendering to a single shared test drawer with the source selector inside it; adds the combined ("test everything") entry point, its cause-ventilated summary display, and attribute-choice for content search in combined mode.
- `filter-dry-run-test`: the endpoint accepts pattern sets for both `group_title` and `tvg_name` in one request instead of exactly one attribute, returns a combined/cause-ventilated summary when both are supplied, and content search accepts a caller-chosen attribute in that case.

## Impact

- Frontend: `frontend/src/components/FilterDryRunPanel.tsx` (reworked into the shared `FilterTestPanelBody`), new `FilterTestDrawer.tsx`, `CreateFilterDialog.tsx` (drawer wiring instead of inline results), `FiltersSection.tsx` (lifted shared drawer state, simplified card triggers, new "Tester l'ensemble" button), i18n additions under `frontend/src/locales/*/filters.json` (and `dialogs.json` as needed). Reuses the existing `.drawer-content` CSS — no new dialog-width variant needed.
- Backend: `internal/api/dto.go` (request/response shapes), `internal/api/filter_dryrun_handlers.go` (combined evaluation mirroring `filter.Manager.ShouldProcess`), `internal/api/filter_dryrun_handlers_test.go`.
- No database schema or `config.yml` schema changes; the dry-run endpoint remains read-only and persists nothing.
- Builds on top of the (implemented, not yet archived) `improve-filter-dryrun-results-layout` change: the two-column top-values grid and the height-capped, scrollable search-results table it introduced are preserved as-is inside the new drawer body.
