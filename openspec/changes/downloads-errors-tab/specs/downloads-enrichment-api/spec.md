## ADDED Requirements

### Requirement: Multi-Value Problem Filter
The `problem` query parameter on `GET /api/v1/downloads` SHALL accept a comma-separated list of two or more of the existing problem values (`missing_year`, `year_mismatch`, `unknown_format`, `low_quality`), in addition to continuing to accept a single value as before. When multiple values are given, the system SHALL include a download in the filtered result if it matches at least one of the listed values (logical OR across the listed problems).

#### Scenario: Combined filter returns the union of matching downloads
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year,unknown_format`
- **THEN** the response SHALL include every download where `!file_info.has_year_in_path` OR `!file_info.is_valid_format`, without requiring both conditions to hold

#### Scenario: Single value keeps its existing exact-match behavior
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year`
- **THEN** the response SHALL behave exactly as before this change, including only downloads where `!file_info.has_year_in_path`

## MODIFIED Requirements

### Requirement: Problem Filter Pagination Accuracy
When the `problem` query parameter is set on `GET /api/v1/downloads` — whether to a single value or a comma-separated list of values — the system SHALL apply the problem filter (matching any listed value) to the full status/type-matched result set before computing `total` and `total_pages` and before slicing the requested `limit`/`offset` page, so that pagination metadata and the returned page reflect the actual filtered result count. This supersedes the previously documented behavior of reporting the pre-filter count as `total` when `problem` is set.

#### Scenario: Total reflects the problem-filtered count
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year&limit=20&offset=0` and, of the downloads matching the active `status`/`type` filters, 45 have `!file_info.has_year_in_path`
- **THEN** the response SHALL report `total: 45`, `total_pages` computed from 45 and the given `limit`, and `data` containing the first 20 of those 45 filtered downloads

#### Scenario: Later pages stay consistent with the filtered total
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year&limit=20&offset=20` under the same conditions as the previous scenario
- **THEN** the response SHALL return items 21-40 of the problem-filtered set (not of the pre-filter set), consistent with the `total` reported for `offset=0`

#### Scenario: Total reflects the combined-filter count for a multi-value request
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year,unknown_format&limit=20&offset=0` and, of the downloads matching the active `status`/`type` filters, 60 have `!file_info.has_year_in_path` or `!file_info.is_valid_format`
- **THEN** the response SHALL report `total: 60`, `total_pages` computed from 60 and the given `limit`, and `data` containing the first 20 of those 60 filtered downloads
