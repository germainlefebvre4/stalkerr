# Spec Delta

## ADDED Requirements

### Requirement: Forced Origin Is Recorded
When a forced-download request is accepted and its `DownloadInfo` record is created or reused, the system SHALL persist an explicit, durable indicator on that record that it originated from a forced download, distinct from the automatic pipeline. This indicator SHALL remain in place for the lifetime of the record and SHALL be readable by other capabilities (e.g. `duplicate-download-detection`) without inferring origin from the destination filename or any other indirect signal.

#### Scenario: A newly created forced download is marked
- **WHEN** a forced-download request is accepted for an occurrence with no existing `DownloadInfo`
- **THEN** the newly created `DownloadInfo` record SHALL carry the forced-origin indicator

#### Scenario: A reused DownloadInfo record is marked
- **WHEN** a forced-download request is accepted for an occurrence that already has a `DownloadInfo` record (e.g. resuming after a server restart)
- **THEN** that existing record SHALL carry the forced-origin indicator after the request is accepted

#### Scenario: An automatic-pipeline download is not marked
- **WHEN** a `DownloadInfo` record is created by the automatic tier-1, tier-2, or incomplete-download resume path
- **THEN** it SHALL NOT carry the forced-origin indicator
