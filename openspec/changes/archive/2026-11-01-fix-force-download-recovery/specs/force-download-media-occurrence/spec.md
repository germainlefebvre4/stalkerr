# Spec Delta

## ADDED Requirements

### Requirement: Re-forcing a cancelled occurrence reactivates it
When a forced-download request targets an occurrence whose download record or own state is `cancelled` and that is otherwise eligible, the system SHALL reactivate it rather than accept the request as a no-op. Reactivation SHALL return the download record to `pending` with its retry count reset and any recorded error and lock cleared, and SHALL return the occurrence to a state eligible for transfer.

#### Scenario: Re-forcing an occurrence cancelled by the scheduler
- **WHEN** a forced-download request targets an occurrence whose download record is `cancelled` because another occurrence of the same movie was already downloaded
- **THEN** the system SHALL set that record back to `pending` with retry count 0 and no error, SHALL set the occurrence back to its pre-download state, and SHALL acknowledge the request

#### Scenario: Re-forcing an occurrence cancelled by the user
- **WHEN** a forced-download request targets an occurrence the user previously cancelled
- **THEN** the system SHALL reactivate it exactly as for a scheduler-cancelled occurrence, since the explicit force expresses the user's new intent

#### Scenario: Reactivated download is visible in the active downloads listing
- **WHEN** a cancelled occurrence has been reactivated by a forced-download request
- **THEN** its download record SHALL appear in the downloads listing's default (active) view as `pending`

#### Scenario: Re-forcing an already-downloaded occurrence is still refused
- **WHEN** a forced-download request targets an occurrence whose own state is `downloaded`
- **THEN** the system SHALL refuse the request and SHALL NOT reactivate or re-create anything

### Requirement: Forced transfer attempts only the forced occurrence at its recorded destination
When the scheduled `download` run picks up an accepted forced download, it SHALL attempt only the occurrence that was forced, and SHALL NOT substitute or precede it with any sibling occurrence, regardless of the quality ranking the automatic pipeline applies. The file SHALL be written to the destination recorded on the download record when the request was accepted, so the resolution, language and variant suffixes persisted at that time are preserved.

#### Scenario: Forced 1080p is attempted even though a 720p sibling ranks higher
- **WHEN** a movie has a pending forced 1080p occurrence and an untried 720p sibling that the automatic ranking prefers
- **THEN** the run SHALL attempt the forced 1080p occurrence and SHALL NOT download the 720p sibling on its behalf

#### Scenario: Forced file keeps its recorded suffixes
- **WHEN** a forced occurrence recorded a destination ending in `[1080p][MULTI]` and its transfer completes
- **THEN** the file SHALL be written under that recorded destination

#### Scenario: Failure of the forced occurrence does not fall through to siblings
- **WHEN** the transfer of a forced occurrence fails
- **THEN** the run SHALL NOT attempt any other occurrence of that movie or episode as a replacement within the same forced item

### Requirement: One forced occurrence per media item per run
When several occurrences of the same movie or TV episode have pending forced downloads, the scheduled `download` run SHALL process only one of them, the one whose download record is oldest, and SHALL leave the others pending for subsequent runs.

#### Scenario: Two forced occurrences of one movie
- **WHEN** two occurrences of the same movie each have a pending forced download
- **THEN** the run SHALL transfer only the one with the oldest download record, and the other SHALL remain `pending` and unchanged

#### Scenario: Remaining forced occurrence is processed on the next run
- **WHEN** the first forced occurrence has completed and the second is still `pending`
- **THEN** the next scheduled run SHALL process the second one
