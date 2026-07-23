## Context

The Stalkerr Helm chart already contains template definitions for independent frontend deployment (`deployment-frontend.yaml`), service (`service-frontend.yaml`), and unified ingress routing (`ingress.yaml`). The default values in `values.yaml` define the `frontend` configuration, but the Helm chart's validation schema (`values.schema.json`) does not define the `frontend` block. This leaves frontend properties unvalidated, potentially leading to hard-to-debug runtime deployment failures if invalid values are passed.

## Goals / Non-Goals

**Goals:**
- Update `values.schema.json` to include validation schemas for the `frontend` block, mirroring the structure of `values.yaml` (e.g. `enabled`, `replicaCount`, `image`, `service`, and `resources`).
- Ensure full backward compatibility with the existing chart defaults.
- Verify that the updated chart passes schema validation with default values.

**Non-Goals:**
- Modifying the actual deployment templates or services of the frontend unless bugs are found.
- Changing the build system or Dockerfiles of the frontend component.

## Decisions

### 1. Match Schema with Current Values Structure
We will add `frontend` property definition under `properties` in `values.schema.json`. The properties defined will include:
- `enabled`: boolean, required.
- `replicaCount`: integer (minimum 1), optional.
- `image`: object with required `repository` and `pullPolicy`, and optional `tag`.
- `service`: object with required `type` and `port`.
- `resources`: object.

### 2. Standard JSON Schema Draft-07 format
We will align the structure exactly with other components (such as `server`) in `values.schema.json` to maintain schema consistency and readability.

## Risks / Trade-offs

- **Risk**: Strict validation could reject existing production custom deployment values.
  - **Mitigation**: We will ensure that none of the newly added validations are overly restrictive or mandatory if they can be omitted in typical usage, except basic types and standard required fields.
