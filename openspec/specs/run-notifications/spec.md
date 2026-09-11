# run-notifications Specification

## Purpose

Gives operators visibility into failed or broken scheduled runs (`download`, `process`, `m3u-download`) via a push notification, without requiring anyone to open the dashboard or inspect job logs.

## Requirements

### Requirement: Notifications are opt-in and off by default
The system SHALL send no notification of any kind unless notifications are explicitly enabled in configuration.

#### Scenario: Notifications disabled
- **WHEN** notifications are not enabled in configuration
- **THEN** no notification is sent regardless of run outcome, including on structural failures

#### Scenario: Notifications enabled with no channel configured
- **WHEN** notifications are enabled but no channel (e.g. ntfy) has valid configuration (server/topic)
- **THEN** the command logs that delivery was skipped and completes with its normal exit code, without failing due to the missing channel configuration

### Requirement: Successful runs stay silent
A run that completes with zero failures SHALL NOT produce a notification.

#### Scenario: Download run with no failures
- **WHEN** a `download` run completes and every item downloaded successfully
- **THEN** no notification is sent

#### Scenario: Process run with no errors
- **WHEN** a `process` run completes with zero errors
- **THEN** no notification is sent

### Requirement: Download run failure summary
When a `download` run completes with at least one failed item, the system SHALL send a notification summarizing the run.

#### Scenario: Download run with at least one failure
- **WHEN** a `download` run finishes and one or more items failed to download
- **THEN** a notification is sent
- **THEN** the notification includes the total item count, the number downloaded, and the number failed

### Requirement: Process run failure summary
When a `process` run completes with one or more errors, the system SHALL send a notification summarizing the run.

#### Scenario: Process run with at least one error
- **WHEN** a `process` run finishes with one or more errors recorded
- **THEN** a notification is sent
- **THEN** the notification includes the total processed count and the error count

### Requirement: Structural failure alerts
When a run aborts before producing a normal summary because a required upstream dependency could not be used, the system SHALL send a notification distinguishing this from a run-with-failures summary.

#### Scenario: Radarr/Sonarr unreachable
- **WHEN** a `download` run cannot build its work set because Radarr and/or Sonarr could not be reached
- **THEN** a notification is sent identifying the failure as unable to reach Radarr/Sonarr
- **THEN** the command still exits with its existing non-zero exit code

#### Scenario: M3U playlist fetch failure
- **WHEN** an `m3u-download` run fails to fetch the configured playlist URL
- **THEN** a notification is sent identifying the failure as a playlist fetch failure
- **THEN** the command still exits with its existing non-zero exit code

#### Scenario: M3U file cannot be parsed
- **WHEN** a `process` run fails to parse its input M3U file
- **THEN** a notification is sent identifying the failure as a file parse failure
- **THEN** the command still exits with its existing non-zero exit code

### Requirement: Notification delivery never changes command outcome
A failure to deliver a notification SHALL NOT alter the exit code, output, or side effects of the command that triggered it.

#### Scenario: Notification channel unreachable
- **WHEN** the configured ntfy server cannot be reached or returns an error
- **THEN** the triggering command completes with the same exit code and output it would have had if notifications were disabled
- **THEN** the delivery failure is logged
