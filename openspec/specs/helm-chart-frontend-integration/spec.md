# helm-chart-frontend-integration Specification

## Purpose
Fully integrate and validate the decoupled React frontend component in the Helm Chart, enabling robust deployment, service exposure, traffic routing, and value-schema safety.

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

### Requirement: Helm Values Schema Validation for Frontend Properties
The Helm Chart's `values.schema.json` file SHALL define and enforce validation rules for the `frontend` configuration block, ensuring type safety and structural correctness for all frontend parameters (including `enabled`, `replicaCount`, `image`, `service`, and `resources`).

#### Scenario: Successful schema validation of frontend values
- **WHEN** a user templates or installs the chart with structurally valid parameters under `frontend` in their values
- **THEN** the Helm schema validation engine SHALL pass successfully without errors.

#### Scenario: Failed schema validation on invalid frontend values
- **WHEN** a user templates or installs the chart with invalid types (e.g. `enabled: "yes"` instead of `true`, or a non-integer `replicaCount`)
- **THEN** the Helm schema validation engine SHALL fail and output descriptive errors specifying the invalid fields.
