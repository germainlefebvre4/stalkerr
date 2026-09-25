# duplicate-download-detection Specification

## Purpose

Identifies movies or TV episodes that the automated download pipeline has fetched more than once into different files, distinguishes that unwanted outcome from legitimate same-path quality upgrades and deliberate manual downloads, and lets an operator reconcile Stalkeer's records with what is actually on disk once the extra file has been cleaned up.

## Requirements

### Requirement: Duplicate Group Definition
A movie, or a specific TV episode, SHALL be considered to have an unexpected automated duplicate when two or more of its `ProcessedLine` occurrences have an associated `DownloadInfo` whose status is `completed`, at least two of those completed occurrences have different stored `download_path` values, and none of the occurrences being compared originated from a manually-triggered forced download (capability `force-download-media-occurrence`). Two completed occurrences that share the exact same `download_path` (a tier-2 quality upgrade that replaced the file in place) SHALL NOT, by themselves, constitute a duplicate.

#### Scenario: Two completed occurrences at different paths is a duplicate
- **WHEN** a movie has two `ProcessedLine` occurrences whose `DownloadInfo` is `completed`, neither originated from a forced download, and their `download_path` values differ
- **THEN** that movie SHALL be reported as having a duplicate

#### Scenario: A same-path tier-2 upgrade is not a duplicate
- **WHEN** a movie has two completed occurrences that were both produced by the automatic pipeline and share the exact same `download_path`
- **THEN** that movie SHALL NOT be reported as having a duplicate

#### Scenario: A forced download is excluded from the comparison
- **WHEN** a movie has one automatic completed occurrence and one completed occurrence that originated from a forced download at a different `download_path`
- **THEN** that movie SHALL NOT be reported as having a duplicate on the basis of those two occurrences alone

#### Scenario: A single completed occurrence is never a duplicate
- **WHEN** a movie or episode has at most one completed occurrence
- **THEN** it SHALL NOT be reported as having a duplicate, regardless of how many non-completed occurrences it has

### Requirement: Continuous Detection During Scheduled Reconciliation
During each run of the scheduled `download` command's existing path-reconciliation pass, the system SHALL evaluate every movie and TV episode with at least one completed download against the Duplicate Group Definition and log each detected duplicate group as a distinct, identifiable warning (surfaced the same way the existing `file_missing` outcome is today). This detection SHALL be report-only: it SHALL NOT modify any `download_info` or `processed_lines` record.

#### Scenario: A duplicate is logged during the scheduled run
- **WHEN** the scheduled `download` command's reconciliation pass runs and finds a movie meeting the Duplicate Group Definition
- **THEN** the system SHALL log a distinguishable warning identifying the movie and the duplicate occurrences, and SHALL NOT change any database record as a result

### Requirement: On-Demand Detection For The Existing Backlog
`GET /api/v1/downloads` (capability `downloads-enrichment-api`) SHALL evaluate the Duplicate Group Definition at request time, independent of the scheduled reconciliation pass, so that movies/episodes already duplicated before this capability existed are visible immediately without waiting for the next scheduled run.

#### Scenario: A pre-existing duplicate is visible without a scheduled run
- **WHEN** a movie already has two completed occurrences at different paths from before this capability was deployed, and no scheduled reconciliation run has occurred since
- **THEN** a request to `GET /api/v1/downloads` SHALL still report that movie's occurrences as duplicates

### Requirement: Manual Resolve Action
The system SHALL provide an on-demand action that reconciles one duplicate group's database state with the filesystem: for every completed occurrence in the group, the system SHALL verify whether its file still exists at its stored `download_path`. Any occurrence whose file no longer exists SHALL be transitioned to the `cancelled` status (`download_info.status` and the corresponding `processed_lines.state`), matching the terminal status the `cancel-download-occurrence` capability already uses. An occurrence whose file is still present SHALL be left unchanged. This action SHALL only ever run when explicitly triggered by an operator request; it SHALL NOT be invoked automatically by the scheduled reconciliation pass or any other background process. The action SHALL NOT delete, move, or otherwise modify any file on disk.

#### Scenario: Resolving a group where one file was manually removed
- **WHEN** an operator triggers the resolve action for a duplicate group where one occurrence's file has been deleted from disk and the other's still exists
- **THEN** the occurrence whose file is gone SHALL be transitioned to `cancelled`, and the occurrence whose file still exists SHALL be left unchanged

#### Scenario: Resolving a group where both files still exist
- **WHEN** an operator triggers the resolve action for a duplicate group where every occurrence's file is still present on disk
- **THEN** no occurrence SHALL be changed, and the group SHALL continue to be reported as a duplicate

#### Scenario: Resolve action never runs unattended
- **WHEN** the scheduled `download` command's reconciliation pass detects a duplicate group
- **THEN** it SHALL only log the finding (per the Continuous Detection requirement) and SHALL NOT invoke the resolve action itself

### Requirement: Resolve Action Endpoint
The API server SHALL expose an endpoint that, given the id of any one `download_info` occurrence belonging to a duplicate group, applies the Manual Resolve Action to that occurrence's entire group and reports, for each occurrence in the group, whether it was cancelled or left unchanged. The endpoint SHALL only ever modify occurrences belonging to the identified group; it SHALL NOT affect any other download.

#### Scenario: Triggering resolution from one occurrence's id
- **WHEN** an operator submits a resolve request identified by the `download_info` id of one occurrence in a duplicate group
- **THEN** the system SHALL apply the Manual Resolve Action to every occurrence in that same group and return each occurrence's resulting outcome

#### Scenario: Resolve request for an occurrence with no duplicate
- **WHEN** an operator submits a resolve request identified by a `download_info` id that does not belong to any duplicate group
- **THEN** the system SHALL refuse the request without changing any record
