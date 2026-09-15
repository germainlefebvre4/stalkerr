## MODIFIED Requirements

### Requirement: Circuit breaker protects Radarr and Sonarr calls
The system SHALL maintain one shared circuit breaker per external service (Radarr, Sonarr) that every call to that service's API passes through, with one explicit exception: the on-demand connectivity-test check (`integration-connectivity-test`) SHALL bypass that breaker entirely — it always attempts a live call regardless of the breaker's state, and its outcome is never recorded against the breaker. The breaker SHALL track consecutive failures across calls for the lifetime of the process (or, for the `download` command, for the lifetime of that run) rather than resetting per individual request.

#### Scenario: Consecutive failures open the circuit
- **WHEN** calls to a service fail repeatedly in a row, reaching the configured failure threshold
- **THEN** the circuit breaker for that service transitions to open, and subsequent calls fail immediately without contacting the service

#### Scenario: Circuit stays closed during isolated failures
- **WHEN** a call to a service fails but is followed by a successful call before the failure threshold is reached
- **THEN** the circuit breaker remains closed and the failure count resets, so the breaker keeps allowing calls through

#### Scenario: Connectivity test bypasses the breaker
- **WHEN** the shared circuit breaker for a service is currently open
- **THEN** an on-demand connectivity test for that service still attempts a live call, and the test's outcome does not alter the breaker's open/closed state
