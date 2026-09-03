# scheduled-jobs Specification

## Purpose
TBD - created by archiving change helm-chart-stalkeer. Update Purpose after archive.
## Requirements
### Requirement: M3U download CronJob
The chart SHALL create a CronJob resource for downloading M3U playlists on a configurable schedule.

#### Scenario: M3U download job created
- **WHEN** jobs.m3uDownload.enabled is true
- **THEN** CronJob resource is created with schedule from values
- **THEN** Job command is ["./stalkeer", "m3u-download", "--config", "/app/config/config.yml"]
- **THEN** Job mounts m3u PVC and config

#### Scenario: M3U download schedule
- **WHEN** jobs.m3uDownload.schedule is "30 23 * * *"
- **THEN** CronJob runs daily at 23:30
- **THEN** CronJob concurrencyPolicy prevents overlapping executions

### Requirement: Process CronJob
The chart SHALL create a CronJob resource for processing M3U playlists and enriching with TMDB data.

#### Scenario: Process job created
- **WHEN** jobs.process.enabled is true
- **THEN** CronJob resource is created with schedule from values
- **THEN** Job command is ["./stalkeer", "process", "--config", "/app/config/config.yml"]
- **THEN** Job has access to m3u PVC and database

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

### Requirement: Job resource configuration
The chart SHALL allow configuration of resources (CPU, memory) for all CronJobs.

#### Scenario: Per-job resources
- **WHEN** jobs.<jobName>.resources is configured
- **THEN** CronJob pod spec includes resource limits and requests
- **THEN** Default resources are provided for each job type

### Requirement: Job restart policy
The chart SHALL configure CronJobs with appropriate restart policies for batch workloads.

#### Scenario: Restart on failure
- **WHEN** CronJob pod fails
- **THEN** Job restartPolicy is "OnFailure"
- **THEN** Failed jobs are retained for debugging (failedJobsHistoryLimit: 3)
- **THEN** Successful jobs are retained (successfulJobsHistoryLimit: 3)

### Requirement: Shared environment configuration
The chart SHALL inject the same configuration (ConfigMap, Secret) into all CronJobs.

#### Scenario: Consistent configuration
- **WHEN** Any CronJob is created
- **THEN** Job pod template includes envFrom referencing shared ConfigMap
- **THEN** Job pod template includes envFrom referencing shared Secret
- **THEN** Database connection is available to all jobs

