# Cancel Download Occurrence Specification

## Purpose

Lets a user immediately stop one specific stuck download occurrence — a
`ProcessedLine`/`DownloadInfo` pair that would otherwise keep being retried
automatically forever — without waiting for its retry budget to exhaust on
its own, and without affecting any sibling occurrence of the same movie or
TV episode.

## Requirements

### Requirement: Cancel targets exactly one occurrence
The system SHALL allow cancelling exactly one specific download occurrence, identified individually by its `ProcessedLine` id. This SHALL NOT affect, alter, or exclude any sibling occurrence of the same movie or TV episode, regardless of those siblings' own state.

#### Scenario: Cancelling one occurrence leaves siblings untouched
- **WHEN** a user cancels one occurrence of a movie that has another untried, in-progress, or completed sibling occurrence
- **THEN** the system SHALL cancel only the requested occurrence, leaving every sibling's own state and future eligibility unchanged

### Requirement: Cancel is refused for ineligible occurrences
The system SHALL refuse a cancel request, returning a clear error and performing no side effect, when the target occurrence's current status is `completed`, is `downloading` (an actively in-progress transfer — this capability does not introduce a mechanism to interrupt a live transfer), or is already in the terminal cancelled/retry-exhausted status. Cancel SHALL be accepted for an occurrence whose status is `pending`, `failed`, or `retrying`.

#### Scenario: Cancel refused for a completed occurrence
- **WHEN** a cancel request targets an occurrence whose status is `completed`
- **THEN** the system SHALL refuse the request without changing any state

#### Scenario: Cancel refused for an actively downloading occurrence
- **WHEN** a cancel request targets an occurrence whose status is `downloading`
- **THEN** the system SHALL refuse the request without interrupting the in-progress transfer or changing any state

#### Scenario: Cancel refused when already cancelled or retry-exhausted
- **WHEN** a cancel request targets an occurrence that is already in the terminal cancelled/retry-exhausted status
- **THEN** the system SHALL refuse the request, reporting that it is already in that status

#### Scenario: Cancel accepted for a failed occurrence
- **WHEN** a cancel request targets an occurrence whose status is `failed`
- **THEN** the system SHALL accept the request and cancel the occurrence

#### Scenario: Cancel accepted for a pending or retrying occurrence
- **WHEN** a cancel request targets an occurrence whose status is `pending` or `retrying`
- **THEN** the system SHALL accept the request and cancel the occurrence

### Requirement: Cancelling excludes the occurrence from all future automatic attempts
Once cancelled, the occurrence SHALL reach the same terminal excluded status a retry-exhausted occurrence reaches (capability `media-download-scheduling`): it SHALL NOT be offered again by the missing-content matcher and SHALL NOT be resumed via the incomplete-download resume path.

#### Scenario: Cancelled occurrence is never retried automatically
- **WHEN** an occurrence has been cancelled and its movie/series is still missing/monitored in Radarr/Sonarr
- **THEN** the next scheduled run SHALL NOT attempt to download that cancelled occurrence

### Requirement: Cancelling performs no filesystem cleanup
Cancel SHALL only update database state; it SHALL NOT attempt to delete, move, or otherwise touch any file on disk. A failed or interrupted transfer's temporary file has already been removed by the existing download cleanup path, and a cancelled occurrence never reaches a state where a final destination file exists.

#### Scenario: No filesystem action on cancel
- **WHEN** a user cancels a `failed` occurrence
- **THEN** no file deletion or move is attempted as part of the cancel action

### Requirement: Cancelling is reversible only through the existing full media reset
The system SHALL NOT provide a targeted "reactivate" action for a cancelled or retry-exhausted occurrence. The only way to make that exact occurrence eligible again SHALL be the existing full reset of its movie or TV show (capability `media-reset`), which clears the entire download history for that movie/show.

#### Scenario: Reversing a cancellation requires a full reset
- **WHEN** a user wants to retry a cancelled occurrence
- **THEN** they SHALL use the existing movie/TV show reset endpoint, which clears all processed lines for that movie/show, rather than a targeted reactivate action

### Requirement: REST API endpoint to cancel one occurrence
The API server SHALL expose a `POST /api/v1/downloads/:id/cancel` endpoint (id referring to the `DownloadInfo` id) that applies the cancel eligibility rules synchronously and reports the outcome in the response — unlike force-download, cancel is not an asynchronous operation, since it only changes database state.

#### Scenario: Successful cancel
- **WHEN** a client sends `POST /api/v1/downloads/:id/cancel` for an eligible occurrence
- **THEN** the server SHALL return `200 OK` with the occurrence's updated status

#### Scenario: Cancel of a non-existent id
- **WHEN** a client sends `POST /api/v1/downloads/:id/cancel` with an id that does not exist
- **THEN** the server SHALL return `404 Not Found`

#### Scenario: Cancel of an ineligible occurrence
- **WHEN** a client sends `POST /api/v1/downloads/:id/cancel` for an occurrence whose status is `completed`, `downloading`, or already cancelled/retry-exhausted
- **THEN** the server SHALL return `409 Conflict` with a clear reason, and SHALL NOT change any state
