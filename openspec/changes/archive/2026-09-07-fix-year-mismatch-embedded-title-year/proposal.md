## Why

`fileparser.Parse` detects a movie/show's year by taking the **first** 4-digit
number matching `\b(19|20)\d{2}\b` anywhere in the download path. When a
title itself contains a year-like number (e.g. `"Valensole 1965"`, released
in 2025), that embedded number is picked up instead of the real release year
that follows in parentheses (`"Valensole 1965 (2025)"`), producing a false
`year_mismatch`/`detected_year` even though the file is named correctly
according to the app's own `Title (Year)` convention.

## What Changes

- In `internal/fileparser/parser.go`, prefer a year found inside parentheses
  (`(YYYY)`, the app's own naming convention) over the first bare 4-digit
  match, using the **last** parenthesized year in the path when several
  exist.
- When no parenthesized year is present, keep the current behavior (first
  bare `\b(19|20)\d{2}\b` match) unchanged, so paths that don't follow the
  `Title (Year)` convention are not affected.
- `has_year_in_path` behavior is unchanged: it stays true whenever any year
  token (parenthesized or bare) is found, regardless of which value is
  ultimately selected as `detected_year`.

## Capabilities

### Modified Capabilities
- `file-metadata-parsing`: year-detection requirement changes to prefer a
  parenthesized year (last occurrence) over an embedded bare year elsewhere
  in the path; adds a documented limitation for paths with multiple
  parenthesized years.

## Impact

- Code: `internal/fileparser/parser.go` (`Parse`, `yearRegex` usage),
  `internal/fileparser/parser_test.go`.
- API: no contract change — `FileInfo.DetectedYear`/`YearMismatch` remain the
  same fields/types in `DownloadEnrichedResponse`; only the computed value
  changes for the affected inputs.
- No DB/migration impact; no other callers of `fileparser.Parse` beyond
  `internal/api/handlers_frontend.go:237`.
