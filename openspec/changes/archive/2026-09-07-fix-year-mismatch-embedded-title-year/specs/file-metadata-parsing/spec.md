## ADDED Requirements

### Requirement: Parenthesized Year Takes Precedence Over Embedded Title Numbers
When a download path contains one or more years enclosed in parentheses
(`(YYYY)`, this app's own `Title (Year)` naming convention), the parser
SHALL treat the last such parenthesized year in the path as `detected_year`,
even when an earlier, non-parenthesized 4-digit number elsewhere in the
path (e.g. a year-like number embedded in the title itself) would otherwise
match first.

#### Scenario: Title contains an embedded year-like number, correct year is parenthesized
- **WHEN** `Parse()` is called with path `/media/movies/Valensole 1965 (2025)/Valensole 1965 (2025).mkv` and `tmdbYear` 2025
- **THEN** `detected_year` SHALL be 2025, `has_year_in_path` SHALL be true, and `year_mismatch` SHALL be false

#### Scenario: No parenthesized year present, existing bare-match behavior is preserved
- **WHEN** `Parse()` is called with a path containing no parenthesized year (e.g. `/media/movies/2001.A.Space.Odyssey.1968.720p/movie.mkv`) and `tmdbYear` 1968
- **THEN** `detected_year` SHALL be 2001 (the first bare `\b(19|20)\d{2}\b` match), unchanged from prior behavior

#### Scenario: Multiple parenthesized years present
- **WHEN** `Parse()` is called with a path containing more than one parenthesized year (e.g. `Title (1984) (2021)`)
- **THEN** `detected_year` SHALL be the last parenthesized year found in the path
