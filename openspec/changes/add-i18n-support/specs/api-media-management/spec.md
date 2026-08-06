## MODIFIED Requirements

### Requirement: Move Media Parent Folder
The system SHALL expose REST endpoints `POST /api/v1/movies/:id/move` and `POST /api/v1/tvshows/:id/move` to safely move the complete directory of a movie or TV show to a new parent directory and update the database records. If the move fails, the response SHALL include a distinct machine-readable error code of `"move_failed"` (rather than a generic error code) so that clients can render a specific, translatable error message.

#### Scenario: Move entire movie parent directory
- **WHEN** a client makes a `POST` request to `/api/v1/movies/12/move` with a JSON payload of `{"destination_parent_dir": "/media/children-movies"}`
- **THEN** the system SHALL locate the movie on disk, move its entire containing folder to `/media/children-movies`, and update the `download_path` columns of all associated `download_info` records to reflect their new paths.

#### Scenario: Move entire TV show parent directory
- **WHEN** a client makes a `POST` request to `/api/v1/tvshows/5/move` with a JSON payload of `{"destination_parent_dir": "/media/children-tv"}`
- **THEN** the system SHALL locate the TV show on disk, move its entire containing directory (including all sub-directories like seasons) to `/media/children-tv`, and update all associated `download_info` records to reflect their new paths.

#### Scenario: Move fails partway through
- **WHEN** a client makes a `POST` request to `/api/v1/movies/12/move` and the recursive copy or the destination verification fails
- **THEN** the system SHALL return a `500 Internal Server Error` response with `ErrorResponse.error` set to `"move_failed"`
