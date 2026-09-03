## REMOVED Requirements

### Requirement: Sonarr sync CronJobs
**Reason**: Replaced by a single unified download CronJob (see ADDED Requirements below) that fetches and downloads from both Radarr and Sonarr in one run, sharing one concurrency budget instead of two independently-scheduled, independently-concurrent pipelines.
**Migration**: Remove `jobs.sonarrSync` (and `jobs.sonarrSync.forceSync`) from values; configure the new `jobs.download` key instead. The force variant's schedule is no longer a separate CronJob — force/upgrade candidates are now drawn probabilistically within every unified run (see `media-download-scheduling`).

The chart SHALL create CronJob resources for syncing with Sonarr (normal and force variants).

#### Scenario: Normal Sonarr sync
- **WHEN** jobs.sonarrSync.enabled is true
- **THEN** CronJob "sonarr-sync" is created with schedule from values
- **THEN** Job command is ["./stalkeer", "sonarr", "--config", "/app/config/config.yml"]

#### Scenario: Force Sonarr sync
- **WHEN** jobs.sonarrSync.forceSync.enabled is true
- **THEN** CronJob "sonarr-sync-force" is created with separate schedule
- **THEN** Job command includes --force flag
- **THEN** Force sync schedule defaults to "30 2 * * *"

### Requirement: Radarr sync CronJobs
**Reason**: Replaced by the same single unified download CronJob described above.
**Migration**: Remove `jobs.radarrSync` (and `jobs.radarrSync.forceSync`) from values; configure the new `jobs.download` key instead.

The chart SHALL create CronJob resources for syncing with Radarr (normal and force variants).

#### Scenario: Normal Radarr sync
- **WHEN** jobs.radarrSync.enabled is true
- **THEN** CronJob "radarr-sync" is created with schedule from values
- **THEN** Job command is ["./stalkeer", "radarr", "--config", "/app/config/config.yml"]

#### Scenario: Force Radarr sync
- **WHEN** jobs.radarrSync.forceSync.enabled is true
- **THEN** CronJob "radarr-sync-force" is created with separate schedule
- **THEN** Job command includes --force flag
- **THEN** Force sync schedule defaults to "30 5 * * *"

## ADDED Requirements

### Requirement: Unified download CronJob
The chart SHALL create a single CronJob resource that runs the unified `download` command against both Radarr and Sonarr, on a single configurable schedule, replacing the separate Radarr/Sonarr (normal and force) CronJobs.

#### Scenario: Unified download job created
- **WHEN** jobs.download.enabled is true
- **THEN** CronJob "download" is created with schedule from values
- **THEN** Job command is ["./stalkeer", "download", "--config", "/app/config/config.yml"]
- **THEN** CronJob concurrencyPolicy prevents overlapping executions

#### Scenario: Unified job carries both service credentials
- **WHEN** the unified download CronJob's pod is created
- **THEN** it SHALL have access to both the Radarr and Sonarr API key secrets
- **THEN** it SHALL mount both the M3U and media storage volumes

#### Scenario: No separate force schedule
- **WHEN** the unified download CronJob is configured
- **THEN** there SHALL be no separate `forceSync`-style CronJob or schedule; force/upgrade candidates are handled within every run of the single schedule
