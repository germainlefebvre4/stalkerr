## ADDED Requirements

### Requirement: Configurable list of M3U sources
The system SHALL support an `m3u.sources` configuration list, where each entry has a unique `name` and its own `file_path` and `download` settings. `m3u.sources` SHALL be required and MUST be a non-empty list; the system SHALL fail to load configuration with a clear error when `m3u.sources` is absent or empty.

#### Scenario: Sources list defines every M3U source
- **WHEN** a configuration file sets a non-empty `m3u.sources` list
- **THEN** the system SHALL use every entry in `m3u.sources`, each identified by its own `name`, `file_path`, and `download` settings

#### Scenario: Missing or empty sources list is rejected
- **WHEN** a configuration file omits `m3u.sources` or sets it to an empty list
- **THEN** the system SHALL fail to load configuration and SHALL report a clear error indicating that at least one M3U source must be configured

### Requirement: Per-source M3U processing and provenance tagging
The `process` command SHALL iterate every configured source and process each source's downloaded file independently within a single run. Every `ProcessedLine` created or updated during processing SHALL be tagged with the `name` of the source it was parsed from, in a `source_name` field. When a source's downloaded file is missing (for example, because a prior `m3u-download` run failed for that source), the system SHALL skip that source with a warning and SHALL still process the remaining sources.

#### Scenario: Processed lines are tagged with their source
- **WHEN** `process` runs with two configured sources, `provider-a` and `provider-b`
- **THEN** every `ProcessedLine` parsed from `provider-a`'s file SHALL have `source_name = "provider-a"`, and every `ProcessedLine` parsed from `provider-b`'s file SHALL have `source_name = "provider-b"`

#### Scenario: Missing source file does not abort the run
- **WHEN** `process` runs and one configured source has no downloaded file on disk
- **THEN** the system SHALL log a warning for that source, SHALL skip it, and SHALL still process every other configured source

## REMOVED Requirements

### Requirement: Configurable list of M3U sources with legacy fallback
**Reason**: The singular `m3u.file_path` / `m3u.download.*` configuration and its implicit-source fallback are removed. `m3u.sources` is now the only supported configuration shape (see the new "Configurable list of M3U sources" requirement).
**Migration**: Replace a singular `m3u.file_path` / `m3u.download.*` block with a one-entry `m3u.sources` list, giving that entry a `name` (for example, `default`). Note that the effective download destination and archive directory will gain a `<name>/` subdirectory that the singular block did not have — existing archives should be moved into that subdirectory, or simply left to repopulate on the next `m3u-download` run.

### Requirement: Per-source processing and provenance tagging
**Reason**: This requirement's "or the single implicit legacy source" branch and its "Legacy single-source processing keeps its implicit tag" scenario no longer apply now that `m3u.sources` is required. Replaced by the new "Per-source M3U processing and provenance tagging" requirement above, scoped to the sources-only behavior.
**Migration**: No user-facing migration; every configured source (including a single one) is now processed the same way, tagged with its own configured `name` rather than an implicit legacy tag.
