## ADDED Requirements

### Requirement: Enriched Downloads Listing
`GET /api/v1/downloads` SHALL return paginated download records enriched with content metadata (from the associated Movie or TVShow via the download's first ProcessedLine) and parsed file metadata (via the fileparser package), replacing the plain download listing.

#### Scenario: Enriched listing includes content and file metadata
- **WHEN** a client requests `GET /api/v1/downloads`
- **THEN** each returned download SHALL include a `content` object built from its first ProcessedLine's associated Movie or TVShow (ordered by `created_at` ascending) when one exists, and a `file_info` object parsed from its `download_path` when non-null

### Requirement: Status and Type Filtering
`GET /api/v1/downloads` SHALL support `status` and `type` query parameters that restrict the result set to downloads matching the given status or content type respectively.

#### Scenario: Filtering by status
- **WHEN** a client requests `GET /api/v1/downloads?status=completed`
- **THEN** the response SHALL include only downloads whose status is `completed`

#### Scenario: Filtering by content type
- **WHEN** a client requests `GET /api/v1/downloads?type=movies`
- **THEN** the response SHALL include only downloads whose content type is `movies`

### Requirement: Detected-Problem Filter Values
The `problem` query parameter on `GET /api/v1/downloads` SHALL support the values `missing_year` (`file_info.has_year_in_path` is false), `year_mismatch` (`file_info.year_mismatch` is true), `unknown_format` (`file_info.is_valid_format` is false), and `low_quality` (detected resolution is `480p` or `360p`), each restricting the result set to downloads matching that condition.

#### Scenario: Filtering by missing_year
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year`
- **THEN** the response SHALL include only downloads where `file_info.has_year_in_path` is false

#### Scenario: Filtering by low_quality
- **WHEN** a client requests `GET /api/v1/downloads?problem=low_quality`
- **THEN** the response SHALL include only downloads whose detected resolution is `480p` or `360p`

### Requirement: Content Metadata Field Mapping
When building `content` from a Movie association, the system SHALL set `type` to `movies` and populate title, year, genres, duration, and resolution from the movie and processed line. When building `content` from a TVShow association, the system SHALL set `type` to `tvshows`, format the title as `{title} S{season:02d}E{episode:02d}` when season/episode are present, and populate year, genres, season, and episode.

#### Scenario: Movie content metadata
- **WHEN** a download's first ProcessedLine has a Movie association
- **THEN** `content.type` SHALL be `movies` and SHALL include title, year, genres, duration, and resolution from that movie/processed line

#### Scenario: TV show content metadata includes formatted episode title
- **WHEN** a download's first ProcessedLine has a TVShow association with season 5 and episode 14
- **THEN** `content.type` SHALL be `tvshows` and `content.title` SHALL be formatted as `{tmdb_title} S05E14`

### Requirement: Missing Association Handling
When a download has no ProcessedLines, `content` SHALL be nil. When a download has a null `download_path`, `file_info` SHALL be nil.

#### Scenario: Orphan download has no content
- **WHEN** a download has no ProcessedLines
- **THEN** `content` SHALL be nil in the response

#### Scenario: Download with no path has no file_info
- **WHEN** a download's `download_path` is null
- **THEN** `file_info` SHALL be nil in the response

### Requirement: Database Error Response
When a database error occurs while serving `GET /api/v1/downloads`, the system SHALL respond with HTTP 500 and a JSON body identifying the error (`error`/`message` fields).

#### Scenario: Database failure returns 500 with error body
- **WHEN** the database query for `GET /api/v1/downloads` fails
- **THEN** the response SHALL have status 500 and a JSON body of the form `{"error": "database_error", "message": "..."}`

### Requirement: Response Time Budget
`GET /api/v1/downloads` SHALL respond in under 500ms for a page of 20 fully-enriched results under normal load.

#### Scenario: Standard page responds within budget
- **WHEN** a client requests a page of 20 results with full enrichment under normal load
- **THEN** the response SHALL be returned in under 500ms
