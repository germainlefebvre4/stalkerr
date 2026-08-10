## Purpose

Automatically complete missing rich TMDB metadata (poster, synopsis, external identifiers) on `Movie`/`TVShow` records that were created before this metadata was tracked, without requiring a separate manual command.

## ADDED Requirements

### Requirement: Automatic backfill runs on every `process` invocation
Each time the `stalkeer process` command runs, the system SHALL, as part of that run, query all `Movie` and `TVShow` records where `tmdb_id` is set but `poster_path` is still null, fetch their details and external IDs from TMDB, and persist `poster_path`, `overview`, `imdb_id`, and `tvdb_id`. This backfill SHALL run automatically and silently — no CLI flag SHALL be required or provided to trigger or skip it.

#### Scenario: Legacy movie backfilled during a process run
- **WHEN** `stalkeer process` runs and a `Movie` record exists with a non-null `tmdb_id` and a null `poster_path`
- **THEN** the system SHALL fetch that movie's details and external IDs from TMDB during the run and update the record with `poster_path`, `overview`, `imdb_id`, and `tvdb_id`

#### Scenario: No records need backfill
- **WHEN** `stalkeer process` runs and no `Movie`/`TVShow` record has a non-null `tmdb_id` with a null `poster_path`
- **THEN** the system SHALL perform only the lookup query, make no TMDB API calls for the backfill, and add no meaningful delay to the run

#### Scenario: TV show rows deduplicated by TMDB ID
- **WHEN** several `TVShow` rows needing backfill share the same `tmdb_id` (different seasons/episodes)
- **THEN** the system SHALL call the TMDB details/external-IDs endpoints only once per unique `tmdb_id` and update every matching row from that single response

#### Scenario: A single record's TMDB lookup fails
- **WHEN** the TMDB details or external-IDs lookup fails for one record during the backfill
- **THEN** the system SHALL log the failure, leave that record unchanged, continue backfilling the remaining records, and SHALL NOT fail the overall `process` command because of it

#### Scenario: TMDB integration disabled
- **WHEN** TMDB integration is disabled in configuration, or `stalkeer process` is run with `--skip-tmdb`
- **THEN** the backfill SHALL be skipped entirely for that run

### Requirement: Backfill respects the existing TMDB rate limiter
The backfill SHALL issue its TMDB requests through the same rate-limited TMDB client used elsewhere in the application, honoring the configured requests-per-second limit and `Retry-After` handling.

#### Scenario: Backfill requests are throttled
- **WHEN** the backfill needs to fetch details for multiple records in a single `process` run
- **THEN** the requests SHALL be spaced according to the configured TMDB rate limit, the same as any other TMDB API usage in the application
