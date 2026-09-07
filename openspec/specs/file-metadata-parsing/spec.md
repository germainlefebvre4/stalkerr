# File Metadata Parsing

## Description

A pure Go package that extracts technical file metadata from download paths using regex pattern matching. Provides structured information about video files including extension, resolution, year detection, and format validation.

## Requirements

### Functional Requirements

**GIVEN** a download path string and optional TMDB year  
**WHEN** Parse() is called  
**THEN** the system SHALL return a FileInfo struct containing:
- File extension (lowercase, with dot: `.mkv`, `.mp4`, etc.)
- Folder name (last directory component)
- File name (basename)
- Year detection flag (`has_year_in_path`)
- Detected year (4-digit number from path, if present)
- Year mismatch flag (detected year != TMDB year)
- Detected resolution (1080p, 720p, 4K, etc., if present)
- Format validation flag (`is_valid_format` based on known video extensions)

**GIVEN** a null or empty download path  
**WHEN** Parse() is called  
**THEN** the system SHALL return nil

**GIVEN** a path containing one or more years enclosed in parentheses (`(YYYY)`, this app's own `Title (Year)` naming convention)  
**WHEN** Parse() is called  
**THEN** the system SHALL treat the last such parenthesized year in the path as `detected_year`, even when an earlier, non-parenthesized 4-digit number elsewhere in the path (e.g. a year-like number embedded in the title itself) would otherwise match first

**GIVEN** a path containing multiple 4-digit years and no parenthesized year  
**WHEN** Parse() is called  
**THEN** the system SHALL extract the first year matching the pattern `\b(19|20)\d{2}\b`

**GIVEN** a path containing resolution indicators (2160p, 4K, 1080p, 720p, 480p, 360p)  
**WHEN** Parse() is called  
**THEN** the system SHALL extract the resolution (case-insensitive)

**GIVEN** a file with extension in the valid list (mkv, mp4, avi, mov, m4v, wmv, flv, webm)  
**WHEN** Parse() is called  
**THEN** `is_valid_format` SHALL be true

**GIVEN** a detected year and a TMDB year that differ  
**WHEN** Parse() is called  
**THEN** `year_mismatch` SHALL be true

### Non-Functional Requirements

- The parser SHALL have no external dependencies (pure stdlib)
- The parser SHALL not perform file system operations
- The parser SHALL be safe for concurrent use
- Unit test coverage SHALL be ≥ 95%

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
