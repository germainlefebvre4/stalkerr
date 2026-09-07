# File Metadata Parsing

## Purpose

A pure Go package that extracts technical file metadata from download paths using regex pattern matching. Provides structured information about video files including extension, resolution, year detection, and format validation.

## Requirements

### Requirement: Structured Metadata Extraction From a Download Path
Given a download path and an optional reference year (e.g. from TMDB), the parser SHALL return a FileInfo struct containing: file extension (lowercase, with leading dot), folder name (last directory component), file name (basename), a year-detected flag, the detected year (if any), a year-mismatch flag (detected year vs. reference year), the detected resolution (if any), and a format-validity flag.

#### Scenario: Complete movie path yields fully populated metadata
- **WHEN** Parse is called with path `/media/movies/Matrix.1999.1080p.BluRay.x264/matrix.mkv` and reference year `1999`
- **THEN** it SHALL return extension `.mkv`, folder name `Matrix.1999.1080p.BluRay.x264`, file name `matrix.mkv`, has-year-in-path `true`, year mismatch `false`, detected year `1999`, detected resolution `1080p`, and valid format `true`

#### Scenario: TV show path yields season-folder metadata with no year
- **WHEN** Parse is called with path `/media/tvshows/Breaking.Bad/Season.05/Breaking.Bad.S05E14.1080p.mkv` and no reference year
- **THEN** it SHALL return folder name `Season.05`, file name `Breaking.Bad.S05E14.1080p.mkv`, has-year-in-path `false`, detected year unset, detected resolution `1080p`, and valid format `true`

### Requirement: Null or Empty Path Returns No Result
The parser SHALL return nil when called with a null or empty download path.

#### Scenario: Empty path returns nil
- **WHEN** Parse is called with an empty or null path
- **THEN** it SHALL return nil

### Requirement: Parenthesized Year Takes Precedence Over a Bare Year Match
The parser SHALL treat the last year enclosed in parentheses (`(YYYY)`) in the path as the detected year, in preference to any bare 4-digit year found elsewhere in the path (including one embedded in the title itself), and SHALL fall back to the first bare `\b(19|20)\d{2}\b` match only when no parenthesized year is present.

#### Scenario: Embedded year-like title number does not shadow the parenthesized year
- **WHEN** Parse is called with path `/media/movies/Valensole 1965 (2025)/Valensole 1965 (2025).mkv` and reference year `2025`
- **THEN** the detected year SHALL be `2025`, not the embedded `1965`

#### Scenario: Multiple parenthesized years use the last one
- **WHEN** Parse is called with path `/media/movies/Title (1984) (2021)/movie.mkv` and reference year `2021`
- **THEN** the detected year SHALL be `2021`, not `1984`

#### Scenario: No parenthesized year falls back to the first bare year match
- **WHEN** Parse is called with path `/media/movies/2001.A.Space.Odyssey.1968.720p/movie.mkv` and reference year `1968`
- **THEN** the detected year SHALL be `2001` (the first bare match), and year mismatch SHALL be `true` since `2001 != 1968`

### Requirement: Year Mismatch Detection
The parser SHALL set the year-mismatch flag to `true` when the detected year differs from the supplied reference year, and to `false` when no reference year is supplied, no year is detected, or the detected year matches the reference year.

#### Scenario: No year detected and no mismatch reported
- **WHEN** Parse is called with path `/media/movies/Avatar.BluRay.1080p/avatar.mkv` and reference year `2009`
- **THEN** has-year-in-path SHALL be `false`, detected year SHALL be unset, and year mismatch SHALL be `false`

#### Scenario: Detected year differs from reference year
- **WHEN** Parse is called with path `/media/movies/Avatar.2010.1080p/avatar.mkv` and reference year `2009`
- **THEN** the detected year SHALL be `2010` and year mismatch SHALL be `true`

### Requirement: Resolution Detection
The parser SHALL extract a resolution indicator, case-insensitively, when the path contains one of `2160p`, `4K`, `1080p`, `720p`, `480p`, `360p`, and SHALL leave the detected resolution unset otherwise.

#### Scenario: Resolution present in path
- **WHEN** Parse is called with a path containing `1080p`
- **THEN** the detected resolution SHALL be `1080p`

#### Scenario: Resolution absent from path
- **WHEN** Parse is called with a path containing no resolution token, e.g. `Valensole 1965 (2025)/Valensole 1965 (2025).mkv`
- **THEN** the detected resolution SHALL be unset

### Requirement: File Format Validation
The parser SHALL set the format-validity flag to `true` when the file extension, case-insensitively, is one of `.mkv`, `.mp4`, `.avi`, `.mov`, `.m4v`, `.wmv`, `.flv`, `.webm`, and to `false` otherwise.

#### Scenario: Recognized video extension
- **WHEN** Parse is called with a path ending in `.mkv`
- **THEN** the format-validity flag SHALL be `true`

#### Scenario: Unrecognized extension
- **WHEN** Parse is called with a path whose extension is not in the supported list
- **THEN** the format-validity flag SHALL be `false`

### Requirement: Safe for Concurrent Use Without Filesystem Access
The parser SHALL operate purely on the input string, performing no filesystem access, and SHALL be safe to call concurrently from multiple goroutines without external synchronization.

#### Scenario: Concurrent calls produce independent, correct results
- **WHEN** Parse is called concurrently from multiple goroutines with different paths
- **THEN** each call SHALL return the result corresponding to its own input, independent of any other concurrent call

## Examples

### Example 1: Complete Movie Path

```go
Input:
  path = "/media/movies/Matrix.1999.1080p.BluRay.x264/matrix.mkv"
  tmdbYear = 1999

Output:
  &FileInfo{
    Extension:       ".mkv",
    FolderName:      "Matrix.1999.1080p.BluRay.x264",
    FileName:        "matrix.mkv",
    HasYearInPath:   true,
    YearMismatch:    false,
    DetectedYear:    1999,
    DetectedRes:     "1080p",
    IsValidFormat:   true,
  }
```

### Example 2: Missing Year

```go
Input:
  path = "/media/movies/Avatar.BluRay.1080p/avatar.mkv"
  tmdbYear = 2009

Output:
  &FileInfo{
    Extension:       ".mkv",
    FolderName:      "Avatar.BluRay.1080p",
    FileName:        "avatar.mkv",
    HasYearInPath:   false,
    YearMismatch:    false,
    DetectedYear:    nil,
    DetectedRes:     "1080p",
    IsValidFormat:   true,
  }
```

### Example 3: Year Mismatch

```go
Input:
  path = "/media/movies/Avatar.2010.1080p/avatar.mkv"
  tmdbYear = 2009

Output:
  &FileInfo{
    Extension:       ".mkv",
    FolderName:      "Avatar.2010.1080p",
    FileName:        "avatar.mkv",
    HasYearInPath:   true,
    YearMismatch:    true,
    DetectedYear:    2010,
    DetectedRes:     "1080p",
    IsValidFormat:   true,
  }
```

### Example 4: Ambiguous Year (no parenthesized year)

```go
Input:
  path = "/media/movies/2001.A.Space.Odyssey.1968.720p/movie.mkv"
  tmdbYear = 1968

Output:
  &FileInfo{
    Extension:       ".mkv",
    FolderName:      "2001.A.Space.Odyssey.1968.720p",
    FileName:        "movie.mkv",
    HasYearInPath:   true,
    YearMismatch:    true,  // 2001 != 1968
    DetectedYear:    2001,  // First bare match (no parenthesized year present)
    DetectedRes:     "720p",
    IsValidFormat:   true,
  }
```

### Example 4b: Embedded Year-Like Title Number, Parenthesized Year Wins

```go
Input:
  path = "/media/movies/Valensole 1965 (2025)/Valensole 1965 (2025).mkv"
  tmdbYear = 2025

Output:
  &FileInfo{
    Extension:       ".mkv",
    FolderName:      "Valensole 1965 (2025)",
    FileName:        "Valensole 1965 (2025).mkv",
    HasYearInPath:   true,
    YearMismatch:    false,
    DetectedYear:    2025,  // Last parenthesized year, not the embedded "1965"
    DetectedRes:     nil,
    IsValidFormat:   true,
  }
```

### Example 4c: Multiple Parenthesized Years

```go
Input:
  path = "/media/movies/Title (1984) (2021)/movie.mkv"
  tmdbYear = 2021

Output:
  &FileInfo{
    Extension:       ".mkv",
    FolderName:      "Title (1984) (2021)",
    FileName:        "movie.mkv",
    HasYearInPath:   true,
    YearMismatch:    false,
    DetectedYear:    2021,  // Last parenthesized year found
    DetectedRes:     nil,
    IsValidFormat:   true,
  }
```

### Example 5: TV Show

```go
Input:
  path = "/media/tvshows/Breaking.Bad/Season.05/Breaking.Bad.S05E14.1080p.mkv"
  tmdbYear = nil

Output:
  &FileInfo{
    Extension:       ".mkv",
    FolderName:      "Season.05",
    FileName:        "Breaking.Bad.S05E14.1080p.mkv",
    HasYearInPath:   false,
    YearMismatch:    false,
    DetectedYear:    nil,
    DetectedRes:     "1080p",
    IsValidFormat:   true,
  }
```

## Implementation Notes

**File Location**: `internal/fileparser/parser.go`

**No external dependencies**: the parser SHALL be implemented using only the Go standard library (pure stdlib).

**Regex Patterns**:
- Parenthesized year (checked first): `\((19|20)\d{2}\)` (matches a strict `(YYYY)` group only; a group with extra text around the year, e.g. `(2021 Remaster)`, does not match)
- Year (fallback, used only when no parenthesized year matches): `\b(19|20)\d{2}\b` (matches 1900-2099 with word boundaries)
- Resolution: `(?i)\b(2160p|4K|1080p|720p|480p|360p)\b` (case-insensitive)

**Valid Extensions**: Map-based lookup for O(1) validation
- `.mkv`, `.mp4`, `.avi`, `.mov`, `.m4v`, `.wmv`, `.flv`, `.webm`

**Edge Cases to Handle**:
- Empty/null paths: return nil
- Paths without extension: extension = ""
- Paths without year: detected_year = nil
- Paths without resolution: detected_res = nil
- Multiple years, no parenthesized year: use first bare regex match
- Multiple parenthesized years (e.g. `Title (1984) (2021)`): use the last parenthesized match; picking a non-canonical year (e.g. an edition/remaster year) when more than one parenthesized year is present is a known, accepted limitation
- Case variations: normalize to lowercase for extensions, preserve case for paths

## Testing Requirements

**Unit Tests** (`internal/fileparser/parser_test.go`):
- Valid movie paths with all fields
- Valid TV show paths
- Missing year scenarios
- Year mismatch scenarios
- Ambiguous year scenarios (multiple years in path, no parenthesized year)
- Parenthesized-year precedence over an embedded year-like title number
- Multiple parenthesized years (last one selected)
- Invalid/unknown formats
- Edge cases (null, empty, malformed paths)
- Case sensitivity (MKV vs mkv)
- Resolution variations (4K, 2160p, 1080P)

**Test Coverage Goal**: ≥ 95%

**Benchmark Tests**: Optional for performance validation
