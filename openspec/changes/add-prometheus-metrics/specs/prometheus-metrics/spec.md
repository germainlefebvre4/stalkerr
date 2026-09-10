## Purpose

Exposes a config-gated, Prometheus text-format `/metrics` endpoint on the server's admin port, surfacing job-run outcomes, download state, and TMDB circuit-breaker health for an operator's own external monitoring stack to scrape.

## ADDED Requirements

### Requirement: Metrics Exposition Configurable and Disabled by Default
The system SHALL accept `metrics.enabled` (boolean, default `false`), `metrics.port` (integer, default `8081`), and `metrics.path` (string, default `/metrics`) configuration values.

#### Scenario: Metrics disabled by default
- **WHEN** no `metrics` configuration is provided
- **THEN** the system does not open a listener for metrics exposition

#### Scenario: Metrics endpoint absent when disabled
- **WHEN** `metrics.enabled` is `false`
- **THEN** the configured metrics port is not listened on, and no path on it returns Prometheus-formatted output

#### Scenario: Metrics endpoint served when enabled
- **WHEN** `metrics.enabled` is `true`
- **THEN** a `GET` request to `metrics.path` on `metrics.port` SHALL return a `200 OK` response with a Prometheus text-exposition-format body

### Requirement: Job Run Status and Duration Exposed Per Action
When enabled, the metrics endpoint SHALL expose, for each recognized action (`process`, `download`, `m3u-download`, `resume-downloads`, `enrich-tvdb`, `backfill-metadata`), the status and duration of that action's most recent run — sourced from `processing_logs` for the first three actions and from `job_runs` (see the `job-run-history` capability) for the latter three — under one shared metric namespace regardless of source table.

#### Scenario: Last run status exposed
- **WHEN** the most recent `process` run completed with status `success`
- **THEN** the metrics endpoint exposes a metric reflecting that action's last run as successful

#### Scenario: Last run duration exposed
- **WHEN** the most recent `enrich-tvdb` run took 42 seconds to complete
- **THEN** the metrics endpoint exposes a duration of 42 seconds for that action

#### Scenario: Actions from both sources share one metric family
- **WHEN** a client queries metrics covering both a `processing_logs`-backed action (e.g. `process`) and a `job_runs`-backed action (e.g. `resume-downloads`)
- **THEN** both appear as different label values under the same metric name, not under separate, source-specific metric names

### Requirement: Job Run Item Counts Exposed as Both Cumulative and Snapshot
When enabled, the metrics endpoint SHALL expose, for each action, its processed/succeeded/matched/skipped item counts both as a cumulative total across all historical runs of that action and as a snapshot of only its most recent run.

#### Scenario: Cumulative count increases across runs
- **WHEN** action `process` completes an additional run that processed 12 new items, following prior runs that together processed 100 items
- **THEN** the cumulative metric for that action's processed-item count reflects `112`

#### Scenario: Snapshot reflects only the latest run
- **WHEN** action `process`'s most recent run processed 12 items, following a prior run that processed 40
- **THEN** the snapshot metric for that action's processed-item count reflects `12`, not `52`

### Requirement: Downloads Exposed by Status
When enabled, the metrics endpoint SHALL expose the current count of tracked downloads grouped by status, and the current total bytes downloaded across all tracked downloads.

#### Scenario: Download status counts reflect current state
- **WHEN** the download tracking table has 5 downloads with status `completed`, 2 with status `failed`, and 1 with status `in_progress`
- **THEN** the metrics endpoint exposes those three counts under their respective status label values

#### Scenario: Total downloaded bytes reflects current state, not a monotonic total
- **WHEN** the sum of tracked downloads' bytes decreases because previously tracked downloads are pruned from the database
- **THEN** the exposed metric value decreases accordingly rather than continuing to reflect a higher historical peak

### Requirement: Download Retry Distribution Exposed
When enabled, the metrics endpoint SHALL expose a bucketed distribution of tracked downloads by retry count.

#### Scenario: Retry buckets reflect current retry counts
- **WHEN** 10 downloads have never been retried, 3 have been retried once, and 1 has been retried 5 times
- **THEN** the metrics endpoint exposes bucketed counts consistent with that distribution

### Requirement: TMDB Circuit Breaker State Exposed
When enabled, and when TMDB integration is active, the metrics endpoint SHALL expose the TMDB client's circuit breaker current state (closed, open, or half-open) and current failure count.

#### Scenario: Circuit breaker state reflects current in-process state
- **WHEN** the TMDB client's circuit breaker has tripped to the open state after repeated failures
- **THEN** the metrics endpoint reports that breaker's state as open

#### Scenario: Circuit breaker metrics absent when TMDB is disabled
- **WHEN** TMDB integration is disabled in configuration and no TMDB client is constructed
- **THEN** the metrics endpoint SHALL omit the TMDB circuit breaker metrics rather than reporting a misleading closed state
