## ADDED Requirements

### Requirement: Resync Single Download Path with Radarr/Sonarr
The system SHALL expose a REST endpoint `POST /api/v1/downloads/:id/resync-path` that performs an on-demand correction (capability `sonarr-series-path-routing`, Requirement: On-Demand Correction of a Single Completed Download's Path) of exactly one completed download's stored `download_path`, identified by its `download_info` id. The endpoint SHALL respond with a machine-readable outcome distinguishing three cases: the path was corrected (including the old and new path), the path already matched Radarr/Sonarr's current data (no change made), and the download's movie/series is not currently found in the user's Radarr/Sonarr library (`"not_managed_by_radarr_sonarr"`, no change made). The endpoint SHALL only ever modify the targeted download's own `download_path`; no other `download_info` record shall be affected.

#### Scenario: Path corrected
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/42/resync-path`, where download 42's stored root no longer matches its movie's/series' current Radarr/Sonarr `Path`
- **THEN** the system SHALL move the file to the corrected path, update `download_info.id=42`'s `download_path`, and return a `200 OK` response indicating the path was corrected along with the old and new values

#### Scenario: Path already up to date
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/:id/resync-path` for a download whose stored root already matches its movie's/series' current Radarr/Sonarr `Path`
- **THEN** the system SHALL make no file or database changes and return a `200 OK` response indicating no correction was needed

#### Scenario: Download not managed by Radarr/Sonarr
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/:id/resync-path` for a download whose movie/series is not found in the user's current Radarr/Sonarr library
- **THEN** the system SHALL make no file or database changes and return a response with a machine-readable code of `"not_managed_by_radarr_sonarr"`

#### Scenario: Resync attempted on an incomplete download
- **WHEN** a client makes a `POST` request to `/api/v1/downloads/:id/resync-path` for a `download_info` record that has no `download_path` (not yet completed)
- **THEN** the system SHALL return an error response and SHALL NOT modify any file or database record

#### Scenario: Resync blocked by an existing destination
- **WHEN** the corrected destination file path already exists on disk
- **THEN** the system SHALL abort without touching any file or database record and return a distinct machine-readable error code of `"rename_target_exists"`
