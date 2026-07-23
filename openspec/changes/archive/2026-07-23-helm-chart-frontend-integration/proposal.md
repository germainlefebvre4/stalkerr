## Why

The Stalkerr Helm Chart supports deploying the React frontend alongside the Go API server, but the `values.schema.json` file is missing validation definitions for the `frontend` block. This leads to untyped and unvalidated frontend configurations in Helm. Adding these schema validations will align the Helm chart's validation engine with its actual capabilities and prevent misconfigurations.

## What Changes

- Update `charts/stalkerr/values.schema.json` to declare and validate the `frontend` configuration block (including `enabled`, `replicaCount`, `image`, `service`, and `resources`).
- Update `openspec/specs/helm-chart-frontend-integration/spec.md` to fully document the specification of the frontend integration within the Helm Chart.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `helm-chart-frontend-integration`: Fully define, validate, and integrate the frontend component properties within the Helm Chart structure.

## Impact

- **Helm Chart Validation Schema**: `charts/stalkerr/values.schema.json` will be updated with frontend schemas.
- **Project Specifications**: `openspec/specs/helm-chart-frontend-integration/spec.md` will be updated to document the completed specs.
