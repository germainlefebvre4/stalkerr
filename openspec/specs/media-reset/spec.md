# media-reset Specification

## Purpose
TBD - created by archiving change m3u-url-reset-and-db-pruning. Update Purpose after archive.
## Requirements
### Requirement: Surgical Reset of Specific Media
The system SHALL support resetting individual movies or TV shows by deleting all `processed_lines` associated with their database ID.

#### Scenario: Successful Movie Reset
- **WHEN** a reset request for a movie is executed with a valid ID
- **THEN** the system deletes all `processed_lines` where `movie_id` matches the target movie's ID.

#### Scenario: Successful TV Show Reset
- **WHEN** a reset request for a TV show is executed with a valid ID
- **THEN** the system deletes all `processed_lines` where `tv_show_id` (or `tvshow_id`) matches the target TV show's ID.

### Requirement: REST API Endpoints for Surgical Reset
The API server SHALL expose POST endpoints to surgically reset movies and TV shows.

#### Scenario: POST /api/v1/movies/:id/reset
- **WHEN** a client sends a POST request to `/api/v1/movies/:id/reset` with a valid ID
- **THEN** the server returns `200 OK` with a JSON body indicating success, and all associated `processed_lines` are deleted.

#### Scenario: POST /api/v1/tvshows/:id/reset
- **WHEN** a client sends a POST request to `/api/v1/tvshows/:id/reset` with a valid ID
- **THEN** the server returns `200 OK` with a JSON body indicating success, and all associated `processed_lines` are deleted.

#### Scenario: Reset requesting non-existent ID
- **WHEN** a client sends a POST request to `/api/v1/movies/:id/reset` or `/api/v1/tvshows/:id/reset` with a non-existent ID
- **THEN** the server returns `404 Not Found` with a clear error message.

