## 1. Shared Pattern Parsing

- [x] 1.1 Add `filter.ParsePatternList(raw string) []string` in `internal/filter/filter.go`: split on `,`, trim whitespace per token, drop empty tokens.
- [x] 1.2 Unit test `ParsePatternList` for: empty string, single pattern, multiple patterns, trailing comma, extra whitespace.

## 2. Loading Engine - Format Fix and Isolation

- [x] 2.1 Update `Manager.LoadFromDatabase()` to build `includePatterns`/`excludePatterns` via `ParsePatternList(*dbFilter.IncludePatterns)` / `ParsePatternList(*dbFilter.ExcludePatterns)` instead of `json.Unmarshal`.
- [x] 2.2 Update `Manager.LoadFromConfig()` so a `loadFilterSet` compile failure for one attribute is logged (attribute, offending pattern, error) via `internal/logger` and that attribute is skipped, instead of returning an error that aborts the whole function.
- [x] 2.3 Update `Manager.LoadFromDatabase()` so a `loadFilterSet` compile failure for one runtime filter row is logged (attribute, filter name, offending pattern, error) and that row is skipped, instead of aborting the whole function.
- [x] 2.4 Confirm `Manager.LoadAll()` no longer short-circuits before `LoadFromDatabase()` runs when `LoadFromConfig()` encounters a bad pattern. (Confirmed: `LoadFromConfig()` now always returns `nil` for pattern-compile failures, so `LoadAll()` naturally proceeds to `LoadFromDatabase()` — no code change needed in `LoadAll()` itself.)
- [x] 2.5 Update/add tests in `internal/filter/filter_test.go`: one bad attribute in `LoadFromConfig` doesn't block the other attribute or `LoadFromDatabase`; one bad row in `LoadFromDatabase` doesn't block other rows.

## 3. Pattern Validation at Write Time

- [x] 3.1 In `createFilter` (`internal/api/handlers.go`), before persisting, validate every token of `include_patterns`/`exclude_patterns` (via `ParsePatternList` + `filter.ValidatePattern`) and respond `400 invalid_pattern` on the first failure without creating the row.
- [x] 3.2 In `updateFilter`, apply the same validation to any submitted `IncludePatterns`/`ExcludePatterns` before saving; on failure respond `400 invalid_pattern` and leave the existing row unchanged.
- [x] 3.3 Add backend tests in `internal/api/handlers_filters_test.go` covering: create with an invalid pattern (rejected, not persisted), update with an invalid pattern (rejected, original row unchanged).

## 4. Frontend - Surface Invalid Pattern Errors

- [x] 4.1 Add an `invalid_pattern` translation key to `frontend/src/locales/{en,fr}/common.json` (where `filter_create_failed` and the rest of the `errors.*` map already live) so `useApiErrorMessage` resolves it to a specific message.
- [x] 4.2 Verify `CreateFilterDialog.tsx` already displays this via its existing `translateApiError`/`filterError` path. (Confirmed — no code change needed; the dialog's generic catch handler already resolves any `ErrorResponse.error` code through `errors.<code>`.)

## 5. Verification

- [x] 5.1 Run `go test ./internal/filter/... ./internal/api/...` and confirm all pass. (71 passed across `internal/filter`, `internal/api`, `internal/config`.)
- [x] 5.2 Manually verify: fix the local `config.yml` `tvg_name` typo is NOT required for the isolation fix to work. (Verified against the real, still-unfixed `config.yml` and the live Postgres DB via a throwaway diagnostic script: `LoadAll()` returns nil, logs `"skipping tvg_name origin filter"` with the exact offending pattern, `group_title`'s origin + a live runtime override both load and match correctly, and `tvg_name` falls back to permissive/allow-all instead of taking down everything else. Script deleted immediately after.)
- [x] 5.3 Manually verify: attempt to create a filter with an invalid pattern via the UI and confirm it is rejected with a clear message and not persisted. (Verified live in a browser: submitting `"*"` as an include pattern shows "⚠️ One of the patterns is not a valid regular expression. Please fix it and try again." inline, dialog stays open, and the Filters page shows no new override afterward.)
