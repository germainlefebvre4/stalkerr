## Purpose

Persist the rich TMDB metadata (poster, synopsis, and external identifiers) already returned by existing TMDB API calls on `Movie` and `TVShow` records, so downstream consumers can display visuals and deep links without any additional TMDB request.

## ADDED Requirements

### Requirement: Rich TMDB fields on Movie and TVShow
The `Movie` and `TVShow` GORM models SHALL include nullable `poster_path` (string), `overview` (text), and `imdb_id` (string) columns, in addition to the existing `tmdb_id` and `tvdb_id` columns.

#### Scenario: Schema includes new columns after migration
- **WHEN** the application runs its database auto-migration
- **THEN** the `movies` and `tv_shows` tables SHALL contain the columns `poster_path`, `overview`, and `imdb_id`, nullable, with no data loss to existing rows

### Requirement: Automatic ingestion enrichment persists rich metadata
When the automatic TMDB enrichment performed during M3U ingestion (`stalkeer process`) successfully retrieves movie or TV show details and external IDs from TMDB for a given item, the system SHALL persist `poster_path`, `overview`, `imdb_id`, and `tvdb_id` on the corresponding `Movie`/`TVShow` record, in addition to the fields already persisted (`tmdb_id`, `tmdb_title`, `tmdb_year`, `tmdb_genres`, and `duration` or `season`/`episode`). No additional TMDB API call SHALL be made to obtain these fields, since the detail and external-ID responses are already fetched by the existing enrichment flow.

#### Scenario: Movie ingested with full TMDB match
- **WHEN** a playlist entry is classified as a movie during `stalkeer process` and TMDB returns a matching movie with a poster, an overview, and an IMDB ID
- **THEN** the persisted `Movie` record SHALL include `poster_path`, `overview`, and `imdb_id` alongside the previously persisted fields

#### Scenario: TV show ingested with full TMDB match
- **WHEN** a playlist entry is classified as a TV show during `stalkeer process` and TMDB returns a matching show with a poster, an overview, and external IDs
- **THEN** the persisted `TVShow` record SHALL include `poster_path`, `overview`, `imdb_id`, and `tvdb_id` alongside the previously persisted fields

### Requirement: Manual override persists rich metadata
When a manual override (`POST /api/v1/items/:id/override`) resolves a `Movie` or `TVShow` record via a TMDB detail and external-ID lookup, the system SHALL persist `poster_path`, `overview`, `imdb_id`, and `tvdb_id` on that record, using the same detail/external-ID responses already fetched for the override — no additional TMDB API call SHALL be made.

#### Scenario: Forcing an association persists poster and overview
- **WHEN** a client calls `POST /api/v1/items/42/override` with a valid `tmdb_id` and `type`, and the TMDB detail lookup succeeds
- **THEN** the resulting `Movie` or `TVShow` record SHALL be created or updated with `poster_path`, `overview`, `imdb_id`, and `tvdb_id` populated from that lookup

### Requirement: API responses expose rich metadata fields
`MovieResponse` and `TVShowResponse` SHALL include `poster_path`, `overview`, `imdb_id`, and `tvdb_id`. Fields SHALL be omitted or null in the JSON response when the underlying value has not yet been populated on the record.

#### Scenario: Fetching an item with fully enriched metadata
- **WHEN** a client fetches a playlist item whose associated `Movie`/`TVShow` has `poster_path`, `overview`, `imdb_id`, and `tvdb_id` populated
- **THEN** the API response SHALL include all four fields with their stored values

#### Scenario: Fetching an item with not-yet-backfilled metadata
- **WHEN** a client fetches a playlist item whose associated `Movie`/`TVShow` has `tmdb_id` set but `poster_path`, `overview`, `imdb_id`, or `tvdb_id` still null
- **THEN** the API response SHALL return null (or omit) those specific fields without error
