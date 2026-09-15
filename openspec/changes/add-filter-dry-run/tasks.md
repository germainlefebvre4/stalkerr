## 1. Backend: reusable pattern evaluator

- [ ] 1.1 In `internal/filter`, add exported `CompilePatterns(includePatterns, excludePatterns []string) (*CompiledFilter, error)` and `(*CompiledFilter) Matches(value string) bool`, extracting the existing exclude-then-include semantics used by `Manager.Matches`/`loadFilterSet`. Verify with a unit test covering: exclude wins over include, empty include patterns means "include all", invalid regex returns a compile error.
- [ ] 1.2 Refactor `Manager.loadFilterSet`/`Manager.Matches` to use the new `CompiledFilter` internally where it doesn't change existing behavior, keeping `filter_test.go`'s existing tests passing unmodified (no behavior change to the real ingestion path).

## 2. Backend: dry-run endpoint

- [ ] 2.1 Add request/response DTOs to `internal/api/dto.go`: dry-run request (`source_name`, `attribute`, `include_patterns` string, `exclude_patterns` string, optional `search`), and response types for the aggregate-summary shape and the search-results shape (including a `no_archive` condition and a `truncated` flag on search). Verify types compile and match the two response shapes described in `specs/filter-dry-run-test/spec.md`.
- [ ] 2.2 Add `filterDryRun` handler in a new `internal/api/filter_dryrun_handlers.go`: validate `attribute` (`group_title`/`tvg_name` only), resolve `source_name` via `settings.EffectiveSources()` (404/400 if unknown), split `include_patterns`/`exclude_patterns` on `,` (trim, drop empties) and compile them with `filter.CompilePatterns` (400 with the offending pattern on compile failure).
- [ ] 2.3 Resolve the source's archive with `m3udownloader.SourcePaths(src.FilePath, src.Download.ArchiveDir, src.Name)` + `m3udownloader.NewArchiveManager(archiveDir, log).GetLatestArchive()`; when none exists, respond with the distinct "no archive available" condition without attempting a download. Verify with a test asserting no HTTP call/download side effect occurs.
- [ ] 2.4 Parse the resolved archive with `parser.NewParser(path, source_name).Parse()` and evaluate every line's `GroupTitle`/`TvgName` (per `attribute`) against the compiled filter, without writing to the database.
- [ ] 2.5 Implement the aggregate-summary response: total lines scanned, matched count, excluded count, and the top 20 distinct attribute values by line count on each side. Verify with a unit test using a small in-memory/temp M3U fixture with known distinct-value counts.
- [ ] 2.6 Implement the search-results response: when `search` is non-empty, filter lines by case-insensitive substring match on the tested attribute's value only, tag each as matched/excluded, cap at 100 results and set a `truncated` flag when more exist. Verify with a unit test including a truncation case and a zero-results case.
- [ ] 2.7 Wire the route in `internal/api/api.go` (e.g. `filters.POST("/dryrun", s.filterDryRun)`) alongside the existing `filters` group. Verify `GET /api/v1/filters` and friends still resolve correctly (no route conflict).
- [ ] 2.8 Add handler-level tests in `internal/api/filter_dryrun_handlers_test.go` covering: unknown source rejected, unsupported attribute rejected, invalid regex rejected, no-archive condition, successful summary, successful search with truncation. Verify `go test ./internal/api/... ./internal/filter/...` passes.

## 3. Frontend: API client and types

- [ ] 3.1 Add `FilterDryRunRequest`/`FilterDryRunSummary`/`FilterDryRunSearchResult` types to `frontend/src/types.ts` matching the backend response shapes (including the no-archive and truncated conditions).
- [ ] 3.2 Add `api.dryRunFilter(payload)` to `frontend/src/services/api.ts` posting to `/api/v1/filters/dryrun`, following the existing `createFilter`/`testIntegration` method conventions. Verify with a unit test in `api.test.ts` mirroring the existing `testIntegration` test.

## 4. Frontend: shared dry-run result UI

- [ ] 4.1 Create a `FilterDryRunPanel` component (or similar) rendering: idle/testing state, the aggregate summary (counts + top matched/excluded value lists), an inline error for invalid patterns, and the no-archive message - reusable from both the create dialog and a filter card.
- [ ] 4.2 Add the content-search field to the same component, shown only once a summary has been rendered (per `frontend-filters-management` - Dry-Run Content Search), calling `api.dryRunFilter` again with `search` set and rendering each line's matched/excluded tag.
- [ ] 4.3 Add unit tests for the new component covering: idle state hides search, summary renders top values, no-archive message, invalid-pattern error, search results with the would-match/would-be-excluded labels.

## 5. Frontend: wire into CreateFilterDialog

- [ ] 5.1 Pass the M3U sources list (from `useM3uSources`, already fetched in `ConfigurationPage`) into `CreateFilterDialog`, add a source `<select>`, and add a "Tester" button next to the include/exclude fields that calls the dry-run endpoint with the form's current (unsaved) attribute/include/exclude/source values.
- [ ] 5.2 Render `FilterDryRunPanel` within the dialog using that result, without submitting the create request. Verify via `CreateFilterDialog.test.tsx`: testing does not call `createFilter`, and switching attribute/source or editing patterns resets any previous test result.

## 6. Frontend: wire into FiltersSection cards

- [ ] 6.1 Pass the M3U sources list into `FiltersSection`; add a "Tester" action to each filter card (origin block and, when present, the override block), each with its own source selection and its own `FilterDryRunPanel` instance, read-only against that card's existing patterns.
- [ ] 6.2 Verify via `FiltersSection.test.tsx`: testing the origin card uses the origin's patterns, testing the override card uses the override's patterns, and neither triggers a save/delete request.

## 7. i18n

- [ ] 7.1 Add new keys to `frontend/src/locales/{en,fr}/filters.json` and/or `dialogs.json` for: source selector label, "Tester" button, testing/summary/no-archive/invalid-pattern/search strings, and the would-match/would-be-excluded labels. Verify both locale files stay structurally identical (same key set) per `frontend-i18n`.

## 8. End-to-end verification

- [ ] 8.1 Run the full backend and frontend test suites (`go test ./...`, frontend test command) and confirm they pass.
- [ ] 8.2 Manually exercise the feature against a real running instance with at least one M3U source that has a downloaded archive: test a pattern in the create dialog, confirm the summary and search results look correct, then test an existing origin/override card and confirm the same.
