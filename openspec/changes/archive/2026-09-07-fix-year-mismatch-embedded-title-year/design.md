## Context

See `proposal.md` - Why. Current implementation, `internal/fileparser/parser.go:21-36`:

```go
var yearRegex = regexp.MustCompile(`\b(19|20)\d{2}\b`)
...
yearStr := yearRegex.FindString(downloadPath)
```

`FindString` returns the first match on the full `downloadPath` (folder +
filename concatenated). The app's own completed-download paths already
follow a `Title (Year)` convention (confirmed via
`internal/api/handlers_frontend.go` and real production data), so a
parenthesized year is a reliable signal of the intended release year when
present.

## Goals / Non-Goals

**Goals:**
- Stop misreading a year-like number embedded in a title as the release
  year when a parenthesized year is present in the same path.
- Preserve all current behavior for paths that don't contain a
  parenthesized year (no regression for scene-release-style names).

**Non-Goals:**
- Handling paths with more than one parenthesized year "correctly" beyond
  picking the last one (e.g. distinguishing an original release year from
  a remaster/edition year in a second parenthesized group). Documented as
  a known limitation.
- Changing `has_year_in_path` semantics or the `missing_year` problem
  filter — both stay based on "is any year token present," independent of
  which token is selected as `detected_year`.
- Touching `resolutionRegex`, `is_valid_format`, or any other `FileInfo`
  field.

## Decisions

**Decision: add a parenthesized-year regex, checked before the existing bare-year regex, keeping the bare regex as fallback.**

```go
var parenYearRegex = regexp.MustCompile(`\((19|20)\d{2}\)`)
```

Algorithm in `Parse`:
1. Run `parenYearRegex.FindAllString(downloadPath, -1)`.
2. If at least one match, take the **last** one, strip the parentheses,
   parse it as `detectedYear`, set `hasYearInPath = true`.
3. Otherwise, fall back to the existing `yearRegex.FindString(downloadPath)`
   logic, unchanged.
4. `year_mismatch` comparison against `tmdbYear` is unchanged either way.

Alternatives considered:
- **Take the last bare-year match instead of requiring parentheses**
  (simpler, no new regex). Rejected: less precise, and changes behavior
  for every existing path with multiple bare years (e.g. the spec's
  documented "Example 4: Ambiguous Year" `2001.A.Space.Odyssey.1968.720p`
  case) rather than only the specific parenthesized-convention case this
  bug is about. Larger blast radius for the same fix.
- **Whitelist known numeric titles via TMDB** (e.g. flag when TMDB title
  itself contains a year-like token, suppress mismatch check). Rejected:
  requires passing the title (not just the year) into `fileparser.Parse`,
  widening its signature and coupling it to TMDB title text; the
  parenthesized-year signal is simpler, local to the path string, and
  matches the app's own naming convention directly.
- **Change `has_year_in_path`/fallback to also prefer last-match.**
  Rejected per user's explicit scope decision: only `detected_year`'s
  role in the mismatch comparison is refined; `has_year_in_path`/
  `missing_year` stay on the existing broad "any year token" check.

## Risks / Trade-offs

- [Multiple parenthesized years in one path (e.g. `Title (1984) (2021)`)
  picks the last one, which may not be the canonical release year] →
  Documented as a known limitation (see spec delta scenario); no
  mitigation needed for this iteration since it's not the reported bug
  and no such paths were observed in production data. Note: the
  `parenYearRegex` only matches a strict `(YYYY)` group — a parenthesized
  group with extra text around the year (e.g. `(2021 Remaster)`) is not
  recognized as a parenthesized year at all and falls through to the
  bare-year fallback for that group.
- [A title that itself contains a parenthesized year-like number, e.g.
  `"Movie (1999) Anniversary Edition (2020)"` where 1999 is part of the
  title, not a release year] → Same last-match limitation as above;
  out of scope.
- [Existing unit tests asserting first-bare-match behavior on paths that
  also happen to contain parentheses] → Mitigated by reviewing
  `internal/fileparser/parser_test.go` during implementation and updating
  any test fixtures that use parentheses incidentally rather than to
  denote a year.

## Migration Plan

Pure logic change in a stateless pure function; no data migration, no
feature flag needed. Deploy as a normal code change. Rollback is a plain
revert if regressions surface.
