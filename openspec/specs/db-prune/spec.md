# db-prune Specification

## Purpose
TBD - created by archiving change m3u-url-reset-and-db-pruning. Update Purpose after archive.
## Requirements
### Requirement: Database Pruning of Expired M3U Streams
The system SHALL support pruning `processed_lines` from the database that are no longer present in the currently active M3U playlist file.

#### Scenario: Soft Pruning (Preserve Download History)
- **WHEN** the `db-prune` command is run without the `--hard` flag
- **THEN** the system deletes all `processed_lines` whose `line_hash` is not in the active M3U file, **except** those in `StateDownloaded` or `StateDownloading` states.

#### Scenario: Hard Pruning (Full Database Reset)
- **WHEN** the `db-prune` command is run with the `--hard` flag
- **THEN** the system deletes all `processed_lines` whose `line_hash` is not in the active M3U file, including those in `StateDownloaded` or `StateDownloading` states.

#### Scenario: Dry Run Mode
- **WHEN** the `db-prune` command is run with the `--dry-run` flag
- **THEN** the system calculates and prints the number of lines and orphaned metadata records that would be deleted, without making any modifications to the database.

### Requirement: Orphaned Metadata Cleanup
Following the pruning of processed lines, the system SHALL automatically delete orphaned `movies` and `tvshows` that no longer have any associated `processed_lines`.

#### Scenario: Delete movies and tvshows without processed lines
- **WHEN** the `db-prune` command is executed successfully (excluding dry-run)
- **THEN** the system deletes all records from `movies` and `tvshows` that are not referenced by any remaining `processed_lines`.

