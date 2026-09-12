# radarr-sonarr-resilience Specification

## Purpose

Protects the Radarr and Sonarr integrations from a downed or degraded instance by retrying only genuinely transient failures and failing fast once a service is confirmed unhealthy, instead of blocking every caller on a fresh timeout.

## Requirements

### Requirement: Circuit breaker protects Radarr and Sonarr calls
The system SHALL maintain one shared circuit breaker per external service (Radarr, Sonarr) that every call to that service's API passes through. The breaker SHALL track consecutive failures across calls for the lifetime of the process (or, for the `download` command, for the lifetime of that run) rather than resetting per individual request.

#### Scenario: Consecutive failures open the circuit
- **WHEN** calls to a service fail repeatedly in a row, reaching the configured failure threshold
- **THEN** the circuit breaker for that service transitions to open, and subsequent calls fail immediately without contacting the service

#### Scenario: Circuit stays closed during isolated failures
- **WHEN** a call to a service fails but is followed by a successful call before the failure threshold is reached
- **THEN** the circuit breaker remains closed and the failure count resets, so the breaker keeps allowing calls through

### Requirement: Open circuit recovers via a bounded probe
Once the configured timeout has elapsed since a circuit opened, the breaker SHALL allow a limited number of trial calls through (half-open) to test whether the service has recovered, closing the circuit again on success or reopening it on failure.

#### Scenario: Successful probe closes the circuit
- **WHEN** the circuit is open, the configured timeout has elapsed, and the next trial call to the service succeeds
- **THEN** the circuit breaker transitions to closed and subsequent calls are allowed through normally

#### Scenario: Failed probe reopens the circuit
- **WHEN** the circuit is half-open and the trial call to the service fails
- **THEN** the circuit breaker transitions back to open and the timeout period restarts

### Requirement: Only transient failures are retried
Before a call counts as a failure toward the circuit breaker, the Radarr and Sonarr clients SHALL retry it according to their configured retry policy only when the failure is transient (connection failure, timeout, HTTP 429, or a 5xx response). A non-transient failure (e.g. a 4xx response other than 429) SHALL NOT be retried.

#### Scenario: Transient failure is retried before failing
- **WHEN** a Radarr or Sonarr call fails with a timeout, connection error, HTTP 429, or a 5xx response, and retry attempts remain
- **THEN** the client retries the call according to its configured backoff before reporting failure

#### Scenario: Non-transient failure is not retried
- **WHEN** a Radarr or Sonarr call fails with a non-retryable error (e.g. HTTP 404 or 400)
- **THEN** the client reports the failure immediately without retrying, and the failure still counts toward the circuit breaker

### Requirement: Circuit breaker state is independent per service
Radarr and Sonarr SHALL each have their own circuit breaker instance, so the open/closed/half-open state of one SHALL NOT affect calls to the other.

#### Scenario: One service's outage does not affect the other
- **WHEN** Radarr's circuit breaker is open
- **THEN** calls to Sonarr are unaffected and continue to be attempted normally
