## Context

Stalkeer is currently deployed on a bare-metal NUC using Docker Compose with system cron triggering one-shot containers. The application consists of:
- A long-running API server (ports 8080/8081)
- Batch jobs (m3u-download, process, sonarr-sync, radarr-sync) triggered by crontab
- PostgreSQL database
- Shared volumes with Jellyfin/Sonarr/Radarr for media files

The target Kubernetes cluster already hosts Jellyfin, Sonarr, and Radarr in the `jellyfin` namespace with:
- Traefik as ingress controller
- Authentik for authentication (managed separately, not in this chart)
- NFS-backed storage with ReadWriteMany support
- Existing PVCs for media storage at `/media/{wd-elements,toshiba}/nuc6i5syh/jellyfin/`

Current cron schedule:
- 23:30 daily: m3u-download
- 00:00 daily: process
- */2 hours: sonarr-sync
- 02:30 daily: sonarr-sync --force
- 1-23/2 hours: radarr-sync (odd hours)
- 05:30 daily: radarr-sync --force

## Goals / Non-Goals

**Goals:**
- Create a production-ready Helm chart following best practices (similar to CVWonder Studio reference chart)
- Maintain functional parity with Docker Compose deployment
- Enable declarative configuration through values.yaml
- Support shared storage with existing media stack (Jellyfin/Sonarr/Radarr)
- Allow cross-namespace service discovery for Sonarr/Radarr APIs
- Provide flexible PostgreSQL deployment (embedded or external)
- Support standard Kubernetes Ingress for API exposure

**Non-Goals:**
- Authentik integration (authentication handled externally via middleware)
- Traefik IngressRoute CRDs (use standard Ingress only)
- Horizontal scaling of the API server (single replica is sufficient for batch workload)
- Migration automation from Docker Compose (manual configuration mapping required)
- Management of Sonarr/Radarr deployments (chart only consumes their APIs)

## Decisions

### 1. Chart Structure: Follow CVWonder Studio Pattern

**Decision**: Model the chart structure after the CVWonder Studio Helm chart, adapting for Stalkeer's batch-oriented architecture.

**Rationale**:
- CVWonder provides a proven template for web apps with PostgreSQL
- Similar patterns: ConfigMap/Secret separation, optional subchart dependency, storage management
- Well-structured with values.schema.json for validation

**Differences**:
- Stalkeer adds CronJobs (CVWonder is deployment-only)
- No need for Gotenberg-like auxiliary services
- Different storage layout (m3u + media vs. sessions + themes)

### 2. Job Architecture: CronJobs with Force Variants

**Decision**: Create separate CronJob resources for normal and force-sync operations rather than using a single CronJob with arguments.

**Rationale**:
- Clear separation in scheduling (regular sync vs. full resync)
- Easier to enable/disable independently via values.yaml
- Simpler to monitor and debug (distinct job names in kubectl/logs)

**Implementation**:
```yaml
jobs:
  sonarrSync:
    enabled: true
    schedule: "0 */2 * * *"
    forceSync:
      enabled: true
      schedule: "30 2 * * *"
```

Generates two CronJob resources:
- `stalkeer-sonarr-sync` with `["./stalkeer", "sonarr"]`
- `stalkeer-sonarr-sync-force` with `["./stalkeer", "sonarr", "--force"]`

**Alternative Considered**: Single CronJob with schedule array (not supported by Kubernetes CronJob spec).

### 3. Storage Strategy: Existing PVC Reference

**Decision**: Use `existingClaim` as the primary pattern for media storage, with optional new PVC creation as fallback.

**Rationale**:
- Jellyfin/Sonarr/Radarr already have PVCs for media files
- Stalkeer must write to the same storage paths that Jellyfin reads
- Avoids PVC duplication and ensures data consistency
- User knows their storage topology better than chart assumptions

**Configuration**:
```yaml
storage:
  m3u:
    enabled: true
    existingClaim: ""  # Optional
    size: 1Gi
    accessMode: ReadWriteMany
  media:
    enabled: true
    existingClaim: "jellyfin-media"  # Recommended
    # Or create new PVC with same backend
```

**Mount Paths**:
- `/data/m3u` → m3u playlist PVC
- `/media/jellyfin/radarr` → media PVC subpath (matches Radarr)
- `/media/jellyfin/sonarr` → media PVC subpath (matches Sonarr)
- `/media/jellyfin/iptv` → media PVC subpath (IPTV downloads)

### 4. Cross-Namespace Integration: DNS-Based Service Discovery

**Decision**: Use Kubernetes DNS for cross-namespace service access without custom NetworkPolicy in the chart.

**Rationale**:
- Kubernetes DNS provides built-in cross-namespace resolution: `<service>.<namespace>.svc.cluster.local`
- NetworkPolicy setup varies by cluster (some don't use it)
- User can add NetworkPolicy via values if needed

**Configuration**:
```yaml
config:
  sonarr:
    url: http://sonarr.jellyfin.svc.cluster.local
  radarr:
    url: http://radarr.jellyfin.svc.cluster.local
```

**Optional NetworkPolicy** (if user enables it):
```yaml
networkPolicy:
  enabled: false  # User sets to true if needed
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: jellyfin
```

### 5. PostgreSQL: Bitnami Subchart Dependency

**Decision**: Include Bitnami PostgreSQL chart as optional dependency (enabled by default).

**Rationale**:
- Simple default deployment (no external DB required)
- Production users can disable and use external managed PostgreSQL
- Proven chart with good defaults
- Auto-generates connection URL when enabled

**Configuration**:
```yaml
postgresql:
  enabled: true  # Disable for external DB
  auth:
    database: stalkeer
    username: stalkeer
    password: ""  # Auto-generated if empty
```

When `postgresql.enabled=true`, the chart auto-populates database connection in environment:
```
DATABASE_URL=postgres://stalkeer:${password}@stalkeer-postgresql:5432/stalkeer?sslmode=disable
```

### 6. Configuration: ConfigMap + Secret Split

**Decision**: Separate non-sensitive config (ConfigMap) from credentials (Secret) with optional `existingSecret` support.

**Rationale**:
- Follows Kubernetes security best practices
- Compatible with external secret managers (Sealed Secrets, External Secrets Operator)
- Clear boundary between what's safe to version control vs. what's not

**Sensitive Values** (Secret):
- Database password
- TMDB API key
- Radarr API key
- Sonarr API key

**Non-Sensitive Values** (ConfigMap):
- Log levels, formats
- Download paths
- M3U file settings
- Service URLs (without credentials)

### 7. Ingress: Standard Kubernetes Ingress Only

**Decision**: Provide only standard Kubernetes Ingress resource, not Traefik IngressRoute CRD.

**Rationale**:
- Portability across ingress controllers (Traefik, nginx, etc.)
- User requirement to avoid IngressRoute complexity
- Users can manually create IngressRoute if needed (chart documentation shows how)

**Configuration**:
```yaml
ingress:
  enabled: true
  className: traefik
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: stalkeer.example.com
      paths:
        - path: /
          pathType: Prefix
          port: 8080
```

## Risks / Trade-offs

**[Risk: Storage path misalignment]**
→ **Mitigation**: Document exact mount paths and provide examples for common setups. Include validation in values.schema.json.

**[Risk: CronJob overlap causing storage conflicts]**
→ **Mitigation**: Use ReadWriteMany PVCs. Schedule jobs to avoid overlaps (e.g., radarr at odd hours, sonarr at even hours). Document concurrency behavior.

**[Risk: PostgreSQL subchart version drift]**
→ **Mitigation**: Pin Bitnami PostgreSQL version in Chart.yaml dependencies. Test upgrades before releasing.

**[Risk: Cross-namespace service discovery blocked by NetworkPolicy]**
→ **Mitigation**: Document NetworkPolicy requirements. Provide example NetworkPolicy in chart README.

**[Trade-off: No automatic migration from Docker Compose]**
→ **Accepted**: Manual configuration mapping required. Provide migration guide in chart README comparing docker-compose.yml to values.yaml.

**[Trade-off: Single replica API server]**
→ **Accepted**: Stalkeer is primarily batch-oriented. API is used for monitoring/management only. Horizontal scaling not needed initially, can be added later if required.

**[Trade-off: Force-sync as separate CronJobs doubles resource definitions]**
→ **Accepted**: Clarity and manageability outweigh template duplication. Reduces operational complexity for enabling/disabling force syncs.

## Migration Plan

**Phase 1: Chart Development**
1. Create chart structure and base templates
2. Implement server deployment + service
3. Add CronJob templates
4. Configure storage and database integration
5. Add ingress and RBAC

**Phase 2: Testing**
1. Test with internal PostgreSQL (default)
2. Test with external PostgreSQL
3. Verify CronJob schedules and job execution
4. Validate storage mounts with dummy data
5. Test cross-namespace Sonarr/Radarr connectivity

**Phase 3: Production Migration**
1. Create namespace (`kubectl create ns stalkeer`)
2. Map docker-compose.yml settings to values.yaml
3. Create secrets externally (TMDB, Radarr, Sonarr API keys)
4. Reference existing media PVC in values
5. Deploy chart with PostgreSQL enabled
6. Run database migration (via init container or manual job)
7. Verify CronJobs execute successfully
8. Update DNS/Ingress to point to new deployment
9. Disable Docker Compose stack
10. Monitor for one week before decommissioning old setup

**Rollback Strategy**:
- Keep Docker Compose configuration intact during migration
- Database backup before migration
- DNS can be reverted to Docker Compose host
- CronJobs can be suspended without deleting chart

## Open Questions

None - all major decisions have been made during exploration phase.
