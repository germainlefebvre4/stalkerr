# api-processing-logs Specification

## Purpose
Exposes a REST endpoint for retrieving paginated background processing run logs, including per-run statistics, so operators can monitor and filter past processing runs.
## Requirements
### Requirement: List Processing Logs
The system SHALL expose a REST endpoint `GET /api/v1/processing-logs` to retrieve paginated background processing log entries from the database. Each returned entry SHALL include the per-run statistics defined by capability `processing-run-statistics` (movies count, TV shows count, new-items count, TMDB matched count, TMDB unmatched count, and the distinct `group_title` list), represented as absent/null for entries that predate that capability rather than as zero or an empty list.

#### Scenario: Retrieve paginated list of processing logs
- **WHEN** a client makes a `GET` request to `/api/v1/processing-logs` with parameters `limit=10` and `offset=0`
- **THEN** the system SHALL query the `processing_logs` table and return a `200 OK` JSON response containing the list of logs, the total count, and pagination metadata.

#### Scenario: Filter processing logs by status
- **WHEN** a client makes a `GET` request to `/api/v1/processing-logs` with a status query parameter of `in_progress`
- **THEN** the system SHALL return only those logs that have a status of `in_progress`.

#### Scenario: A completed run's entry includes its per-run statistics
- **WHEN** a client requests `GET /api/v1/processing-logs` and the response includes an entry for a completed run that processed 12 movies, 5 TV shows, created 8 new items, matched 17 items to TMDB, failed to match 3, and touched the group titles "ACTION-FR" and "ANIMATION"
- **THEN** that entry SHALL include a movies count of `12`, a TV shows count of `5`, a new-items count of `8`, a TMDB matched count of `17`, a TMDB unmatched count of `3`, and a group title list of `["ACTION-FR", "ANIMATION"]`

#### Scenario: A pre-migration entry omits per-run statistics
- **WHEN** a client requests `GET /api/v1/processing-logs` and the response includes an entry created before per-run statistics were introduced
- **THEN** that entry's statistics fields SHALL be `null`/absent rather than `0` or an empty list

