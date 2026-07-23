# helm-chart-structure Specification

## Purpose
TBD - created by archiving change helm-chart-stalkeer. Update Purpose after archive.
## Requirements
### Requirement: Chart metadata structure
The chart SHALL provide a valid Chart.yaml file with API version v2, chart name, description, version, appVersion, and dependencies.

#### Scenario: Valid Chart.yaml
- **WHEN** Helm validates the chart
- **THEN** Chart.yaml contains all required fields (apiVersion, name, version, appVersion)
- **THEN** Chart.yaml declares PostgreSQL dependency with condition flag

### Requirement: Default configuration values
The chart SHALL provide a values.yaml file with comprehensive default settings for all configurable aspects.

#### Scenario: Complete values structure
- **WHEN** User installs chart without custom values
- **THEN** All required configuration sections exist (image, server, config, jobs, storage, postgresql, ingress)
- **THEN** Sensitive defaults (API keys, passwords) are empty strings requiring user input

### Requirement: JSON Schema validation
The chart SHALL provide a values.schema.json file for validating user-supplied values.

#### Scenario: Schema validation
- **WHEN** User provides invalid values (wrong type, missing required field)
- **THEN** Helm validates values against schema before installation
- **THEN** Clear error messages indicate what's invalid

### Requirement: Template helpers
The chart SHALL provide _helpers.tpl with reusable template functions for labels, names, and selectors.

#### Scenario: Consistent labeling
- **WHEN** Any resource is created
- **THEN** Common labels (app.kubernetes.io/name, app.kubernetes.io/instance, etc.) are applied consistently
- **THEN** Helper functions generate deterministic resource names

### Requirement: Chart documentation
The chart SHALL provide a README.md with installation instructions, configuration reference, and migration guide.

#### Scenario: Complete documentation
- **WHEN** User reads README
- **THEN** Installation steps are clear (prerequisites, values to set, helm install command)
- **THEN** All values.yaml options are documented with descriptions
- **THEN** Migration guide from Docker Compose is provided

