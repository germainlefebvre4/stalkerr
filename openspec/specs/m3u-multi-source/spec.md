# m3u-multi-source Specification

## Purpose

Defines how the system configures, downloads, archives, and processes more than one M3U playlist source in a single run, keeping every source's streams available as redundant candidates while isolating failures and dedup scope per source.

## Requirements

### Requirement: Configurable list of M3U sources
The system SHALL support an `m3u.sources` configuration list, where each entry has a unique `name` and its own `file_path` and `download` settings. `m3u.sources` SHALL be required and MUST be a non-empty list; the system SHALL fail to load configuration with a clear error when `m3u.sources` is absent or empty.

#### Scenario: Sources list defines every M3U source
- **WHEN** a configuration file sets a non-empty `m3u.sources` list
- **THEN** the system SHALL use every entry in `m3u.sources`, each identified by its own `name`, `file_path`, and `download` settings

#### Scenario: Missing or empty sources list is rejected
- **WHEN** a configuration file omits `m3u.sources` or sets it to an empty list
- **THEN** the system SHALL fail to load configuration and SHALL report a clear error indicating that at least one M3U source must be configured

### Requirement: Per-source download and archive isolation
The `m3u-download` command SHALL attempt the download and archive step for every configured source independently within a single run. A failure downloading or archiving one source SHALL be logged and SHALL NOT prevent the remaining sources from being attempted. The command SHALL exit with a non-zero status if at least one source failed, but only after every configured source has been attempted. Each source SHALL download and archive into its own subdirectory (named after the source) so that two sources' files or archives never overwrite each other.

#### Scenario: One source fails, others still run
- **WHEN** `m3u-download` runs with two configured sources and the download for the first source fails (network error or invalid response)
- **THEN** the system SHALL log the failure for the first source, SHALL still attempt the download and archive for the second source, and SHALL exit with a non-zero status after both have been attempted

#### Scenario: Archives from different sources do not collide
- **WHEN** two configured sources are downloaded and archived in the same run
- **THEN** each source's downloaded file and archive copies SHALL be written under a path segment specific to that source's `name`

### Requirement: Per-source M3U processing and provenance tagging
The `process` command SHALL iterate every configured source and process each source's downloaded file independently within a single run. Every `ProcessedLine` created or updated during processing SHALL be tagged with the `name` of the source it was parsed from, in a `source_name` field. When a source's downloaded file is missing (for example, because a prior `m3u-download` run failed for that source), the system SHALL skip that source with a warning and SHALL still process the remaining sources.

#### Scenario: Processed lines are tagged with their source
- **WHEN** `process` runs with two configured sources, `provider-a` and `provider-b`
- **THEN** every `ProcessedLine` parsed from `provider-a`'s file SHALL have `source_name = "provider-a"`, and every `ProcessedLine` parsed from `provider-b`'s file SHALL have `source_name = "provider-b"`

#### Scenario: Missing source file does not abort the run
- **WHEN** `process` runs and one configured source has no downloaded file on disk
- **THEN** the system SHALL log a warning for that source, SHALL skip it, and SHALL still process every other configured source

### Requirement: Per-source dedup scoping
The uniqueness constraint used to detect duplicate `ProcessedLine` entries SHALL be scoped to the combination of `source_name` and the existing content hash, rather than the content hash alone. Two entries with an identical hash but different `source_name` SHALL both be retained as distinct `ProcessedLine` records.

#### Scenario: Identical-looking entries from two sources are both kept
- **WHEN** two different configured sources each produce an entry that would hash identically (same `tvg_name` and stream URL)
- **THEN** the system SHALL persist both entries as separate `ProcessedLine` records, one per source, rather than rejecting the second as a duplicate

#### Scenario: True duplicates within the same source are still deduplicated
- **WHEN** the same source's playlist file contains the same entry twice in one run
- **THEN** the system SHALL deduplicate within that source exactly as it does today

### Requirement: Redundant candidates across sources remain usable for download fallback
A movie or TV episode available from more than one source SHALL still be represented as a single content item (matched by existing TMDB-based matching), with one `ProcessedLine` candidate per source. The existing quality/language candidate ordering and download-fallback loop SHALL consider these candidates without treating `source_name` as a ranking factor.

#### Scenario: A second source's stream is usable when the first source's stream fails
- **WHEN** a movie has one `ProcessedLine` candidate from `provider-a` and another from `provider-b`, and the candidate download from `provider-a` fails
- **THEN** the existing download-fallback loop SHALL attempt the candidate from `provider-b` next, ordered the same way it would order two candidates from a single source
