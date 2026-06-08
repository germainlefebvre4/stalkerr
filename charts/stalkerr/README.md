# Stalkerr Helm Chart

A Helm chart for deploying Stalkerr - an M3U playlist parser and media downloader that integrates with Sonarr and Radarr.

## Overview

Stalkerr parses M3U playlists, enriches content metadata using TMDB, and automatically downloads missing media items for your Sonarr/Radarr libraries. It runs as a set of scheduled jobs on Kubernetes with a REST API server for management.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Storage class supporting ReadWriteMany access mode (for shared media volumes)
- PostgreSQL 12+ (either via included subchart or external)
- Existing Jellyfin/Sonarr/Radarr deployment (optional, for media integration)

## Installation

### Add Repository (if published)

```bash
helm repo add stalkerr https://your-charts-repository
helm repo update
```

### Install from Local Chart

```bash
# Build dependencies
helm dependency build charts/stalkerr

# Install with default values
helm install stalkerr charts/stalkerr \
  --namespace jellyfin \
  --create-namespace

# Install with custom values
helm install stalkerr charts/stalkerr \
  --namespace jellyfin \
  --create-namespace \
  --values my-values.yaml
```

## Configuration

### Configuration Architecture

This Helm chart uses a **hybrid configuration approach** for maximum flexibility and security:

1. **Config File**: Application configuration is mounted as `/etc/stalkerr/config.yaml` from a ConfigMap
2. **Environment Variables**: Secrets (API keys, passwords) are injected as `STALKERR_*` prefixed environment variables
3. **Viper Override**: Environment variables override config file values at runtime

This design achieves:
- ✅ **1:1 Docker Compose compatibility**: Config structure matches exactly
- ✅ **Full feature support**: Complex filters, nested settings, arrays work correctly  
- ✅ **Kubernetes-native secrets**: Sensitive values stay in Secrets, not ConfigMaps
- ✅ **Reproducibility**: Values.yaml directly mirrors `config.yaml` structure

### Required Values

The following values **must** be configured before installation:

```yaml
# API Keys (required)
secrets:
  tmdbApiKey: "your-tmdb-api-key"
  radarrApiKey: "your-radarr-api-key"
  sonarrApiKey: "your-sonarr-api-key"

# M3U Playlist URL (required)
config:
  m3u:
    download:
      url: "https://example.com/playlist.m3u"
```

### Content Filtering Configuration

Configure include/exclude patterns to filter content from the M3U playlist:

```yaml
config:
  filter:
    group_title:
      include_patterns:
        - ".*"  # Match all
      exclude_patterns:
        - "^AR"  # Exclude Arabic content
        - "^TR"  # Exclude Turkish content
        - "^ES"  # Exclude Spanish content
    tvg_name:
      include_patterns:
        - ".*"
      exclude_patterns:
        - "Trailer$"   # Exclude trailers
        - "Sample$"    # Exclude samples
        - "^Test"      # Exclude test content
```

### Download Resume Configuration

Enable resumable downloads with progress tracking:

```yaml
config:
  downloads:
    resume_enabled: true                  # Enable download resume
    progress_interval_mb: 500             # Log progress every 500MB
    progress_interval_seconds: 60         # Or every 60 seconds
    lock_timeout_minutes: 5               # Timeout for stale locks
    max_retry_attempts: 5                 # Max retries for failed downloads
    movies_path: /media/jellyfin/radarr/movies
    tvshows_path: /media/jellyfin/sonarr/tv
    temp_dir: /tmp/stalkerr
    max_parallel: 1                       # Downloads in parallel
    timeout: 3600                         # Per-download timeout (seconds)
```

### M3U Download Configuration

Configure automatic M3U playlist downloads with archiving:

```yaml
config:
  m3u:
    file_path: /data/m3u/playlist.m3u
    update_interval: 3600
    download:
      enabled: true
      url: "https://provider.com/playlist.m3u"
      archive_dir: /data/m3u
      retention_count: 5                  # Keep last 5 downloads
      max_file_size_mb: 1024
      timeout_seconds: 300
      retry_attempts: 3
      # Optional HTTP authentication
      auth_username: ""
      auth_password: ""
      schedule_enabled: true
      interval_hours: 24
```

### Common Configuration Scenarios

#### 1. Using Existing PVCs for Media Sharing

**Recommended** for integration with Jellyfin/Sonarr/Radarr:

```yaml
storage:
  media:
    existingClaim: "jellyfin-media-pvc"  # Use existing shared PVC
  m3u:
    existingClaim: "stalkerr-m3u-pvc"    # Or create new PVC
```

#### 2. Using External PostgreSQL Database

```yaml
postgresql:
  enabled: false

config:
  database:
    host: "postgres.example.com"
    port: 5432
    user: "stalkerr"
    name: "stalkerr"
    sslmode: "require"

secrets:
  databasePassword: "your-db-password"
```

#### 3. Using Existing Secret

**Recommended** for production:

```yaml
secrets:
  existingSecret: "stalkerr-secrets"  # Secret must contain required keys
```

Required keys in existing secret:
- `tmdb-api-key`
- `radarr-api-key`
- `sonarr-api-key`
- `database-password` (if using external DB)

#### 4. Cross-Namespace Service Discovery

If Sonarr/Radarr are in a different namespace (e.g., `jellyfin`):

```yaml
config:
  sonarr:
    url: "http://sonarr.jellyfin.svc.cluster.local"
  radarr:
    url: "http://radarr.jellyfin.svc.cluster.local"
```

#### 5. Enabling Ingress with TLS

```yaml
ingress:
  enabled: true
  className: "traefik"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    traefik.ingress.kubernetes.io/router.middlewares: "authentik-ak-outpost@kubernetescrd"
  hosts:
    - host: stalkerr.example.com
      paths:
        - path: /
          pathType: Prefix
          port: 8080
  tls:
    - secretName: stalkerr-tls
      hosts:
        - stalkerr.example.com
```

## Migration from Docker Compose

### 1. Configuration Mapping

The Helm chart maintains **1:1 compatibility** with Docker Compose `config.yml` structure using snake_case field names:

| Docker Compose Config | Helm values.yaml |
|----------------------|------------------|
| `database.host` | `config.database.host` (auto-set if `postgresql.enabled=true`) |
| `database.dbname` | `config.database.dbname` |
| `m3u.file_path` | `config.m3u.file_path` |
| `m3u.download.url` | `config.m3u.download.url` |
| `m3u.download.archive_dir` | `config.m3u.download.archive_dir` |
| `m3u.download.retention_count` | `config.m3u.download.retention_count` |
| `filter.group_title.include_patterns` | `config.filter.group_title.include_patterns` |
| `filter.group_title.exclude_patterns` | `config.filter.group_title.exclude_patterns` |
| `filter.tvg_name.include_patterns` | `config.filter.tvg_name.include_patterns` |
| `filter.tvg_name.exclude_patterns` | `config.filter.tvg_name.exclude_patterns` |
| `downloads.movies_path` | `config.downloads.movies_path` |
| `downloads.tvshows_path` | `config.downloads.tvshows_path` |
| `downloads.temp_dir` | `config.downloads.temp_dir` |
| `downloads.max_parallel` | `config.downloads.max_parallel` |
| `downloads.resume_enabled` | `config.downloads.resume_enabled` |
| `downloads.progress_interval_mb` | `config.downloads.progress_interval_mb` |
| `logging.format` | `config.logging.format` |
| `logging.app.level` | `config.logging.app.level` |
| `logging.database.level` | `config.logging.database.level` |
| `tmdb.language` | `config.tmdb.language` |
| `tmdb.requests_per_second` | `config.tmdb.requests_per_second` |
| `radarr.url` | `config.radarr.url` |
| `radarr.sync_interval` | `config.radarr.sync_interval` |
| `radarr.quality_profile_id` | `config.radarr.quality_profile_id` |
| `sonarr.url` | `config.sonarr.url` |
| `sonarr.sync_interval` | `config.sonarr.sync_interval` |
| `sonarr.quality_profile_id` | `config.sonarr.quality_profile_id` |
| Cron schedules | `jobs.<jobName>.schedule` |

**Note**: Secret values (API keys, passwords) are injected as environment variables from Kubernetes Secrets, not from values.yaml.

### 2. Prepare Storage

Option A: Use existing volumes
```bash
# Create PVCs from existing volumes
kubectl create -f existing-pvcs.yaml
```

Option B: Migrate data
```bash
# Copy data to new PVCs
kubectl cp ./m3u_playlist/ pod-name:/data/m3u/
```

### 3. Create Secrets

```bash
kubectl create secret generic stalkerr-secrets \
  --from-literal=tmdb-api-key="YOUR_KEY" \
  --from-literal=radarr-api-key="YOUR_KEY" \
  --from-literal=sonarr-api-key="YOUR_KEY" \
  --from-literal=database-password="YOUR_PASSWORD" \
  -n media
```

### 4. Install Chart

```bash
helm install stalkerr charts/stalkerr \
  --namespace media \
  --set secrets.existingSecret=stalkerr-secrets \
  --set storage.media.existingClaim=jellyfin-media \
  --set config.m3u.downloadUrl="YOUR_M3U_URL"
```

## Storage Paths and Media Organization

### Mount Paths

- `/data/m3u` - M3U playlist storage
- `/media/jellyfin/iptv` - IPTV downloads (default output)
- `/media/jellyfin/radarr` - Radarr movie downloads
- `/media/jellyfin/sonarr` - Sonarr TV show downloads

### Sharing Media with Jellyfin/Sonarr/Radarr

Use the **same PVC** for media storage across all applications:

```yaml
# Stalkerr
storage:
  media:
    existingClaim: "shared-media-pvc"

# Sonarr (in jellyfin namespace)
persistence:
  media:
    existingClaim: "shared-media-pvc"

# Radarr (in jellyfin namespace)
persistence:
  media:
    existingClaim: "shared-media-pvc"

# Jellyfin (in jellyfin namespace)
persistence:
  media:
    existingClaim: "shared-media-pvc"
```

Mount paths should align:
- Stalkerr writes to: `/media/jellyfin/radarr/Movie.mkv`
- Radarr reads from: `/media/jellyfin/radarr/Movie.mkv`
- Jellyfin scans: `/media/jellyfin/radarr/`

## Database Migration

### Running Migrations

Migrations run automatically when the server starts. To run manually:

```bash
kubectl exec -it deployment/stalkerr-server -- ./stalkeer migrate
```

### Backing Up Database

For embedded PostgreSQL:

```bash
# Dump database
kubectl exec -it statefulset/stalkerr-postgresql-0 -- \
  pg_dump -U stalkerr stalkerr > backup.sql

# Restore database
kubectl exec -i statefulset/stalkerr-postgresql-0 -- \
  psql -U stalkerr stalkerr < backup.sql
```

## Troubleshooting

### PVC Not Binding

Check storage class availability:
```bash
kubectl get storageclass
kubectl get pvc -n media
kubectl describe pvc stalkerr-media -n media
```

Ensure `accessMode: ReadWriteMany` is supported by your storage provider.

### CronJob Not Running

Check job history:
```bash
kubectl get cronjobs -n media
kubectl get jobs -n media
kubectl logs job/stalkerr-m3u-download-xxxxx -n media
```

Verify schedule syntax:
```bash
kubectl get cronjob stalkerr-m3u-download -n media -o yaml
```

### Cross-Namespace Access Issues

Verify DNS resolution:
```bash
kubectl run -it --rm debug --image=busybox --restart=Never -- \
  nslookup sonarr.jellyfin.svc.cluster.local
```

Check NetworkPolicy (if enabled):
```bash
kubectl get networkpolicy -n media
```

### API Keys Not Working

Verify secrets are created:
```bash
kubectl get secret stalkerr-secrets -n media
kubectl get secret stalkerr-secrets -n media -o jsonpath='{.data.tmdb-api-key}' | base64 -d
```

Check environment variables in pods:
```bash
kubectl exec deployment/stalkerr-server -n media -- env | grep API_KEY
```

## Values Reference

### Global Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `ghcr.io/yourusername/stalkerr` |
| `image.tag` | Image tag (defaults to chart appVersion) | `""` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Server Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `server.enabled` | Enable API server deployment | `true` |
| `server.replicaCount` | Number of replicas | `1` |
| `server.resources.limits.cpu` | CPU limit | `2000m` |
| `server.resources.limits.memory` | Memory limit | `2Gi` |

### CronJob Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `jobs.m3uDownload.schedule` | M3U download schedule | `"30 23 * * *"` (23:30 daily) |
| `jobs.process.schedule` | Processing schedule | `"0 0 * * *"` (midnight daily) |
| `jobs.sonarrSync.schedule` | Sonarr sync schedule | `"0 */2 * * *"` (every 2 hours) |
| `jobs.sonarrSync.forceSync.schedule` | Sonarr force sync | `"30 2 * * *"` (02:30 daily) |
| `jobs.radarrSync.schedule` | Radarr sync schedule | `"0 1-23/2 * * *"` (odd hours) |
| `jobs.radarrSync.forceSync.schedule` | Radarr force sync | `"30 5 * * *"` (05:30 daily) |

### Storage Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `storage.m3u.enabled` | Enable M3U PVC creation | `true` |
| `storage.m3u.existingClaim` | Use existing PVC | `""` |
| `storage.m3u.size` | PVC size | `1Gi` |
| `storage.media.enabled` | Enable media PVC creation | `true` |
| `storage.media.existingClaim` | Use existing PVC (recommended) | `""` |
| `storage.media.size` | PVC size | `100Gi` |

### Database Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `postgresql.enabled` | Deploy PostgreSQL subchart | `true` |
| `postgresql.auth.username` | Database username | `stalkerr` |
| `postgresql.auth.database` | Database name | `stalkerr` |
| `config.database.host` | External DB host (if postgresql.enabled=false) | `""` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class | `traefik` |
| `ingress.hosts[0].host` | Hostname | `stalkerr.example.com` |

For complete values reference, see `values.yaml`.

## Chart Metadata

- **Version**: 0.1.0
- **App Version**: 1.0.0
- **Kubernetes Version**: >= 1.19.0
- **Dependencies**: PostgreSQL 15.5.38 (Bitnami)

## License

See project repository for license information.

## Support

For issues and questions:
- GitHub Issues: https://github.com/yourusername/stalkerr/issues
- Documentation: https://github.com/yourusername/stalkerr/docs
