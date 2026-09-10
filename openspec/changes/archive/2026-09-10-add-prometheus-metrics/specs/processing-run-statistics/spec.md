## ADDED Requirements

### Requirement: Per-Run Metadata Backfill Statistics Persisted
For every M3U processing run, the system SHALL persist on that run's `processing_logs` entry the number of items whose metadata was successfully backfilled by the run's rich-metadata backfill step, and the number of items on which that step encountered an error. These fields SHALL be persisted when the run's `processing_logs` entry is finalized, following the same "absent, not zero, on pre-existing rows" rule already established by this capability's other statistics fields.

#### Scenario: Backfill counts reflect the step's outcome
- **WHEN** a processing run's rich-metadata backfill step successfully updates 5 items and encounters an error on 1 item
- **THEN** the run's `processing_logs` entry SHALL record a metadata-backfilled count of `5` and a metadata-backfill-errors count of `1`

#### Scenario: Runs predating this requirement report the fields as absent
- **WHEN** a client reads the metadata-backfill fields of a `processing_logs` entry created before this requirement was introduced
- **THEN** the system SHALL represent those fields as absent/null rather than as `0`, so a consumer can tell the difference from a run whose backfill step genuinely updated nothing
