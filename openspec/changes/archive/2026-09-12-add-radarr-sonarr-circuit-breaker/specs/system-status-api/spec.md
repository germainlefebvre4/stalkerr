## MODIFIED Requirements

### Requirement: Aggregated on-demand system status endpoint
The system SHALL expose an endpoint that returns, in a single response, the status of the database, Radarr, Sonarr, TMDB, and the app's configured storage paths. The endpoint SHALL compute every section fresh on each call; it SHALL NOT cache or persist results between calls - except that the Radarr and Sonarr sections SHALL report their circuit breaker's open state without a live call, as described in the three-state reachability requirement below.

#### Scenario: Successful aggregation
- **WHEN** the status endpoint is called
- **THEN** the response SHALL include a database section, a Radarr section, a Sonarr section, a TMDB section, and a disk usage section

#### Scenario: One dependency's failure does not affect the others
- **WHEN** one dependency (e.g. Radarr) is unreachable
- **THEN** the endpoint SHALL still return HTTP 200 with accurate status for the database, Sonarr, TMDB, and disk usage sections

### Requirement: Radarr/Sonarr/TMDB three-state reachability
Each of the Radarr, Sonarr, and TMDB sections SHALL report exactly one of three states: **OK** (reachable and responding correctly), **KO** with a short human-readable reason (e.g. unreachable, invalid credentials, timed out, circuit open), or **not configured** (the integration is disabled or missing required settings). A **not configured** integration SHALL NOT trigger any outbound call to that service. For Radarr and Sonarr specifically, when that service's shared circuit breaker is currently open, the section SHALL report **KO** with a reason indicating the circuit is open, without attempting a live reachability call to that service.

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

#### Scenario: Radarr or Sonarr circuit breaker is open
- **WHEN** the status endpoint is called and the shared circuit breaker for Radarr or Sonarr is currently open
- **THEN** the corresponding section SHALL report **KO** with a reason indicating the circuit is open, without making an outbound call to that service, and without waiting for the reachability timeout
