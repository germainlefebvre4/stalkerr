## ADDED Requirements

### Requirement: ConfigMap renders multiple M3U sources
The chart SHALL allow rendering an `m3u.sources` list into the generated `config.yml`, in addition to the existing singular `m3u` block, when the chart user configures more than one M3U source via values.

#### Scenario: Multiple sources rendered in ConfigMap
- **WHEN** `values.config.m3u.sources` is set to a non-empty list, each entry providing a `name`, `file_path`, and `download` block
- **THEN** the generated `config.yml` SHALL include an `m3u.sources` list with one entry per configured source, each carrying its own `file_path` and `download` settings

#### Scenario: Single-source configuration is unaffected
- **WHEN** `values.config.m3u.sources` is not set
- **THEN** the generated `config.yml` SHALL contain only the existing singular `m3u` block, unchanged from today's behavior
