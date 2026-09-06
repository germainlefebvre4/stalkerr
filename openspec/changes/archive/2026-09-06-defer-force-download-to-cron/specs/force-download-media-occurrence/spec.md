## MODIFIED Requirements

### Requirement: Forced download executes asynchronously
Once eligibility, the existence check, and the monitored check are confirmed, the system SHALL accept the forced-download request immediately without starting the file transfer in the request path. The actual file transfer SHALL be performed by the next scheduled `download` run, which picks up the request through the same incomplete-download resume mechanism used for any interrupted transfer, using the same download-tracking record and state transitions as the existing automatic pipeline, so its progress and completion remain observable through the existing downloads listing.

#### Scenario: Request is accepted before the file transfer completes
- **WHEN** a forced-download request passes eligibility, the existence check, and the monitored check
- **THEN** the system SHALL acknowledge acceptance of the request and persist a pending download record without waiting for the file transfer to finish, and SHALL NOT itself start that transfer

#### Scenario: File transfer happens on the next scheduled download run
- **WHEN** a forced-download request has been accepted
- **THEN** the file transfer for that occurrence SHALL only begin when the next scheduled `download` run picks it up as an incomplete/pending download, not before

#### Scenario: Forced download progress is visible through existing tracking
- **WHEN** an accepted forced download is pending, in progress, or has completed
- **THEN** its status SHALL be retrievable through the same download listing used for automatically-triggered downloads

## ADDED Requirements

### Requirement: Forced download requires the target to be monitored
In addition to confirming existence, the live check performed when a forced-download request is submitted SHALL confirm that the associated movie (or, for a TV episode, its series) is marked monitored in the connected Radarr/Sonarr instance. When it is confirmed not monitored, the system SHALL refuse the request, report that the media is not monitored, and SHALL NOT create or update any download record. This check exists because a request accepted for unmonitored media would otherwise be silently skipped by the deferred pickup mechanism, which never resumes an incomplete download confirmed unmonitored.

#### Scenario: Movie confirmed monitored
- **WHEN** a forced-download request targets a movie occurrence and Radarr confirms the movie exists and is monitored
- **THEN** the system SHALL proceed to persist the pending download record

#### Scenario: Movie confirmed unmonitored
- **WHEN** a forced-download request targets a movie occurrence and Radarr confirms the movie exists but is not monitored
- **THEN** the system SHALL refuse the request, reporting that the movie is not monitored, and SHALL NOT create or update any download record

#### Scenario: Episode's series confirmed unmonitored
- **WHEN** a forced-download request targets a TV episode occurrence and Sonarr confirms the episode exists but its series is not monitored
- **THEN** the system SHALL refuse the request, reporting that the series is not monitored, and SHALL NOT create or update any download record
