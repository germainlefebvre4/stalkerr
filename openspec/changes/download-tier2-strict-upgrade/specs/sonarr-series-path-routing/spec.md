## MODIFIED Requirements

### Requirement: Sonarr series download path routing
When downloading a TV show episode automatically — whether it is reported missing by Sonarr (tier-1) or it is an already-downloaded episode with a strictly-better untried candidate (tier-2) — the system SHALL derive the full destination directory from `series.Path` as returned by the Sonarr API, which encodes the root folder chosen in Sonarr's Media Management. For a tier-2 download, `series.Path` SHALL be obtained through a live lookup independent of the missing/wanted list, so the same series always resolves to the same destination directory regardless of which tier triggered the download.

#### Scenario: Series in primary root folder
- **WHEN** `series.Path` is `/downloads/sonarr/Breaking Bad`
- **THEN** the episode is downloaded to `/downloads/sonarr/Breaking Bad/Season 01/Breaking Bad - S01E01`

#### Scenario: Series in secondary root folder
- **WHEN** `series.Path` is `/downloads/sonarr-bis/Malcolm in the Middle`
- **THEN** the episode is downloaded to `/downloads/sonarr-bis/Malcolm in the Middle/Season 01/Malcolm in the Middle - S01E01`

#### Scenario: Empty series path fallback
- **WHEN** `series.Path` is empty
- **THEN** the system SHALL fall back to constructing the path from `cfg.Downloads.TVShowsPath` and `series.Title` (previous behavior)

#### Scenario: Tier-2 upgrade resolves the same destination as the original download
- **WHEN** a tier-2 stream is built for a series-season whose earlier tier-1 download used `series.Path` of `/downloads/sonarr/Breaking Bad`
- **THEN** the tier-2 download SHALL resolve to the same `/downloads/sonarr/Breaking Bad/Season 01/...` destination directory, so it replaces the previously downloaded episode file instead of creating a duplicate elsewhere

#### Scenario: Tier-2 live lookup finds no matching series
- **WHEN** a tier-2 stream is built for a series-season and the live Sonarr lookup returns no matching series
- **THEN** the system SHALL fall back to constructing the path from `cfg.Downloads.TVShowsPath` and the series' title (the same fallback used for tier-1)

### Requirement: Radarr movie download path routing
When downloading a movie automatically — whether it is reported missing by Radarr (tier-1) or it is an already-downloaded movie with a strictly-better untried candidate (tier-2) — the system SHALL derive the full destination directory from `movie.Path` as returned by the Radarr API. For a tier-2 download, `movie.Path` SHALL be obtained through a live lookup independent of the missing/wanted list, so the same movie always resolves to the same destination directory regardless of which tier triggered the download.

#### Scenario: Movie in primary root folder
- **WHEN** `movie.Path` is `/downloads/radarr/The Matrix (1999)`
- **THEN** the movie is downloaded to `/downloads/radarr/The Matrix (1999)/The Matrix (1999)`

#### Scenario: Movie in secondary root folder
- **WHEN** `movie.Path` is `/downloads/radarr-4k/Inception (2010)`
- **THEN** the movie is downloaded to `/downloads/radarr-4k/Inception (2010)/Inception (2010)`

#### Scenario: Empty movie path fallback
- **WHEN** `movie.Path` is empty
- **THEN** the system SHALL fall back to constructing the path from `cfg.Downloads.MoviesPath` and `movie.Title` (previous behavior)

#### Scenario: Tier-2 upgrade resolves the same destination as the original download
- **WHEN** a tier-2 stream is built for a movie whose earlier tier-1 download used `movie.Path` of `/downloads/radarr/The Matrix (1999)`
- **THEN** the tier-2 download SHALL resolve to the same `/downloads/radarr/The Matrix (1999)/The Matrix (1999)` destination, so it replaces the previously downloaded file instead of creating a duplicate elsewhere

#### Scenario: Tier-2 live lookup finds no matching movie
- **WHEN** a tier-2 stream is built for a movie and the live Radarr lookup returns no matching movie
- **THEN** the system SHALL fall back to constructing the path from `cfg.Downloads.MoviesPath` and the movie's title (the same fallback used for tier-1)
