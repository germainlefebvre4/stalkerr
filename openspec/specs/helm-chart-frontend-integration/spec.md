# helm-chart-frontend-integration Specification

## Purpose
TBD - created by archiving change frontend-ihm-react-radix. Update Purpose after archive.
## Requirements
### Requirement: Independent Frontend Service in Helm Chart
The Helm Chart SHALL provide distinct deployment and service templates to compile, deploy, and scale the frontend application independently from the Go API backend.

#### Scenario: Enable and deploy frontend templates
- **WHEN** a user installs the Helm chart with `frontend.enabled=true`
- **THEN** Kubernetes SHALL deploy a separate replica set for the frontend (using Nginx serving the compiled SPA) and expose it via a dedicated `ClusterIP` Service.

### Requirement: Unified Ingress Routing
The Helm Chart's Ingress template SHALL support unified traffic routing to expose the entire application under a single host address.

#### Scenario: Access frontend and API on a single ingress
- **WHEN** a client requests `/` or other static routes versus `/api/v1/*` through the Ingress controller
- **THEN** the Ingress controller SHALL route standard path requests to the frontend service, and `/api/v1/*` requests directly to the backend API service.

