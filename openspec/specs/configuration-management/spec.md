# configuration-management Specification

## Purpose
Defines how the Helm chart delivers application configuration to pods — via a ConfigMap for non-sensitive settings and a Secret for credentials/API keys — including database URL composition and safe defaults.
## Requirements
### Requirement: ConfigMap for non-sensitive configuration
The chart SHALL create a ConfigMap containing all non-sensitive application configuration.

#### Scenario: ConfigMap created
- **WHEN** Chart is installed
- **THEN** ConfigMap is created with application settings
- **THEN** ConfigMap includes database config (host, port, name, sslmode)
- **THEN** ConfigMap includes the `m3u.sources` list (each entry's name, file path, and download config)
- **THEN** ConfigMap includes download settings (paths, timeouts, parallelism)
- **THEN** ConfigMap includes logging config (format, levels)

#### Scenario: Service URL configuration
- **WHEN** ConfigMap is created
- **THEN** Sonarr URL is included (e.g., http://sonarr.jellyfin.svc.cluster.local)
- **THEN** Radarr URL is included (e.g., http://radarr.jellyfin.svc.cluster.local)
- **THEN** URLs support cross-namespace service discovery

### Requirement: Secret for sensitive credentials
The chart SHALL create a Secret containing all sensitive credentials and API keys.

#### Scenario: Secret created
- **WHEN** existingSecret is not set
- **THEN** Secret is created with application credentials
- **THEN** Secret includes database password
- **THEN** Secret includes TMDB API key
- **THEN** Secret includes Radarr API key
- **THEN** Secret includes Sonarr API key
- **THEN** All values are base64 encoded

### Requirement: External secret reference
The chart SHALL allow referencing an external Secret instead of creating one.

#### Scenario: Existing secret
- **WHEN** existingSecret is set to a secret name
- **THEN** Chart does not create a Secret resource
- **THEN** Pods reference the existing secret
- **THEN** Secret must contain expected keys (database-password, tmdb-api-key, radarr-api-key, sonarr-api-key)

### Requirement: Environment variable injection
The chart SHALL inject configuration into pods via envFrom for ConfigMap and Secret.

#### Scenario: Environment from ConfigMap
- **WHEN** Pod is created
- **THEN** envFrom references ConfigMap
- **THEN** All ConfigMap keys become environment variables
- **THEN** Variables are available to application process

#### Scenario: Environment from Secret
- **WHEN** Pod is created
- **THEN** envFrom references Secret
- **THEN** Secret keys become environment variables
- **THEN** Sensitive values are not logged

### Requirement: Database URL composition
The chart SHALL compose database connection URL from individual configuration values.

#### Scenario: Internal PostgreSQL URL
- **WHEN** postgresql.enabled is true
- **THEN** DATABASE_URL is set via environment variable
- **THEN** URL includes auth from postgresql.auth values
- **THEN** URL includes service name from subchart

#### Scenario: External database URL
- **WHEN** postgresql.enabled is false
- **THEN** DATABASE_URL is composed from config.database values
- **THEN** Password is sourced from Secret

### Requirement: Configuration file generation
The chart SHALL generate a config.yml file mounted into pods as a volume.

#### Scenario: Config file mount
- **WHEN** Pods are created
- **THEN** ConfigMap data is mounted as /app/config/config.yml
- **THEN** File is read-only
- **THEN** Application reads config from this path

### Requirement: Default value safety
The chart SHALL require user to provide sensitive values (no insecure defaults).

#### Scenario: Empty sensitive defaults
- **WHEN** values.yaml is used without overrides
- **THEN** secrets.tmdbApiKey is empty string
- **THEN** secrets.radarrApiKey is empty string
- **THEN** secrets.sonarrApiKey is empty string
- **THEN** Installation fails or warns if required keys are missing

### Requirement: ConfigMap renders the M3U sources list
The chart SHALL require the chart user to configure `values.config.m3u.sources` as a non-empty list, and SHALL render it as the `m3u.sources` list in the generated `config.yml`. The chart SHALL NOT render a singular `m3u.file_path` / `m3u.download` block.

#### Scenario: Multiple sources rendered in ConfigMap
- **WHEN** `values.config.m3u.sources` is set to a list with more than one entry, each providing a `name`, `file_path`, and `download` block
- **THEN** the generated `config.yml` SHALL include an `m3u.sources` list with one entry per configured source, each carrying its own `file_path` and `download` settings

#### Scenario: A single configured source is still a sources list
- **WHEN** `values.config.m3u.sources` is set to a list with exactly one entry
- **THEN** the generated `config.yml` SHALL render that single entry under `m3u.sources`, and SHALL NOT render a separate singular `m3u.file_path` / `m3u.download` block

#### Scenario: Missing or empty sources value fails validation
- **WHEN** `values.config.m3u.sources` is not set, or is set to an empty list
- **THEN** chart installation/templating SHALL fail schema validation, since `m3u.sources` is a required, non-empty field
