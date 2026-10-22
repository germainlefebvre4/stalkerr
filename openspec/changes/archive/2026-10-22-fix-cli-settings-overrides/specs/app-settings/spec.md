# Spec Delta

## MODIFIED Requirements

### Requirement: Effective Field Resolution
For each overridable field, the system SHALL resolve its effective value as the stored override if one exists, or the field's file/env/default-resolved value otherwise, and SHALL apply the effective value wherever that configuration field is used.

#### Scenario: No override stored
- **WHEN** a field has no stored override
- **THEN** the system SHALL use its existing file/env/default-resolved value

#### Scenario: Override stored
- **WHEN** a field has a stored override
- **THEN** the system SHALL use the stored override's value in place of the file/env/default-resolved value

#### Scenario: Override stored and used by a command-line command
- **WHEN** `radarr.api_key` has a stored override and a command-line command (for example `download`) builds its Radarr client
- **THEN** the command SHALL authenticate with the stored override's value, not with the file/env-provided value

## ADDED Requirements

### Requirement: Command-Line Commands Apply Stored Overrides
Every command-line command that reads an overridable applicative field SHALL resolve that field's effective value from stored overrides as well as file/env/default values, exactly as the API server does.

#### Scenario: Scheduled job uses a key corrected from the interface
- **WHEN** the Radarr API key in the Secret is outdated, a valid key was stored as an override from the interface, and the `download` command runs
- **THEN** the command SHALL call Radarr with the stored key and SHALL NOT fail with an authentication error caused by the outdated Secret value

#### Scenario: Overridden log level applies to a command
- **WHEN** `logging.app.level` has a stored override and a command-line command starts
- **THEN** the command SHALL log at the overridden level once the database is connected

#### Scenario: Overridden M3U sources apply to a command
- **WHEN** the M3U source list has a stored override and a command-line command iterates M3U sources
- **THEN** the command SHALL use the effective source list, not only the file-configured one

### Requirement: Unreadable Overrides Are Reported
The system SHALL report, with a warning that names the cause, any resolution of the effective configuration that could not read stored overrides because the database was unavailable, and SHALL then fall back to file/env/default values. It SHALL NOT silently ignore stored overrides.

#### Scenario: Effective configuration resolved without a database connection
- **WHEN** the effective configuration is resolved while no database connection exists
- **THEN** the system SHALL log a warning stating that stored overrides were not applied

#### Scenario: Command that does not require the database cannot reach it
- **WHEN** a command that works without the database (for example `cleanup`) runs and the database is unreachable
- **THEN** the command SHALL log a warning that overrides were not applied and SHALL continue with file/env/default values

#### Scenario: Command that requires the database cannot reach it
- **WHEN** a command that requires the database (for example `download`) runs and the database is unreachable
- **THEN** the command SHALL exit with an error and SHALL NOT proceed using file/env values alone
