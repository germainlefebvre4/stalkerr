# processing-run-statistics Specification

## Purpose

Computes and persists per-run statistics on each M3U processing run's `processing_logs` entry, so consumers such as the Home dashboard can show what the last run actually did without recomputing it from item-level data.

## Requirements

### Requirement: Per-Run Content Statistics Persisted
For every M3U processing run, the system SHALL persist on that run's `processing_logs` entry, alongside the existing `item_count`: the number of movies processed by the run, the number of TV shows processed by the run, and the number of newly created items (playlist entries that did not exist before this run, as opposed to existing entries updated by this run). These fields SHALL be persisted when the run's `processing_logs` entry is finalized (whether the run completes successfully or fails partway through), reflecting only the items processed up to that point.

#### Scenario: Movies and TV shows counts reflect the run's content mix
- **WHEN** a processing run processes 12 items classified as movies and 5 items classified as TV shows
- **THEN** the run's `processing_logs` entry SHALL record a movies count of `12` and a TV shows count of `5`

#### Scenario: New items are distinguished from updated items
- **WHEN** a processing run creates 8 playlist entries that did not previously exist and updates 3 playlist entries that already existed (e.g. a forced re-process)
- **THEN** the run's `processing_logs` entry SHALL record a new-items count of `8`, not `11`

#### Scenario: A run that fails partway still persists partial statistics
- **WHEN** a processing run processes 6 items before encountering a fatal error and being marked `failed`
- **THEN** the run's `processing_logs` entry SHALL record the movies, TV shows, and new-items counts accumulated from those 6 items, not zero or null values

### Requirement: Per-Run TMDB Match Statistics Persisted
For every M3U processing run, the system SHALL persist on that run's `processing_logs` entry the number of items the run successfully matched to a TMDB entry (TMDB matched count) and the number of items the run attempted to match but could not (TMDB unmatched count).

#### Scenario: Matched and unmatched counts reflect TMDB lookup outcomes
- **WHEN** a processing run attempts TMDB matching for 20 items, successfully matching 17 and failing to find a match for 3
- **THEN** the run's `processing_logs` entry SHALL record a TMDB matched count of `17` and a TMDB unmatched count of `3`

#### Scenario: Items skipped from TMDB matching are excluded from both counts
- **WHEN** a processing run processes items with TMDB matching disabled or skipped (e.g. channels, or the `SkipTMDB` option)
- **THEN** those items SHALL NOT be counted in either the TMDB matched or TMDB unmatched count for that run

### Requirement: Per-Run Group Title List Persisted
For every M3U processing run, the system SHALL persist on that run's `processing_logs` entry the distinct list of `group_title` values across all items the run processed, with no duplicate values.

#### Scenario: Group title list is deduplicated
- **WHEN** a processing run processes 15 items across which the `group_title` "ACTION-FR" appears 6 times and "ANIMATION" appears 9 times
- **THEN** the run's `processing_logs` entry SHALL record a group title list containing exactly `["ACTION-FR", "ANIMATION"]` (in either order), not 15 entries

#### Scenario: A run with no processed items records an empty group title list
- **WHEN** a processing run completes having processed zero items (e.g. every line was a duplicate and skipped)
- **THEN** the run's `processing_logs` entry SHALL record an empty group title list, not a null or missing value

### Requirement: Runs Predating This Capability Report Absent Statistics
`processing_logs` entries created before this capability was introduced SHALL NOT have the new statistics fields populated. Consumers reading these fields for such an entry SHALL be able to distinguish "not available" (pre-existing entry) from "zero" (a run that legitimately processed nothing).

#### Scenario: A pre-existing run reports its statistics as absent, not zero
- **WHEN** a client reads the new statistics fields of a `processing_logs` entry created before this capability existed
- **THEN** the system SHALL represent those fields as absent/null rather than as `0` or an empty list, so a consumer can tell the difference from a run that genuinely processed nothing
