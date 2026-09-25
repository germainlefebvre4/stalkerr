# Spec Delta

## MODIFIED Requirements

### Requirement: Detected-Problem Filter Values
The `problem` query parameter on `GET /api/v1/downloads` SHALL support the values `missing_year` (`file_info.has_year_in_path` is false), `year_mismatch` (`file_info.year_mismatch` is true), `unknown_format` (`file_info.is_valid_format` is false), `low_quality` (detected resolution is `480p` or `360p`), and `duplicate_download` (the download's movie/episode meets the Duplicate Group Definition, capability `duplicate-download-detection`), each restricting the result set to downloads matching that condition.

#### Scenario: Filtering by missing_year
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year`
- **THEN** the response SHALL include only downloads where `file_info.has_year_in_path` is false

#### Scenario: Filtering by low_quality
- **WHEN** a client requests `GET /api/v1/downloads?problem=low_quality`
- **THEN** the response SHALL include only downloads whose detected resolution is `480p` or `360p`

#### Scenario: Filtering by duplicate_download
- **WHEN** a client requests `GET /api/v1/downloads?problem=duplicate_download`
- **THEN** the response SHALL include only downloads that are part of a duplicate group per the Duplicate Group Definition

## ADDED Requirements

### Requirement: Duplicate Group Detail Field
`GET /api/v1/downloads` SHALL include, for every download that is part of a duplicate group, a field identifying its sibling occurrence(s) in that group (at minimum each sibling's `download_info` id, `download_path`, and `completed_at`). This field SHALL be omitted for a download that is not part of a duplicate group.

#### Scenario: Duplicate response includes sibling detail
- **WHEN** a listed download is part of a duplicate group with one sibling occurrence
- **THEN** its response entry SHALL include that sibling's `download_info` id, `download_path`, and `completed_at`

#### Scenario: Non-duplicate response omits the field
- **WHEN** a listed download is not part of any duplicate group
- **THEN** its response entry SHALL omit the duplicate-group detail field
