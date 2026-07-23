## ADDED Requirements

### Requirement: Helm Values Schema Validation for Frontend Properties
The Helm Chart's `values.schema.json` file SHALL define and enforce validation rules for the `frontend` configuration block, ensuring type safety and structural correctness for all frontend parameters (including `enabled`, `replicaCount`, `image`, `service`, and `resources`).

#### Scenario: Successful schema validation of frontend values
- **WHEN** a user templates or installs the chart with structurally valid parameters under `frontend` in their values
- **THEN** the Helm schema validation engine SHALL pass successfully without errors.

#### Scenario: Failed schema validation on invalid frontend values
- **WHEN** a user templates or installs the chart with invalid types (e.g. `enabled: "yes"` instead of `true`, or a non-integer `replicaCount`)
- **THEN** the Helm schema validation engine SHALL fail and output descriptive errors specifying the invalid fields.
