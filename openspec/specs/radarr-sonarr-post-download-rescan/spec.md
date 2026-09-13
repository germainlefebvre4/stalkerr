# radarr-sonarr-post-download-rescan Specification

## Purpose

Notifies Radarr and Sonarr about the specific movies and series that the `download` command just finished downloading, so their own missing/wanted status updates promptly instead of waiting on their periodic library scan.

## Requirements

### Requirement: Rescan only attempted for a configured, reachable service
The system SHALL only attempt to notify Radarr about a completed movie when Radarr is configured (base URL and API key present) for this run, and SHALL only attempt to notify Sonarr about a completed episode when Sonarr is configured for this run. It SHALL NOT attempt a rescan call for a service that is not configured.

#### Scenario: Sonarr not configured
- **WHEN** the `download` command completes an episode's download and Sonarr is not configured for this run
- **THEN** the system SHALL NOT attempt any Sonarr rescan call

### Requirement: Distinct completed work units are tracked per run
For the `download` command, the system SHALL track every distinct movie (by Radarr movie ID) and every distinct series (by Sonarr series ID) that had at least one item successfully complete during that single command invocation.

#### Scenario: Multiple episodes of the same series completing in one run
- **WHEN** three episodes of the same series each complete successfully within one `download` run
- **THEN** that series SHALL be tracked once as a completed work unit for that run, not three times

### Requirement: A targeted rescan is requested once per distinct completed work unit
When a run's set of completed movies and/or series is non-empty, the system SHALL request Radarr to rescan/refresh each distinct completed movie, and SHALL request Sonarr to rescan/refresh each distinct completed series, so each is requested at most once per run regardless of how many of its items completed.

#### Scenario: One rescan request per completed movie
- **WHEN** a movie's download completes successfully during a `download` run
- **THEN** the system SHALL request Radarr to rescan/refresh that specific movie exactly once for that run

#### Scenario: One rescan request per completed series
- **WHEN** one or more episodes of a series complete successfully during a `download` run
- **THEN** the system SHALL request Sonarr to rescan/refresh that specific series exactly once for that run

### Requirement: A rescan notification failure is non-fatal
A failure to notify Radarr or Sonarr of a completed download (unreachable service, timeout, authentication error, or any other error) SHALL NOT change the run's reported download statistics or exit status, and SHALL NOT prevent or delay completion of the run beyond a bounded timeout.

#### Scenario: Radarr rescan call fails
- **WHEN** the Radarr rescan request for a completed movie fails
- **THEN** the run's reported statistics and exit status SHALL be unaffected, and the failure SHALL be logged

### Requirement: Radarr and Sonarr rescans are independent of each other
A failure or unavailability affecting Radarr rescan notifications SHALL NOT prevent Sonarr rescan notifications from being attempted for completed series in the same run, and vice versa.

#### Scenario: Radarr unreachable does not block Sonarr rescans
- **WHEN** Radarr is unreachable during a run's post-download rescan step, and Sonarr is configured and reachable
- **THEN** the system SHALL still attempt the Sonarr rescan requests for that run's completed series
