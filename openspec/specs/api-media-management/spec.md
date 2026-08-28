# api-media-management Specification

## Purpose
TBD - created by archiving change frontend-ihm-react-radix. Update Purpose after archive.
## Requirements
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

### Requirement: Rename Single Download Parent Folder
The system SHALL expose a REST endpoint `POST /api/v1/downloads/:id/rename` that renames the parent folder of exactly one downloaded item (a movie's file, or a single TV episode's file), identified by its `download_info` id, without moving, renaming, or otherwise modifying any other download that shares the same original parent directory. The request payload SHALL be `{"new_name": string, "destination_parent_dir"?: string}`. When `destination_parent_dir` is omitted, the item SHALL stay under its current library root; when provided, the item SHALL be moved under that root as part of the same operation. On success, the system SHALL update the `download_path` column of the targeted `download_info` record only. If the computed destination path already exists, the system SHALL abort without touching any file and return a distinct machine-readable error code of `"rename_target_exists"`. If the rename operation fails for any other reason, the response SHALL include a distinct machine-readable error code of `"rename_failed"`.

#### Scenario: Rename a movie's parent folder
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/42/rename` with `{"new_name": "Correct Movie Title (2020)"}`, where download 42 is a completed movie download
- **THEN** the system SHALL move the movie's file into `{current library root}/Correct Movie Title (2020)/`, update `download_info.id=42`'s `download_path` accordingly, and leave every other `download_info` record untouched.

#### Scenario: Rename one TV episode's parent folder without affecting sibling episodes
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/77/rename` with `{"new_name": "Corrected Series Name (2021)"}`, where download 77 is a completed episode file stored at `.../Series Name (2020)/Season 01/Series Name (2020) - S01E01.mkv`, and other episodes of the same series (e.g. S01E02, S02E01) are downloaded and stored under `.../Series Name (2020)/...`
- **THEN** the system SHALL move only download 77's file to `.../Corrected Series Name (2021)/Season 01/Corrected Series Name (2021) - S01E01.mkv`, update only `download_info.id=77`'s `download_path`, and leave the files and `download_path` of every other episode's `download_info` record unchanged in the original `Series Name (2020)` directory.

#### Scenario: Rename combined with a different destination root
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/77/rename` with `{"new_name": "Corrected Series Name (2021)", "destination_parent_dir": "/media/tv-archive"}`
- **THEN** the system SHALL move download 77's file to `/media/tv-archive/Corrected Series Name (2021)/Season 01/Corrected Series Name (2021) - S01E01.mkv` and update only its `download_path`, without moving any other episode.

#### Scenario: Original directory cleaned up when left empty
- **WHEN** the renamed item was the only file remaining in its original season directory and/or series directory
- **THEN** the system SHALL delete the now-empty season directory and, if it too becomes empty, the now-empty series directory, without affecting any directory that still contains other downloads.

#### Scenario: Rename blocked by an existing destination
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/:id/rename` and the computed destination file path already exists on disk
- **THEN** the system SHALL return an error response with `ErrorResponse.error` set to `"rename_target_exists"`, and SHALL NOT move, delete, or modify any file or `download_info` record.

#### Scenario: Rename attempted on an incomplete download
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/:id/rename` for a `download_info` record that has no `download_path` (not yet completed)
- **THEN** the system SHALL return an error response and SHALL NOT modify any file or database record.

### Requirement: Retrieve Configured Root Paths
The system SHALL expose a REST endpoint `GET /api/v1/config/paths` to retrieve the default paths configured for movies and TV shows.

#### Scenario: Get configured paths
- **WHEN** a client makes a `GET` request to `/api/v1/config/paths`
- **THEN** the system SHALL return a `200 OK` JSON response containing the values of `downloads.movies_path` and `downloads.tvshows_path`.

