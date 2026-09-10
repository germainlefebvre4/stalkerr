# job-run-history Specification

## Purpose

Persists a durable run history (status, duration, and outcome counts) for the `resume-downloads` and `enrich-tvdb` commands, whose results are today only written to logs and lost once the process exits.

## Requirements

### Requirement: Per-Invocation Job Run Persisted
For every invocation of `resume-downloads` or `enrich-tvdb`, the system SHALL persist a `job_runs` entry recording: the action name, status (`success`, `failed`, or `in_progress`), start time, completion time, and the number of items the invocation succeeded on, failed on, and skipped.

#### Scenario: A resume-downloads run records outcome counts
- **WHEN** a `resume-downloads` invocation resumes 8 downloads successfully, fails to resume 2, and skips 1 already-complete download, then exits normally
- **THEN** its `job_runs` entry records a succeeded count of `8`, a failed count of `2`, a skipped count of `1`, and status `success`

#### Scenario: A crash partway still persists partial counts
- **WHEN** an `enrich-tvdb` invocation processes 6 items before encountering a fatal error and exiting
- **THEN** its `job_runs` entry records status `failed`, the succeeded/failed/skipped counts accumulated from those 6 items (not zero or null), and a non-empty error message

#### Scenario: A running invocation is visible before completion
- **WHEN** a `resume-downloads` invocation is still running
- **THEN** its `job_runs` entry exists with status `in_progress` and no completion time, from the moment the invocation starts

### Requirement: Job Run History Independent of Metrics Exposition
The system SHALL persist `job_runs` entries regardless of whether Prometheus metrics exposition (see the `prometheus-metrics` capability) is enabled or disabled.

#### Scenario: History persists with metrics exposition disabled
- **WHEN** `metrics.enabled` is `false` or unset
- **THEN** a `resume-downloads` or `enrich-tvdb` invocation still creates and finalizes its `job_runs` entry exactly as it would with metrics exposition enabled

### Requirement: Job Run Duration Derivable
Each finalized `job_runs` entry SHALL record its start time and completion time, from which the run's duration can be derived.

#### Scenario: Duration computed from recorded timestamps
- **WHEN** a `job_runs` entry has completed
- **THEN** its duration equals its completion time minus its start time
