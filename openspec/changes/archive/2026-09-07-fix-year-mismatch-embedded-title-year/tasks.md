## 1. Implementation

- [x] 1.1 Add `parenYearRegex = regexp.MustCompile(`\((19|20)\d{2}\)`)` to `internal/fileparser/parser.go`
- [x] 1.2 In `Parse`, use `parenYearRegex.FindAllString(downloadPath, -1)`, take the last match, strip parentheses, and set `detectedYear`/`hasYearInPath` from it when at least one match is found
- [x] 1.3 Keep the existing `yearRegex.FindString(downloadPath)` logic as the fallback path when no parenthesized year is found, and verify no other code path changed
- [x] 1.4 Verify `year_mismatch` computation (`tmdbYear != nil && y != *tmdbYear`) is untouched and still runs against whichever `detectedYear` was selected

## 2. Tests

- [x] 2.1 Add a unit test in `internal/fileparser/parser_test.go` reproducing the reported case (`"Valensole 1965 (2025)/Valensole 1965 (2025).mkv"`, tmdbYear 2025) and verify `detected_year == 2025`, `year_mismatch == false`
- [x] 2.2 Add a unit test for a path with no parenthesized year (e.g. the existing "Example 4" scene-release-style path) and verify the first-bare-match behavior is unchanged (covered by the existing "Ambiguous Year" test, unaffected by this change)
- [x] 2.3 Add a unit test for a path with two parenthesized years and verify the last one is selected as `detected_year` (fixture uses `Title (1984) (2021)`, not `(2021 Remaster)` — see note below)
- [x] 2.4 Review existing fixtures in `internal/fileparser/parser_test.go` for incidental parentheses that aren't meant to denote a year, and update any that would now resolve differently (none found)
- [x] 2.5 Run `go test ./internal/fileparser/...` and verify all tests pass, keeping coverage ≥ 95% per the capability's non-functional requirement (100% coverage, all tests pass)

## 3. Spec sync

- [x] 3.1 After implementation, run `openspec sync` (or equivalent) to fold the `file-metadata-parsing` delta spec into the main spec, including updating the "Example 4: Ambiguous Year" example if it no longer reflects the new fallback-only behavior (Example 4 unchanged — no parens, still fallback path; added Examples 4b/4c for the new precedence behavior)

**Note on 2.3 / design conflict resolved during implementation:** the spec's original "Multiple parenthesized years" example (`Title (1984) (2021 Remaster)`) does not actually match `parenYearRegex` (`\((19|20)\d{2}\)` requires the closing paren immediately after the digits), so it would not have exercised the intended behavior and would have produced `detected_year = 1984`, not `2021`. Per user decision, kept the regex exactly as specified in design.md and updated the spec/design example to `Title (1984) (2021)` instead of loosening the regex. This is documented in `design.md`'s Risks/Trade-offs section and in the synced main spec's Implementation Notes.
