## ADDED Requirements

### Requirement: Problem Filter Pagination Accuracy
When the `problem` query parameter is set on `GET /api/v1/downloads`, the system SHALL apply the problem filter to the full status/type-matched result set before computing `total` and `total_pages` and before slicing the requested `limit`/`offset` page, so that pagination metadata and the returned page reflect the actual filtered result count. This supersedes the previously documented behavior of reporting the pre-filter count as `total` when `problem` is set.

#### Scenario: Total reflects the problem-filtered count
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year&limit=20&offset=0` and, of the downloads matching the active `status`/`type` filters, 45 have `!file_info.has_year_in_path`
- **THEN** the response SHALL report `total: 45`, `total_pages` computed from 45 and the given `limit`, and `data` containing the first 20 of those 45 filtered downloads

#### Scenario: Later pages stay consistent with the filtered total
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year&limit=20&offset=20` under the same conditions as the previous scenario
- **THEN** the response SHALL return items 21–40 of the problem-filtered set (not of the pre-filter set), consistent with the `total` reported for `offset=0`
