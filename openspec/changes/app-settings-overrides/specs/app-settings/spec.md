## Purpose

Defines the generic backend mechanism that lets any applicative configuration field (Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads, logging) be overridden at runtime from the application itself, with each field's effective value and origin resolved independently, and read-only exposure of the bootstrap configuration this mechanism does not cover.

## ADDED Requirements

### Requirement: Overridable Applicative Settings Fields
The system SHALL allow each applicative configuration field (Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads, and logging level/format fields) to be overridden independently via a settings-override API, without requiring any other field of the same section to also be overridden.

#### Scenario: Overriding one field leaves siblings untouched
- **WHEN** an override is set for `radarr.api_key` only
- **THEN** `radarr.url`, `radarr.enabled`, `radarr.sync_interval`, and `radarr.quality_profile_id` SHALL continue to resolve from their file/env/default values, unaffected by the `radarr.api_key` override

### Requirement: Effective Field Resolution
For each overridable field, the system SHALL resolve its effective value as the stored override if one exists, or the field's file/env/default-resolved value otherwise, and SHALL apply the effective value wherever that configuration field is used.

#### Scenario: No override stored
- **WHEN** a field has no stored override
- **THEN** the system SHALL use its existing file/env/default-resolved value

#### Scenario: Override stored
- **WHEN** a field has a stored override
- **THEN** the system SHALL use the stored override's value in place of the file/env/default-resolved value

### Requirement: Field Origin Exposure
The system SHALL expose, for each overridable field, its effective value together with a two-state origin: "interface" when a stored override is applied, or "config" when the file/env/default-resolved value is applied. The system SHALL NOT distinguish a file-provided value from an environment-provided value in this origin.

#### Scenario: Fetching an unoverridden field's origin
- **WHEN** a client requests an overridable field that has no stored override
- **THEN** the system SHALL report its origin as "config"

#### Scenario: Fetching an overridden field's origin
- **WHEN** a client requests an overridable field that has a stored override
- **THEN** the system SHALL report its origin as "interface"

### Requirement: Clearing a Field Override
The system SHALL allow removing a stored override for a field, after which the field's effective value SHALL revert to its file/env/default-resolved value.

#### Scenario: Clearing a stored override
- **WHEN** a stored override for `tmdb.api_key` is removed
- **THEN** the effective value of `tmdb.api_key` SHALL revert to its file/env/default-resolved value, and its origin SHALL report as "config"

### Requirement: Bootstrap Configuration Remains Read-Only
The system SHALL expose `database.host`, `database.port`, `database.user`, `database.dbname`, `database.sslmode`, `api.port`, `metrics.port`, and `metrics.enabled` for display with their origin, but SHALL NOT accept a stored override for any of these fields, since the system needs their file/env-resolved values before it can reach its database or bind its ports.

#### Scenario: Reading bootstrap configuration
- **WHEN** a client requests the bootstrap configuration
- **THEN** the system SHALL return the file/env/default-resolved value of each bootstrap field, always reporting its origin as "config"

#### Scenario: Attempting to override a bootstrap field
- **WHEN** a client attempts to store an override for `database.host`, `api.port`, `metrics.port`, or `metrics.enabled`
- **THEN** the system SHALL reject the request without storing any override

### Requirement: Boot Without Applicative Configuration
The system SHALL start successfully when none of the overridable applicative fields are set in `config.yml`, environment variables, or as stored overrides, provided the bootstrap configuration is valid.

#### Scenario: Starting with no applicative configuration
- **WHEN** `config.yml` and environment variables define only valid bootstrap configuration and no Radarr, Sonarr, TMDB, Jellyfin, Notifications, or Downloads configuration exists anywhere
- **THEN** the system SHALL start successfully, with every overridable applicative field resolving to its built-in default (or empty/disabled where no default exists) until a value is set via `config.yml`, an environment variable, or a stored override

### Requirement: Sensitive Field Masking
The system SHALL mask the value of sensitive fields (any API key, auth token, or password field) when returning them via the settings API, both for their stored override and their file/env/default-resolved value, returning only whether a value is set rather than the value itself.

#### Scenario: Reading a sensitive field with a value set
- **WHEN** a client requests the effective value of `radarr.api_key` and a non-empty value is resolved (from override or config)
- **THEN** the system SHALL indicate that a value is set without returning the raw value

#### Scenario: Reading a sensitive field with no value set
- **WHEN** a client requests the effective value of `notifications.ntfy.auth_token` and no value is resolved from any source
- **THEN** the system SHALL indicate that no value is set
