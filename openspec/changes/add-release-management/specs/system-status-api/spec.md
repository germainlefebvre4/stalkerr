## ADDED Requirements

### Requirement: Build version metadata in status response
The aggregated status response SHALL include the running instance's version, commit, and build date, sourced from the embedded build metadata.

#### Scenario: Status response includes build metadata
- **WHEN** the status endpoint is called
- **THEN** the response SHALL include the version, commit, and date of the running build
