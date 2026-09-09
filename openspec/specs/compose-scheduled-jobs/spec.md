# compose-scheduled-jobs Specification

## Purpose

Gives Docker Compose deployments the same scheduled, periodic execution of the M3U download/process/download pipeline that Kubernetes deployments already get from the Helm chart's `CronJob` resources.

## Requirements

### Requirement: One-shot job services match current CLI
The Docker Compose `stalkerr` profile SHALL provide one-shot job services for exactly the commands supported by the current CLI (`process`, `download`, `m3u-download`), with no service invoking a removed or obsolete command.

#### Scenario: Job services reflect current CLI
- **WHEN** a developer lists the `stalkerr` profile's services
- **THEN** the one-shot job services include `process`, `download`, and `m3u-download`
- **THEN** no service invokes the removed `sonarr`/`radarr` commands or the obsolete `resume-downloads` command

### Requirement: Optional cron sidecar drives periodic execution
`docker-compose.yml` SHALL provide an optional cron sidecar service that triggers the `m3u-download`, `process`, and `download` one-shot job services on a schedule, without requiring any scheduling logic inside the Stalkeer application itself.

#### Scenario: Enabling the cron sidecar
- **WHEN** a user starts the stack with the cron sidecar service enabled
- **THEN** the sidecar periodically triggers `m3u-download`, `process`, and `download` on their configured schedules by invoking each one-shot job service

#### Scenario: Cron sidecar is optional
- **WHEN** a user starts the stack without enabling the cron sidecar service
- **THEN** the stack behaves as it does today (server plus one-shot job services available for manual or externally-triggered runs) and no periodic execution occurs

### Requirement: Default schedule matches Helm chart cadence
The cron sidecar SHALL default to the same schedule as the Helm chart's `CronJob` resources, so the two deployment methods behave consistently out of the box.

#### Scenario: Default cadence
- **WHEN** the cron sidecar runs with its default configuration
- **THEN** `m3u-download` runs daily at 23:30
- **THEN** `process` runs daily at 00:00
- **THEN** `download` runs every 2 hours

### Requirement: No overlapping scheduled runs
The cron sidecar SHALL NOT start a new scheduled run of a job while a previous run of that same job is still in progress.

#### Scenario: Overlapping run avoided
- **WHEN** a scheduled trigger for a job fires while the previous run of that same job has not yet completed
- **THEN** the new run is skipped or deferred rather than started concurrently with the still-running one

### Requirement: Compose scheduling setup is documented
The project's `README.md` and `DOCKER-QUICKSTART.md` SHALL describe how to enable scheduled periodic execution for Docker Compose deployments, replacing any vague or missing guidance.

#### Scenario: Reading Compose scheduling instructions
- **WHEN** a developer follows `DOCKER-QUICKSTART.md`'s setup steps
- **THEN** they find a concrete recipe for enabling scheduled `m3u-download`/`process`/`download` runs under Docker Compose, equivalent to the guidance already given for the Helm chart's `CronJob` resources
