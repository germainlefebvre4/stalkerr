## ADDED Requirements

### Requirement: List Downloads
The system SHALL expose a REST endpoint `GET /api/v1/downloads` to retrieve paginated download tracking entries from the database.

#### Scenario: Retrieve paginated list of downloads
- **WHEN** a client makes a `GET` request to `/api/v1/downloads` with parameters `limit=20` and `offset=0`
- **THEN** the system SHALL query the `download_info` table and return a `200 OK` JSON response containing the list of download records (including file size, bytes downloaded, status, and download path) and pagination metadata.

#### Scenario: Filter downloads by status
- **WHEN** a client makes a `GET` request to `/api/v1/downloads` with a status query parameter of `downloading`
- **THEN** the system SHALL return only those downloads that have a status of `downloading`.
