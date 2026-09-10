# system-status-api Specification

## Purpose

Expose a single, on-demand backend endpoint that reports database connectivity, Radarr/Sonarr/TMDB reachability, and deduplicated disk usage for the app's storage paths, so problems can be diagnosed without reading logs.

## Requirements

### Requirement: Aggregated on-demand system status endpoint
The system SHALL expose an endpoint that returns, in a single response, the status of the database, Radarr, Sonarr, TMDB, and the app's configured storage paths. The endpoint SHALL compute every section fresh on each call; it SHALL NOT cache or persist results between calls.

#### Scenario: Successful aggregation
- **WHEN** the status endpoint is called
- **THEN** the response SHALL include a database section, a Radarr section, a Sonarr section, a TMDB section, and a disk usage section

#### Scenario: One dependency's failure does not affect the others
- **WHEN** one dependency (e.g. Radarr) is unreachable
- **THEN** the endpoint SHALL still return HTTP 200 with accurate status for the database, Sonarr, TMDB, and disk usage sections

### Requirement: Database connectivity check
The database section SHALL report whether the database is currently reachable, reusing the existing database health check.

#### Scenario: Database reachable
- **WHEN** the database responds to the health check
- **THEN** the database section SHALL report an OK status

#### Scenario: Database unreachable
- **WHEN** the database does not respond to the health check
- **THEN** the database section SHALL report a KO status with a short, human-readable reason

### Requirement: Radarr/Sonarr/TMDB three-state reachability
Each of the Radarr, Sonarr, and TMDB sections SHALL report exactly one of three states: **OK** (reachable and responding correctly), **KO** with a short human-readable reason (e.g. unreachable, invalid credentials, timed out), or **not configured** (the integration is disabled or missing required settings). A **not configured** integration SHALL NOT trigger any outbound call to that service.

#### Scenario: Integration disabled
- **WHEN** Radarr, Sonarr, or TMDB is disabled (or missing required configuration) in the app configuration
- **THEN** the corresponding section SHALL report **not configured** without attempting to contact that service

#### Scenario: Integration configured and reachable
- **WHEN** an enabled integration responds successfully to its reachability check
- **THEN** the corresponding section SHALL report **OK**

#### Scenario: Integration configured but unreachable
- **WHEN** an enabled integration's reachability check fails to connect (e.g. connection refused, DNS failure)
- **THEN** the corresponding section SHALL report **KO** with a reason distinguishing "unreachable" from other failure kinds

#### Scenario: Integration configured but rejects credentials
- **WHEN** an enabled integration's reachability check returns an authentication/authorization failure (e.g. invalid API key)
- **THEN** the corresponding section SHALL report **KO** with a reason distinguishing "invalid credentials" from a plain connectivity failure

#### Scenario: Slow integration does not block the response
- **WHEN** an enabled integration does not respond within a bounded timeout
- **THEN** the corresponding section SHALL report **KO** with a reason indicating a timeout, and the overall endpoint response SHALL NOT be delayed beyond that bound waiting for it

### Requirement: Deduplicated disk usage per configured storage path
The disk usage section SHALL report available/used space for each of the app's configured storage paths (movies download path, TV shows download path, temp directory when set, and M3U archive directory). Paths that resolve to the same mounted volume SHALL be merged into a single reported entry rather than listed redundantly.

#### Scenario: Distinct volumes reported separately
- **WHEN** two configured storage paths reside on different mounted volumes
- **THEN** the disk usage section SHALL report a separate entry for each

#### Scenario: Paths sharing a volume are merged
- **WHEN** two or more configured storage paths reside on the same mounted volume
- **THEN** the disk usage section SHALL report a single entry for that volume, referencing all of the paths it backs

#### Scenario: Unset temp directory is omitted
- **WHEN** the temp directory is not configured (falls back to the OS default)
- **THEN** the disk usage section SHALL NOT include an entry for it

#### Scenario: Configured path is missing or unreadable
- **WHEN** a configured storage path does not exist or its usage cannot be read
- **THEN** the disk usage section SHALL report that entry as unavailable with a short reason, rather than failing the entire endpoint response
