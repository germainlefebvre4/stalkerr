# storage-configuration Specification

## Purpose
TBD - created by archiving change helm-chart-stalkeer. Update Purpose after archive.
## Requirements
### Requirement: M3U playlist storage
The chart SHALL provide PVC for M3U playlist file storage with configurable size and access mode.

#### Scenario: M3U PVC created
- **WHEN** storage.m3u.enabled is true
- **THEN** PersistentVolumeClaim is created for m3u storage
- **THEN** PVC size is configurable (default 1Gi)
- **THEN** Access mode is configurable (default ReadWriteMany)
- **THEN** Storage class is configurable or uses cluster default

#### Scenario: Existing M3U PVC
- **WHEN** storage.m3u.existingClaim is set
- **THEN** No new PVC is created
- **THEN** Existing PVC is mounted to pods

### Requirement: Media storage configuration
The chart SHALL provide PVC for media downloads shared with Jellyfin/Sonarr/Radarr.

#### Scenario: Media PVC created
- **WHEN** storage.media.enabled is true
- **THEN** PersistentVolumeClaim is created for media storage
- **THEN** PVC size is configurable (default 100Gi)
- **THEN** Access mode defaults to ReadWriteMany for sharing

#### Scenario: Existing media PVC (recommended)
- **WHEN** storage.media.existingClaim is set
- **THEN** No new PVC is created
- **THEN** Existing PVC is mounted to pods
- **THEN** Mount path is /media allowing subpath access

### Requirement: Volume mounts for server
The chart SHALL mount storage volumes to the server deployment at specified paths.

#### Scenario: Server volume mounts
- **WHEN** Server deployment is created
- **THEN** M3U PVC is mounted at /data/m3u
- **THEN** Media PVC is mounted at /media
- **THEN** ConfigMap is mounted at /app/config/config.yml (read-only)

### Requirement: Volume mounts for CronJobs
The chart SHALL mount storage volumes to all CronJob pods.

#### Scenario: CronJob volume mounts
- **WHEN** Any CronJob is created
- **THEN** M3U PVC is mounted at /data/m3u
- **THEN** Media PVC is mounted at /media
- **THEN** CronJobs can read and write to shared storage

### Requirement: Temporary storage
The chart SHALL configure temporary storage for download operations.

#### Scenario: Temp directory
- **WHEN** Pods are created
- **THEN** emptyDir volume is created for /tmp
- **THEN** Temp storage is ephemeral (cleared on pod restart)
- **THEN** config.downloads.tempDir points to /tmp/stalkeer

### Requirement: Subpath configuration
The chart SHALL support subpath mounts for organizing media within shared storage.

#### Scenario: Media subpaths
- **WHEN** storage.media.subPaths is configured
- **THEN** Radarr downloads go to media PVC subPath "radarr"
- **THEN** Sonarr downloads go to media PVC subPath "sonarr"
- **THEN** IPTV downloads go to media PVC subPath "iptv"
- **THEN** Subpaths match Jellyfin/Sonarr/Radarr expectations

