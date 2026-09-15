## Purpose

Provides an on-demand backend check that validates whether a caller-supplied Radarr, Sonarr, or Jellyfin URL and API key combination is reachable and authenticates, without persisting anything or depending on previously saved settings.

## ADDED Requirements

### Requirement: On-Demand Connectivity Test Endpoint
The system SHALL expose an endpoint that accepts a target service identifier (Radarr, Sonarr, or Jellyfin) and connection values (base URL and, optionally, an API key) supplied by the caller, and SHALL perform a live check against that service using exactly the supplied values, without reading or writing any stored setting for that service.

#### Scenario: Testing caller-supplied values
- **WHEN** the caller submits a service, a base URL, and an API key
- **THEN** the system performs a live reachability-and-authentication call to that URL using that key, not the currently saved configuration for that service

#### Scenario: Testing without an API key
- **WHEN** the caller omits the API key or submits it empty
- **THEN** the system SHALL still attempt the call using no credentials, rather than substituting the currently saved key for that service

### Requirement: Three-State Result Classification
The endpoint SHALL classify each check as exactly one of: OK (reachable and authenticated), or KO with one of the reasons unreachable, unauthorized, timeout, or unavailable — reusing the same classification already established by `system-status-api`.

#### Scenario: Successful check
- **WHEN** the target service responds successfully to the check using the supplied values
- **THEN** the endpoint reports OK

#### Scenario: Invalid credentials
- **WHEN** the target service is reachable but rejects the supplied API key with an authentication failure
- **THEN** the endpoint reports KO with reason "unauthorized", distinct from a reachability failure

#### Scenario: Unreachable target
- **WHEN** the supplied URL cannot be connected to (connection refused, DNS failure)
- **THEN** the endpoint reports KO with reason "unreachable"

#### Scenario: Timed out
- **WHEN** the check does not complete within its timeout
- **THEN** the endpoint reports KO with reason "timeout"

### Requirement: Input Validation
The endpoint SHALL require a non-empty base URL and SHALL reject a request missing one without attempting any outbound call. The endpoint SHALL support only Radarr, Sonarr, and Jellyfin as a target service; a request naming any other service (including TMDB) SHALL be rejected as invalid without attempting any outbound call.

#### Scenario: Missing URL
- **WHEN** the caller submits a request with an empty base URL
- **THEN** the system SHALL reject the request without attempting a network call

#### Scenario: Unsupported service
- **WHEN** the caller submits "tmdb" (or any service other than radarr, sonarr, or jellyfin) as the target
- **THEN** the system SHALL reject the request without attempting any outbound call

### Requirement: Independent of the Shared Circuit Breaker
For Radarr and Sonarr, the check SHALL bypass that service's shared circuit breaker: it SHALL always attempt a live call regardless of the breaker's current state, and the call's outcome SHALL NOT be recorded as a success or failure toward that breaker's state.

#### Scenario: Breaker open does not block the test
- **WHEN** a service's shared circuit breaker is currently open
- **THEN** the connectivity test still attempts a live call to that service rather than immediately reporting a circuit-open failure, and the call's outcome does not change the breaker's state

### Requirement: No Persistence
The endpoint SHALL NOT create, modify, or clear any stored settings override as a side effect of performing a check.

#### Scenario: Check does not affect saved configuration
- **WHEN** a check is performed with values that differ from the currently saved configuration for that service
- **THEN** the saved configuration for that service SHALL remain unchanged after the check completes
