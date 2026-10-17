# Spec Delta

## ADDED Requirements

### Requirement: Monitoring endpoints tolerate slow Radarr and Sonarr responses
The Radarr/Sonarr monitoring endpoints (movie and series listings, per-item match detail, and stats) SHALL allow each live upstream call up to 30 seconds to complete, and the stats endpoint SHALL allow up to 60 seconds overall, since it queries both services under one budget. These budgets SHALL be independent of the shorter fail-fast budget used by forced-download and other single-lookup checks.

#### Scenario: Upstream answers slower than the fail-fast budget
- **WHEN** Radarr or Sonarr takes longer than 10 seconds but less than 30 seconds to return its catalog
- **THEN** the listing endpoint SHALL return the normal successful response instead of an upstream-unavailable error

#### Scenario: Upstream exceeds the monitoring budget
- **WHEN** Radarr or Sonarr does not answer within 30 seconds
- **THEN** the listing endpoint SHALL return the same upstream-unavailable error as for any other connectivity failure

#### Scenario: Stats combine both services within one budget
- **WHEN** the stats endpoint queries Radarr and then Sonarr and the two calls together take longer than 30 seconds but less than 60 seconds
- **THEN** the endpoint SHALL return both sections normally rather than reporting the second service as unreachable

#### Scenario: Fail-fast checks keep their shorter budget
- **WHEN** a forced download, resync, or duplicate-handling check performs its live Radarr/Sonarr lookup
- **THEN** it SHALL still give up after the shorter fail-fast budget, unaffected by the monitoring endpoints' budgets

#### Scenario: Server does not cut off a slow successful response
- **WHEN** a monitoring endpoint takes longer than 15 seconds to finish within its allowed budget
- **THEN** the HTTP server SHALL still deliver the completed response to the client

### Requirement: Upstream fetch failures are logged with their cause
When a monitoring listing endpoint cannot fetch the Radarr movie catalog or the Sonarr series catalog, the system SHALL write a log entry that includes the underlying error, so an operator can distinguish a timeout, a refused connection, and an open circuit breaker. The client-facing error response SHALL stay the generic upstream-unavailable response.

#### Scenario: Catalog fetch times out
- **WHEN** the Radarr or Sonarr catalog fetch for a listing request fails because the upstream did not answer in time
- **THEN** a log entry SHALL record the failure together with its underlying error, and the client SHALL still receive the generic upstream-unavailable response

#### Scenario: Catalog fetch rejected by an open circuit
- **WHEN** the catalog fetch fails because the service's circuit breaker is open
- **THEN** the log entry SHALL make that cause identifiable, distinct from a timeout
