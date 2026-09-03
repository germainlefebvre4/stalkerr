# download-transfer-integrity Specification

## Purpose

Guarantees that a media download is only ever reported as successfully completed when the transferred file meets a minimum size, so a dead or expired source link (which can respond `200 OK` with an empty or near-empty body) is retried and then explicitly failed instead of silently being recorded as a successful, unusable file.

## Requirements

### Requirement: Minimum downloaded size for a successful transfer
A download transfer SHALL only be eligible to be marked `completed` if the number of bytes written to disk during that attempt is greater than or equal to a configured minimum file size. The minimum SHALL be configurable, with a default of 1 MB.

#### Scenario: Transfer at or above the minimum completes normally
- **WHEN** a download transfer finishes writing a file whose size is greater than or equal to the configured minimum
- **THEN** the download SHALL proceed through the existing completion path (file moved to its final destination, status set to `completed`)

#### Scenario: Transfer below the minimum is not marked completed
- **WHEN** a download transfer finishes (with no transport-level error) writing fewer bytes than the configured minimum
- **THEN** the download SHALL NOT be marked `completed`, and its file SHALL NOT be moved to the final destination path

### Requirement: Undersized transfers are retried before failing
An undersized transfer SHALL be treated as a retryable failure, sharing the same retry attempt budget already used for network-level failures on that download, rather than a separate, additional counter. Each retry SHALL restart the transfer from the beginning, not resume from the undersized attempt's partial bytes.

#### Scenario: Undersized attempt is retried
- **WHEN** a download transfer attempt finishes below the configured minimum size and retry attempts remain in the shared budget
- **THEN** the system SHALL start a new transfer attempt for the same download, beginning at byte 0

#### Scenario: Retry budget shared with network failures
- **WHEN** a download has already used one or more retry attempts due to network-level errors
- **THEN** an undersized-transfer failure on a subsequent attempt SHALL count against that same remaining budget, not a separate allowance

### Requirement: Exhausted retries fail explicitly with a diagnostic reason
When every attempt within the retry budget ends with the transferred file below the configured minimum size, the download SHALL be marked `failed` with an error message that explicitly states the file was empty or undersized and reports how many bytes were written on the last attempt, distinguishing this failure from a network or HTTP-status failure.

#### Scenario: All attempts undersized
- **WHEN** a download exhausts its retry budget and every attempt's transferred size stayed below the configured minimum
- **THEN** the download's status SHALL be set to `failed` and its error message SHALL state that the downloaded file was empty or undersized, including the last attempt's byte count

### Requirement: Guard applies uniformly to every path that performs a transfer
The minimum-size guard, its shared retry behavior, and its explicit failure reason SHALL apply identically regardless of which caller initiated the download (an on-demand forced download or an automatic resume of an incomplete download), since both rely on the same underlying transfer mechanism.

#### Scenario: Forced download hits a dead link
- **WHEN** a user-triggered forced download's source responds successfully but with a body smaller than the configured minimum, on every retry attempt
- **THEN** the item SHALL end in the `failed` status with the undersized-file error message, not `completed`

#### Scenario: Resumed download hits a dead link
- **WHEN** an automatic resume of a previously incomplete download's source responds successfully but with a body smaller than the configured minimum, on every retry attempt
- **THEN** the item SHALL end in the `failed` status with the undersized-file error message, not `completed`
