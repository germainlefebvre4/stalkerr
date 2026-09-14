## REMOVED Requirements

### Requirement: Configurable list of M3U sources
**Reason**: `m3u.sources` no longer needs to be a non-empty list for configuration to load — replaced by "Configurable, optionally empty list of M3U sources", which relaxes this constraint and defers to the new `m3u-source-overrides` capability for the effective (file + runtime) source list.
**Migration**: No action required. An existing non-empty `m3u.sources` list continues to work exactly as before; only the previously-fatal empty/absent case now succeeds instead of failing to start.

#### Scenario: Sources list defines every M3U source
- **WHEN** a configuration file sets a non-empty `m3u.sources` list
- **THEN** the system SHALL use every entry in `m3u.sources`, each identified by its own `name`, `file_path`, and `download` settings

#### Scenario: Missing or empty sources list is rejected
- **WHEN** a configuration file omits `m3u.sources` or sets it to an empty list
- **THEN** the system SHALL fail to load configuration and SHALL report a clear error indicating that at least one M3U source must be configured

## ADDED Requirements

### Requirement: Configurable, optionally empty list of M3U sources
The system SHALL support an `m3u.sources` configuration list, where each entry has a unique `name` and its own `file_path` and `download` settings. `m3u.sources` MAY be empty or absent; the system SHALL NOT fail to load configuration on that basis. The effective list of M3U sources used by the rest of this specification is the one resolved per the `m3u-source-overrides` capability (file-defined sources plus any runtime overrides or additions), and MAY be empty.

#### Scenario: Sources list defines every M3U source
- **WHEN** a configuration file sets a non-empty `m3u.sources` list and no runtime sources are stored
- **THEN** the system SHALL use every entry in `m3u.sources`, each identified by its own `name`, `file_path`, and `download` settings

#### Scenario: Missing or empty sources list is accepted
- **WHEN** a configuration file omits `m3u.sources` or sets it to an empty list, and no runtime sources are stored
- **THEN** the system SHALL load configuration successfully and SHALL proceed with an empty effective source list, rather than failing to start
