## Why

Stalkeer is currently deployed using Docker Compose with system-level cron jobs, which limits scalability, observability, and integration with the existing Kubernetes-based media stack (Jellyfin, Sonarr, Radarr). A Helm chart enables declarative deployment, better resource management, and native integration with cluster services while maintaining the same functional architecture.

## What Changes

- Create a production-ready Helm chart in `charts/stalkeer/` with full template structure
- Deploy API server as a Kubernetes Deployment with health checks and service exposure
- Convert Docker Compose one-shot services to CronJob resources with configurable schedules
- Integrate PostgreSQL as a Bitnami subchart dependency with optional external database support
- Implement shared storage strategy using existing PVCs for media files (compatible with Jellyfin/Sonarr/Radarr)
- Configure cross-namespace service discovery for Sonarr and Radarr APIs in the `jellyfin` namespace
- Provide standard Kubernetes Ingress for HTTPS exposure (compatible with Traefik)
- Separate configuration into ConfigMap (non-sensitive) and Secret (API keys, credentials)
- Support both new deployments and migration from existing Docker Compose setups

## Capabilities

### New Capabilities
- `helm-chart-structure`: Complete Helm chart with Chart.yaml, values.yaml, values.schema.json, and helper templates
- `api-server-deployment`: Kubernetes Deployment for the Stalkeer server with health checks, resource limits, and service exposure
- `scheduled-jobs`: CronJob resources for m3u-download, process, sonarr-sync, radarr-sync with normal and force-sync variants
- `database-integration`: PostgreSQL deployment using Bitnami subchart with connection configuration and migration support
- `storage-configuration`: PVC templates for m3u playlist storage and shared media volumes compatible with existing media stack
- `ingress-configuration`: Standard Kubernetes Ingress resource for HTTPS access with TLS certificate management
- `configuration-management`: ConfigMap and Secret templates with environment variable injection and optional external secret support

### Modified Capabilities
<!-- No existing capabilities are being modified -->

## Impact

**New Files**:
- `charts/stalkeer/Chart.yaml` - Chart metadata and PostgreSQL dependency
- `charts/stalkeer/values.yaml` - Default configuration values
- `charts/stalkeer/values.schema.json` - JSON Schema validation
- `charts/stalkeer/templates/_helpers.tpl` - Template helper functions
- `charts/stalkeer/templates/deployment.yaml` - Server deployment
- `charts/stalkeer/templates/service.yaml` - Kubernetes service
- `charts/stalkeer/templates/cronjob-*.yaml` - Multiple CronJob resources
- `charts/stalkeer/templates/configmap.yaml` - Application configuration
- `charts/stalkeer/templates/secret.yaml` - Sensitive credentials
- `charts/stalkeer/templates/pvc-*.yaml` - Storage claims
- `charts/stalkeer/templates/ingress.yaml` - Ingress resource
- `charts/stalkeer/templates/serviceaccount.yaml` - RBAC configuration
- `charts/stalkeer/README.md` - Chart documentation

**Configuration Migration**: Requires mapping `config.yml` and `docker-compose.yml` settings to Helm values

**Dependencies**: Bitnami PostgreSQL chart (optional, can use external database)

**Cross-Namespace Integration**: Requires network policies to allow egress to `jellyfin` namespace for Sonarr/Radarr API access

**Storage Requirements**: NFS or ReadWriteMany-capable storage class for shared media volumes
