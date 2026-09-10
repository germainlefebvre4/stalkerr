## MODIFIED Requirements

### Requirement: Database Pruning of Expired M3U Streams
The system SHALL support pruning `processed_lines` from the database that are no longer present in their corresponding configured source's currently downloaded M3U file. The `db-prune` command SHALL evaluate every configured source independently within a single run, pruning each source's stale `processed_lines` against that source's own current file content.

#### Scenario: Soft Pruning (Preserve Download History)
- **WHEN** the `db-prune` command is run without the `--hard` flag
- **THEN** for each configured source, the system deletes all `processed_lines` tagged with that source whose `line_hash` is not in that source's current M3U file, **except** those in `StateDownloaded` or `StateDownloading` states.

#### Scenario: Hard Pruning (Full Database Reset)
- **WHEN** the `db-prune` command is run with the `--hard` flag
- **THEN** for each configured source, the system deletes all `processed_lines` tagged with that source whose `line_hash` is not in that source's current M3U file, including those in `StateDownloaded` or `StateDownloading` states.

#### Scenario: Dry Run Mode
- **WHEN** the `db-prune` command is run with the `--dry-run` flag
- **THEN** the system calculates and prints, across all configured sources, the number of lines and orphaned metadata records that would be deleted, without making any modifications to the database.

#### Scenario: Multiple sources pruned independently
- **WHEN** `db-prune` runs with two configured sources, and one source's current file no longer contains a line that the other source's file still contains with the same `line_hash`
- **THEN** the system SHALL prune the stale `processed_line` tagged with the first source, and SHALL NOT prune the second source's `processed_line` with the same `line_hash`
