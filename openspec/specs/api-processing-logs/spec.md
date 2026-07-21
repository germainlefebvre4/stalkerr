# api-processing-logs Specification

## Purpose
TBD - created by archiving change frontend-ihm-react-radix. Update Purpose after archive.
## Requirements
### Requirement: List Processing Logs
The system SHALL expose a REST endpoint `GET /api/v1/processing-logs` to retrieve paginated background processing log entries from the database.

#### Scenario: Retrieve paginated list of processing logs
- **WHEN** a client makes a `GET` request to `/api/v1/processing-logs` with parameters `limit=10` and `offset=0`
- **THEN** the system SHALL query the `processing_logs` table and return a `200 OK` JSON response containing the list of logs, the total count, and pagination metadata.

#### Scenario: Filter processing logs by status
- **WHEN** a client makes a `GET` request to `/api/v1/processing-logs` with a status query parameter of `in_progress`
- **THEN** the system SHALL return only those logs that have a status of `in_progress`.

