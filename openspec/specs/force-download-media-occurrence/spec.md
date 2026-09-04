# force-download-media-occurrence Specification

## Purpose

Lets a user manually pull one specific ingested playlist occurrence of a movie or TV episode that is already matched to Radarr/Sonarr but was never fetched, bypassing the automatic pipeline's permanent blind spot on media that already has any file.

## Requirements

### Requirement: Force download targets exactly one occurrence
The system SHALL allow triggering a download for one specific playlist occurrence (`ProcessedLine`), identified individually, for a movie or TV episode. This SHALL NOT affect, alter, or re-attempt any sibling occurrence of the same movie/TV episode, regardless of those siblings' own state.

#### Scenario: Forcing an occurrence when a sibling is already downloaded
- **WHEN** a user requests a forced download for an occurrence whose associated movie already has a different occurrence in state "downloaded"
- **THEN** the system SHALL proceed with the requested occurrence only, and SHALL NOT change the state or file of the already-downloaded sibling occurrence

#### Scenario: Forcing an occurrence does not consider other candidates
- **WHEN** a user requests a forced download for a specific occurrence and other untried occurrences exist for the same movie/TV episode
- **THEN** the system SHALL attempt only the requested occurrence, and SHALL NOT automatically substitute or additionally attempt any other occurrence

### Requirement: Forced download is refused for ineligible occurrences
The system SHALL refuse a forced-download request, returning a clear error and performing no side effect, when any of the following holds for the target occurrence:
- it is not matched to a movie or TV show,
- its own state already reflects a completed download,
- a download for that exact occurrence is already in progress.

#### Scenario: Occurrence not matched to a movie or TV show
- **WHEN** a forced-download request targets an occurrence with no associated movie or TV show
- **THEN** the system SHALL refuse the request without contacting Radarr/Sonarr or creating any download record

#### Scenario: Occurrence already downloaded
- **WHEN** a forced-download request targets an occurrence whose own state is already "downloaded"
- **THEN** the system SHALL refuse the request and SHALL NOT re-download or re-create its file

#### Scenario: Occurrence already downloading
- **WHEN** a forced-download request targets an occurrence for which a download is already in progress
- **THEN** the system SHALL refuse the duplicate request rather than starting a second concurrent download of the same occurrence

### Requirement: Live existence check gates the trigger, fail-closed
Before starting a forced download, the system SHALL verify, through a live call made at the moment the request is submitted, that the occurrence's associated movie exists in the connected Radarr instance (or its associated TV episode exists in the connected Sonarr instance). When that media cannot be confirmed to exist, or the live call fails for any reason (timeout, network error, unreachable service, unexpected response), the system SHALL refuse to start the download and SHALL report the failure to the caller. No download record SHALL be created and no state SHALL change for a refused request.

#### Scenario: Movie confirmed present in Radarr
- **WHEN** a forced-download request targets a movie occurrence and Radarr confirms the movie exists in its library
- **THEN** the system SHALL proceed to the asynchronous download step

#### Scenario: Movie not known to Radarr
- **WHEN** a forced-download request targets a movie occurrence and Radarr reports no matching movie in its library
- **THEN** the system SHALL refuse the request and SHALL report that the media was not found in Radarr

#### Scenario: Episode confirmed present in Sonarr
- **WHEN** a forced-download request targets a TV episode occurrence and Sonarr confirms the corresponding series/episode exists in its library
- **THEN** the system SHALL proceed to the asynchronous download step

#### Scenario: Radarr/Sonarr unreachable at request time
- **WHEN** the live existence check to Radarr or Sonarr times out, errors, or is unreachable
- **THEN** the system SHALL refuse the forced-download request and SHALL NOT start any download, treating the indeterminate result as "does not exist"

### Requirement: Existence check is independent of "missing" status
The existence check used to gate a forced download SHALL determine whether the media is known at all to the connected Radarr/Sonarr library, regardless of whether that instance currently has a file for it. It SHALL NOT rely on, or be limited to, that instance's list of missing/wanted items.

#### Scenario: Media already has a file in Radarr
- **WHEN** the target movie already has a file in Radarr and therefore no longer appears in Radarr's missing/wanted list
- **THEN** the existence check SHALL still report that the movie exists, and the forced download SHALL be allowed to proceed (subject to the other eligibility rules)

#### Scenario: Media never added to the library
- **WHEN** the target movie or series was never added to Radarr/Sonarr's library
- **THEN** the existence check SHALL report that the media does not exist, and the forced download SHALL be refused

### Requirement: No existence check outside an actual trigger
The system SHALL NOT perform the Radarr/Sonarr existence check, or any other live external call, as a side effect of a user merely viewing an occurrence's details. That check SHALL be performed only when a forced-download request is actually submitted for that occurrence.

#### Scenario: Viewing occurrence details triggers no external call
- **WHEN** a user views the details of an occurrence without requesting a forced download
- **THEN** the system SHALL NOT call Radarr or Sonarr to check that occurrence's media existence

### Requirement: Forced download executes asynchronously
Once eligibility and existence are confirmed, the system SHALL accept the forced-download request immediately and perform the actual file transfer asynchronously in the background, using the same download-tracking record and state transitions as the existing automatic pipeline, so its progress and completion remain observable through the existing downloads listing.

#### Scenario: Request is accepted before the file transfer completes
- **WHEN** a forced-download request passes eligibility and the existence check
- **THEN** the system SHALL acknowledge acceptance of the request without waiting for the file transfer to finish, and the transfer SHALL continue in the background

#### Scenario: Forced download progress is visible through existing tracking
- **WHEN** an accepted forced download is in progress or has completed
- **THEN** its status SHALL be retrievable through the same download listing used for automatically-triggered downloads

### Requirement: Forced download destination naming avoids overwriting siblings
When a forced download completes, the system SHALL name the destination file so that it includes the occurrence's detected resolution as a distinguishing suffix (for example, a file named `Title (Year) [1080p].ext`), so that it cannot silently replace a sibling occurrence's file already present in the same movie/show destination folder. When the occurrence has no detected resolution, the system SHALL still apply a distinguishing marker unique to that download rather than omit the suffix. In addition, the destination filename SHALL include the occurrence's detected language and, when present, its Québec French variant, as tags following the resolution suffix (for example, `Title (Year) [1080p][MULTI][VFQ].ext`), per the `download-filename-quality-tags` capability. This naming rule SHALL apply only to files produced by a forced download.

#### Scenario: Forced HD download does not overwrite an existing SD file
- **WHEN** a movie already has a downloaded SD file in its Radarr destination folder and a forced download of a different, higher-resolution occurrence of the same movie completes
- **THEN** the new file SHALL be written under a resolution-suffixed name distinct from the existing SD file, and the existing SD file SHALL remain intact

#### Scenario: Forced download with no detected resolution still avoids collision
- **WHEN** a forced download completes for an occurrence with no detected resolution
- **THEN** the system SHALL still write the file under a name that cannot collide with a sibling occurrence's file in the same destination folder

#### Scenario: Forced download filename also carries language and variant tags
- **WHEN** a forced download completes for an occurrence with a detected language and a detected Québec French variant
- **THEN** the destination filename SHALL include the resolution suffix as well as the language and Québec French variant tags

#### Scenario: Automatic pipeline naming is unaffected
- **WHEN** the existing automatic `radarr`/`sonarr`-sourced `download` command completes a download
- **THEN** the destination file name SHALL NOT gain the forced-download-specific anti-collision fallback marker described by this requirement; it follows the shared `download-filename-quality-tags` tagging rule instead (see that capability), which now applies resolution/language/Québec French variant tags to automatic-pipeline downloads too
