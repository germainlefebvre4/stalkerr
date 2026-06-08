## 1. Chart Foundation

- [x] 1.1 Create charts/stalkeer directory structure
- [x] 1.2 Create Chart.yaml with metadata and PostgreSQL dependency
- [x] 1.3 Create initial values.yaml with complete structure
- [x] 1.4 Create values.schema.json for values validation
- [x] 1.5 Create templates/_helpers.tpl with name, labels, and selector helpers

## 2. Configuration Management

- [x] 2.1 Create templates/configmap.yaml for non-sensitive config
- [x] 2.2 Create templates/secret.yaml with conditional creation based on existingSecret
- [x] 2.3 Add database URL composition logic in helpers
- [x] 2.4 Configure environment variable injection patterns for pods

## 3. RBAC and Security

- [x] 3.1 Create templates/serviceaccount.yaml
- [x] 3.2 Add security context configuration to values.yaml
- [x] 3.3 Configure pod security contexts in deployment and cronjob templates

## 4. API Server Deployment

- [x] 4.1 Create templates/deployment.yaml for server
- [x] 4.2 Configure health check probes (liveness and readiness)
- [x] 4.3 Add resource limits and requests configuration
- [x] 4.4 Configure volume mounts (config, m3u PVC, media PVC)
- [x] 4.5 Add environment variable injection (ConfigMap and Secret)
- [x] 4.6 Create templates/service.yaml for ClusterIP service

## 5. Scheduled Jobs

- [x] 5.1 Create templates/cronjob-m3u-download.yaml
- [x] 5.2 Create templates/cronjob-process.yaml
- [x] 5.3 Create templates/cronjob-sonarr-sync.yaml (normal variant)
- [x] 5.4 Create templates/cronjob-sonarr-sync-force.yaml (force variant)
- [x] 5.5 Create templates/cronjob-radarr-sync.yaml (normal variant)
- [x] 5.6 Create templates/cronjob-radarr-sync-force.yaml (force variant)
- [x] 5.7 Configure job restart policies and history limits
- [x] 5.8 Add resource configuration for each job type

## 6. Storage Configuration

- [x] 6.1 Create templates/pvc-m3u.yaml with conditional creation
- [x] 6.2 Create templates/pvc-media.yaml with conditional creation
- [x] 6.3 Configure volume claim references in deployment
- [x] 6.4 Configure volume claim references in all cronjobs
- [x] 6.5 Add emptyDir configuration for temporary storage
- [x] 6.6 Document subpath usage for media organization

## 7. Database Integration

- [x] 7.1 Configure PostgreSQL subchart in Chart.yaml dependencies
- [x] 7.2 Add postgresql configuration section in values.yaml
- [x] 7.3 Implement DATABASE_URL environment variable logic
- [x] 7.4 Add external database configuration support
- [x] 7.5 Document database migration process in README

## 8. Ingress Configuration

- [x] 8.1 Create templates/ingress.yaml with standard Kubernetes Ingress
- [x] 8.2 Configure host and path routing
- [x] 8.3 Add TLS configuration support
- [x] 8.4 Add annotations support for ingress controller features
- [x] 8.5 Add ingressClassName configuration
- [x] 8.6 Make ingress optional via enabled flag

## 9. Values and Defaults

- [x] 9.1 Configure default image repository and tag
- [x] 9.2 Set default CronJob schedules matching current cron setup
- [x] 9.3 Configure default resource limits for all components
- [x] 9.4 Set default storage sizes (m3u: 1Gi, media: 100Gi)
- [x] 9.5 Configure default Sonarr/Radarr URLs for cross-namespace access
- [x] 9.6 Add default logging configuration
- [x] 9.7 Ensure sensitive defaults are empty strings

## 10. Testing and Validation

- [x] 10.1 Test helm template rendering with default values
- [x] 10.2 Test helm template rendering with custom values
- [x] 10.3 Validate all required fields in values.schema.json
- [x] 10.4 Test with postgresql.enabled=true (internal DB)
- [x] 10.5 Test with postgresql.enabled=false (external DB)
- [x] 10.6 Test with existingSecret configuration
- [x] 10.7 Test with existing PVC references
- [x] 10.8 Validate CronJob schedules render correctly

## 11. Documentation

- [x] 11.1 Create charts/stalkeer/README.md with overview
- [x] 11.2 Document installation prerequisites (storage class, namespaces)
- [x] 11.3 Document required values (API keys, database config)
- [x] 11.4 Provide example values files for common scenarios
- [x] 11.5 Create migration guide from Docker Compose
- [x] 11.6 Document cross-namespace integration with Sonarr/Radarr
- [x] 11.7 Add troubleshooting section (PVC binding, CronJob failures)
- [x] 11.8 Document storage paths and media sharing with Jellyfin
- [x] 11.9 Add Chart.yaml keywords, sources, and maintainer info
- [x] 11.10 Create NOTES.txt template for post-install instructions

## 12. Production Readiness

- [x] 12.1 Add NOTES.txt with post-install checklist
- [x] 12.2 Verify all templates follow Helm best practices
- [x] 12.3 Ensure labels and selectors are consistent across all resources
- [x] 12.4 Validate resource naming follows conventions
- [x] 12.5 Test chart installation on development cluster
- [x] 12.6 Verify CronJobs execute successfully
- [x] 12.7 Test storage mounts and file access
- [x] 12.8 Validate cross-namespace service discovery works
