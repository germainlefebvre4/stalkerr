## MODIFIED Requirements

### Requirement: download radarr uses TVDB ID as primary match key
The Radarr-fetch stage of the unified `download` command SHALL pass `movie.TvdbID` to `MatchMovieByTVDB` so the TVDB-primary key path is used when available, falling through to TMDB-ID and fuzzy title matching only when `TvdbID == 0`.

#### Scenario: Movie matched by TVDB ID when available
- **WHEN** a Radarr movie has `TvdbID = 12345` and a `Movie` record exists in the database with `tvdb_id = 12345`
- **THEN** the command SHALL match that record with confidence 100% via the TVDB-primary path

#### Scenario: Falls through to TMDB ID when TVDB ID is zero
- **WHEN** a Radarr movie has `TvdbID = 0` and a `Movie` record exists with the matching `tmdb_id`
- **THEN** the command SHALL match via the TMDB-ID path

#### Scenario: Falls through to fuzzy matching when both IDs miss
- **WHEN** a Radarr movie has `TvdbID = 0` and no `Movie` record matches its `TMDBID`
- **THEN** the command SHALL attempt fuzzy title+year matching as the final fallback
