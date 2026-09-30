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
The system SHALL provide an on-demand action that reconciles one duplicate group's database state with the filesystem: for every completed occurrence in the group, the system SHALL verify whether its file still exists at its stored `download_path`. Any occurrence whose file no longer exists SHALL be transitioned to the `cancelled` status (`download_info.status` and the corresponding `processed_lines.state`), matching the terminal status the `cancel-download-occurrence` capability already uses. An occurrence whose file is still present SHALL be left unchanged. This action SHALL only ever run when explicitly triggered by an operator request; it SHALL NOT be invoked automatically by the scheduled reconciliation pass or any other background process. This verify-only action SHALL NOT delete, move, or otherwise modify any file on disk; deleting a file is only ever performed by the separate Keep One Duplicate Action.

#### Scenario: Resolving a group where one file was manually removed
- **WHEN** an operator triggers the resolve action for a duplicate group where one occurrence's file has been deleted from disk and the other's still exists
- **THEN** the occurrence whose file is gone SHALL be transitioned to `cancelled`, and the occurrence whose file still exists SHALL be left unchanged

#### Scenario: Resolving a group where both files still exist
- **WHEN** an operator triggers the resolve action for a duplicate group where every occurrence's file is still present on disk
- **THEN** no occurrence SHALL be changed, no file SHALL be deleted, and the group SHALL continue to be reported as a duplicate

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

### Requirement: Duplicate Candidates Description
The API server SHALL expose a read-only endpoint that, given the `download_info` id of any one occurrence belonging to a duplicate group, describes the group's distinct files so an operator can choose which one to keep. For each distinct `download_path` in the group it SHALL report: the path, whether the file currently exists on disk, its size in bytes when it exists, the most recent completion date among the occurrences sharing that path, and the `download_info` ids of those occurrences. When the movie's/series' Radarr/Sonarr library entry can be reached, it SHALL also report for each path whether the file's base name matches the canonical name derived from that live entry; when the lookup is unavailable (service not configured, unreachable, or the item not found), the canonical-name information SHALL be omitted and the endpoint SHALL still succeed. The response SHALL flag at most one existing file as suggested: the existing file whose name matches the canonical name if there is one, otherwise the existing file with the most recent completion date; no file SHALL be flagged when none exists. This endpoint SHALL NOT modify any record or file.

#### Scenario: Describing a group with two distinct existing files
- **WHEN** an operator requests the candidates for a duplicate group with two distinct paths whose files both exist, one of them named after the live Radarr title
- **THEN** the response SHALL list both paths with their existence, size, latest completion date and occurrence ids, and SHALL flag the file whose name matches the Radarr title as suggested

#### Scenario: Occurrences sharing a path are one candidate
- **WHEN** two occurrences of the group share the same `download_path`
- **THEN** the response SHALL list that path once, carrying both occurrences' `download_info` ids

#### Scenario: Library lookup unavailable
- **WHEN** Radarr/Sonarr is not configured or unreachable when candidates are requested
- **THEN** the response SHALL still list every candidate without canonical-name information, and SHALL flag as suggested the existing file with the most recent completion date

#### Scenario: Missing file is never suggested
- **WHEN** the most recent candidate's file no longer exists on disk while another candidate's file does
- **THEN** the response SHALL report the missing file with `exists` false and SHALL NOT flag it as suggested

#### Scenario: Candidates requested for an occurrence with no duplicate
- **WHEN** the requested `download_info` id does not belong to any duplicate group
- **THEN** the system SHALL refuse the request without changing any record

### Requirement: Keep One Duplicate Action
The system SHALL provide an operator-triggered action that resolves a duplicate group by keeping exactly one file and removing the rest. The operator identifies the file to keep by the `download_info` id of one occurrence; the file to keep is that occurrence's `download_path`, and every occurrence of the group sharing that exact path SHALL be kept unchanged. For every other distinct `download_path` in the group, the system SHALL delete that file from disk if it exists, and SHALL transition every occurrence of the group with that path to the `cancelled` status (`download_info.status` and the corresponding `processed_lines.state`), matching the terminal status the `cancel-download-occurrence` capability already uses. A path whose file is already absent SHALL still have its occurrences cancelled, without any deletion attempt. The system SHALL refuse the request, changing nothing, when the file to keep does not currently exist on disk. This action SHALL only ever run when explicitly triggered by an operator request and SHALL NOT be invoked by the scheduled reconciliation pass or any other background process. After a successful run, the group SHALL no longer be reported as a duplicate.

#### Scenario: Keeping one of two existing files
- **WHEN** an operator triggers the action for a group with two distinct existing files, identifying the occurrence of file A as the one to keep
- **THEN** file B SHALL be deleted from disk, every occurrence with file B's path SHALL be transitioned to `cancelled`, every occurrence with file A's path SHALL be left unchanged, and the group SHALL no longer be reported as a duplicate

#### Scenario: Multiple occurrences share the kept path
- **WHEN** the file to keep is shared by two occurrences and the other distinct file is shared by two others
- **THEN** both occurrences of the kept path SHALL be left unchanged, the other file SHALL be deleted once, and both occurrences of the removed path SHALL be `cancelled`

#### Scenario: An already-missing file is cancelled without deletion
- **WHEN** a group has three distinct paths, the operator keeps one, and one of the other two files is already absent from disk
- **THEN** the present other file SHALL be deleted and its occurrences cancelled, and the absent file's occurrences SHALL be cancelled without any deletion attempt

#### Scenario: Kept file is missing
- **WHEN** the operator identifies an occurrence whose `download_path` does not exist on disk as the file to keep
- **THEN** the system SHALL refuse the request, SHALL NOT delete any file, and SHALL NOT change any record

#### Scenario: Identified occurrence is not in a duplicate group
- **WHEN** the operator identifies a `download_info` id that does not belong to any duplicate group
- **THEN** the system SHALL refuse the request without deleting any file or changing any record

#### Scenario: Action never runs unattended
- **WHEN** the scheduled `download` command's reconciliation pass detects a duplicate group
- **THEN** it SHALL only log the finding and SHALL NOT invoke the Keep One Duplicate Action

### Requirement: Keep One Deletion Safety
The Keep One Duplicate Action SHALL only ever delete a file whose path is a stored `download_path` of an occurrence in the identified duplicate group, and SHALL NEVER delete the file being kept, nor any file of another download. It SHALL only delete regular files: a path that resolves to a directory or to a symbolic link SHALL NOT be deleted. Before deleting anything, the system SHALL validate every path it intends to delete and SHALL refuse the whole request, changing nothing, if any of them is not a regular file (other than one that is simply absent). Each file's deletion SHALL be completed before the database state of the occurrences that reference it is changed; if deleting a file fails for any reason other than the file being absent, the system SHALL stop, leave the occurrences of that path unchanged, report the failure, and leave paths already processed as processed.

#### Scenario: A directory at a recorded path is never deleted
- **WHEN** a non-kept `download_path` of the group resolves to a directory
- **THEN** the system SHALL refuse the request, SHALL NOT delete any file, and SHALL NOT change any record

#### Scenario: A symbolic link at a recorded path is never followed or deleted
- **WHEN** a non-kept `download_path` of the group is a symbolic link
- **THEN** the system SHALL refuse the request, SHALL NOT delete any file, and SHALL NOT change any record

#### Scenario: A failed deletion leaves the database untouched for that file
- **WHEN** deleting a non-kept file fails with an error other than "not found"
- **THEN** the occurrences referencing that file SHALL remain unchanged, the system SHALL report the failure, and the kept file SHALL be intact

#### Scenario: Files outside the group are untouched
- **WHEN** the action runs for a group
- **THEN** no file other than the group's non-kept recorded paths SHALL be deleted, and no download outside the group SHALL be modified

### Requirement: Keep One Endpoint And Outcome Reporting
The API server SHALL expose an endpoint that, given the `download_info` id of the occurrence to keep, applies the Keep One Duplicate Action to that occurrence's duplicate group. Its response SHALL report, for every occurrence in the group, an outcome of kept, deleted (its file was removed and the occurrence cancelled), or cancelled because its file was already missing, together with the occurrence's `download_path`, and SHALL report the outcome of the Radarr/Sonarr refresh described by the requirement below.

#### Scenario: Response reports every occurrence
- **WHEN** the action completes for a group of four occurrences over two distinct paths
- **THEN** the response SHALL contain four entries, each with the occurrence's `download_info` id, `download_path` and outcome

### Requirement: Radarr And Sonarr Refresh After Keep One
After the Keep One Duplicate Action deletes a file or cancels an occurrence, the system SHALL request Radarr (for a movie) or Sonarr (for a TV episode's series) to perform its "Refresh & Scan" on that specific movie or series, so the service's library state stops referencing the removed files. The request SHALL only be made when the service is configured, SHALL be made at most once per action, and SHALL be attempted even when a later step of the action reported a partial failure after at least one file was deleted. A failure of this request (service unreachable, timeout, authentication error, item not found in the library, or any other error) SHALL NOT undo any deletion or database change and SHALL NOT change the action's success; it SHALL instead be reported in the response as a distinct refresh outcome (requested, failed, not configured, or not found in library) so the operator can retry the refresh manually. This request SHALL NOT be made by the verify-only Manual Resolve Action.

#### Scenario: Refresh requested after a successful deletion
- **WHEN** the action deletes a movie's extra file and Radarr is configured and reachable
- **THEN** the system SHALL request a Radarr "Refresh & Scan" for that movie exactly once and report the refresh outcome as requested

#### Scenario: Refresh failure does not undo the deletion
- **WHEN** the file has been deleted and its occurrences cancelled, but the Radarr refresh request fails
- **THEN** the deletion and cancellations SHALL remain in effect, the action SHALL still be reported as completed, and the refresh outcome SHALL be reported as failed

#### Scenario: Service not configured
- **WHEN** the action completes for a TV episode's group and Sonarr is not configured
- **THEN** the system SHALL NOT attempt a Sonarr call and SHALL report the refresh outcome as not configured
