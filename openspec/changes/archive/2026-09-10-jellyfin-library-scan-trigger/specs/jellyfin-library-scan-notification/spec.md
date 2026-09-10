## Purpose

Notifies a configured Jellyfin server about the specific library folders that changed at the end of a download run, so newly downloaded movies and episodes become visible in Jellyfin without waiting for its own periodic or manual scan.

## ADDED Requirements

### Requirement: Jellyfin integration is opt-in and requires a base URL
The system SHALL only attempt to notify Jellyfin when the Jellyfin integration is enabled AND a base URL is configured. It SHALL NOT attempt any network call otherwise.

#### Scenario: Integration disabled by default
- **WHEN** a download run completes with the Jellyfin integration not explicitly enabled
- **THEN** the system SHALL NOT attempt to contact any Jellyfin server

#### Scenario: Enabled but not fully configured
- **WHEN** the Jellyfin integration is enabled but no base URL is configured
- **THEN** the system SHALL log a warning identifying the missing configuration
- **THEN** the system SHALL NOT attempt any network call to Jellyfin

### Requirement: Changed library paths are collected per run
For each of the `download` and `resume-downloads` commands, the system SHALL track the destination folder of every playlist item successfully completed during that single command invocation: the movie's folder for a movie, or the season's folder for a TV episode.

#### Scenario: Successful movie download is tracked
- **WHEN** a movie download completes successfully during a run
- **THEN** the movie's destination folder SHALL be included in that run's set of changed paths

#### Scenario: Successful episode download is tracked
- **WHEN** a TV episode download completes successfully during a run
- **THEN** the episode's season folder SHALL be included in that run's set of changed paths

#### Scenario: Failed items are not tracked
- **WHEN** an item fails to download during a run
- **THEN** its destination folder SHALL NOT be included in that run's set of changed paths

### Requirement: One targeted notification per run, not one per item
When a run's set of changed paths is non-empty and Jellyfin is configured, the system SHALL send exactly one notification for that run identifying the distinct changed paths, instructing Jellyfin to rescan only those paths rather than requesting a full library scan. Each distinct path SHALL be reported at most once, regardless of how many items within the run mapped to it.

#### Scenario: No items completed
- **WHEN** a run completes with zero successfully downloaded items
- **THEN** the system SHALL NOT send any notification to Jellyfin

#### Scenario: Multiple episodes of the same season in one run
- **WHEN** a run successfully downloads several episodes that belong to the same season
- **THEN** the system SHALL send a single notification for that run
- **THEN** that season's folder SHALL appear only once in the notification

#### Scenario: Multiple distinct items in one run
- **WHEN** a run successfully downloads items belonging to different movies and/or seasons
- **THEN** the system SHALL send a single notification for that run containing all of the distinct changed paths together

### Requirement: Notification delivery is best-effort and non-blocking
A failure to notify Jellyfin (unreachable server, timeout, authentication error, or any other error) SHALL NOT change the run's reported download statistics or exit status, and SHALL NOT prevent or delay completion of the run beyond a bounded timeout.

#### Scenario: Jellyfin unreachable
- **WHEN** the configured Jellyfin server cannot be reached
- **THEN** the run SHALL still report the same download statistics and exit status it would have reported without the Jellyfin integration
- **THEN** the failure SHALL be logged

#### Scenario: Jellyfin slow to respond
- **WHEN** the configured Jellyfin server does not respond within the notification's bounded timeout
- **THEN** the system SHALL abandon the notification attempt and proceed with the run's normal completion

### Requirement: Applies to both download and resume-downloads commands
Both the `download` command and the `resume-downloads` command SHALL independently apply the collection and notification behavior described above for the items they each complete.

#### Scenario: download command notifies
- **WHEN** the `download` command completes a run with at least one successful item and Jellyfin configured
- **THEN** a notification for that run's changed paths SHALL be sent

#### Scenario: resume-downloads command notifies
- **WHEN** the `resume-downloads` command completes a run with at least one successfully resumed item and Jellyfin configured
- **THEN** a notification for that run's changed paths SHALL be sent
