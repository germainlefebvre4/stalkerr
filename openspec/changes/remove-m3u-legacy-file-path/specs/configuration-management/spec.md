## MODIFIED Requirements

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

## REMOVED Requirements

### Requirement: ConfigMap renders multiple M3U sources
**Reason**: This requirement's "in addition to the existing singular m3u block" framing and its "Single-source configuration is unaffected" scenario no longer apply — the singular block is removed and `m3u.sources` becomes the only, required configuration shape. Replaced by the new "ConfigMap renders the M3U sources list" requirement below, scoped to the sources-only behavior.
**Migration**: Chart users must set `values.config.m3u.sources` to a non-empty list; a values file that only set the singular `config.m3u.file_path` / `config.m3u.download` block must be converted to a one-entry `sources` list.

## ADDED Requirements

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
