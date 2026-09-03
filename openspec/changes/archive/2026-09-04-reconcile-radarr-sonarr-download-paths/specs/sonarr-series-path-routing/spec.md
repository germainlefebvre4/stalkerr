## ADDED Requirements

### Requirement: Reconciliation Scope Limited to Radarr/Sonarr-Matched Downloads
A completed download SHALL be eligible for path reconciliation (scheduled or on-demand) only when its associated movie or TV show is currently present in the user's own Radarr or Sonarr library, established by matching against the movie's/series' TMDB or TVDB id in Radarr's/Sonarr's own data. A stored `tvdb_id` on the `Movie`/`TVShow` record alone SHALL NOT be treated as sufficient evidence of a Radarr/Sonarr match, since that id may originate from TMDB's external-id mapping for content the user's Radarr/Sonarr instance does not manage. A download with no such match SHALL be left entirely unchanged by both the scheduled and on-demand reconciliation mechanisms.

#### Scenario: Movie present in Radarr library is eligible
- **WHEN** a completed movie download's `Movie.TMDBID` matches a movie currently in the user's Radarr library
- **THEN** the download SHALL be eligible for reconciliation against that movie's current `Path`

#### Scenario: TV episode present in Sonarr library is eligible
- **WHEN** a completed TV episode download's `TVShow.TMDBID`/`TVDBID` matches a series currently in the user's Sonarr library
- **THEN** the download SHALL be eligible for reconciliation against that series' current `Path`

#### Scenario: TMDB-only match is not eligible
- **WHEN** a completed download's `Movie`/`TVShow` record has a `tvdb_id` populated from TMDB's external-id lookup, but no movie/series with that id is currently found in the user's Radarr/Sonarr library
- **THEN** the download SHALL NOT be reconciled by either mechanism, and SHALL be left with its existing stored path unchanged

### Requirement: Scheduled Correction of Completed Sonarr Series Paths
During each run of the scheduled `download` command, the system SHALL fetch the current Sonarr library once (a single `GetAllMonitoredSeries` call, distinct from the run's existing missing-episode fetch) and compare the series root directory implied by each eligible completed TV episode download's stored `download_path` against that series' current `Path`, and SHALL update `download_path` when they differ, preserving the `Season NN/<file>` segment below the root exactly as it already was. This comparison SHALL NOT require any Sonarr API call beyond that single per-run library fetch.

#### Scenario: Series root renamed since the original download
- **WHEN** a completed episode's `download_path` is `/media/tvshows/Series Name (2020)/Season 01/Series Name (2020) - S01E01.mkv` and the matching series' current Sonarr `Path` is `/media/tvshows/Series Name (2020) - Corrected`
- **THEN** the scheduled run SHALL update `download_path` to `/media/tvshows/Series Name (2020) - Corrected/Season 01/Series Name (2020) - S01E01.mkv`

#### Scenario: Series root unchanged
- **WHEN** a completed episode's stored root directory already matches the series' current Sonarr `Path`
- **THEN** the scheduled run SHALL leave `download_path` unchanged

### Requirement: Scheduled Correction of Completed Radarr Movie Paths
During each run of the scheduled `download` command, the system SHALL fetch the current Radarr library once (a single `GetAllMovies` call, distinct from the run's existing missing-movie fetch) and compare the movie root directory implied by each eligible completed movie download's stored `download_path` against that movie's current `Path`, and SHALL update `download_path` when they differ, preserving the file name below the root exactly as it already was. This comparison SHALL NOT require any Radarr API call beyond that single per-run library fetch.

#### Scenario: Movie root renamed since the original download
- **WHEN** a completed movie's `download_path` is `/media/movies/Movie Title (2019)/Movie Title (2019).mkv` and the matching movie's current Radarr `Path` is `/media/movies/Movie Title (2019) - Corrected`
- **THEN** the scheduled run SHALL update `download_path` to `/media/movies/Movie Title (2019) - Corrected/Movie Title (2019).mkv`

#### Scenario: Movie root unchanged
- **WHEN** a completed movie's stored root directory already matches the movie's current Radarr `Path`
- **THEN** the scheduled run SHALL leave `download_path` unchanged

### Requirement: On-Demand Correction of a Single Completed Download's Path
The system SHALL support correcting a single completed download's stored path against Radarr's/Sonarr's current data immediately, independent of the scheduled `download` run, by performing a live lookup of the associated movie/series and applying the same root-correction logic as the scheduled reconciliation.

#### Scenario: On-demand lookup finds a divergent path and corrects it
- **WHEN** an on-demand resync is requested for a completed download whose eligible movie/series has a current Radarr/Sonarr `Path` different from the download's stored root
- **THEN** the system SHALL update that download's `download_path` to reflect the current root, preserving its existing sub-path segment

#### Scenario: On-demand lookup finds the path already correct
- **WHEN** an on-demand resync is requested for a completed download whose stored root already matches the current Radarr/Sonarr `Path`
- **THEN** the system SHALL leave `download_path` unchanged and report that no correction was needed

#### Scenario: On-demand lookup finds no Radarr/Sonarr match
- **WHEN** an on-demand resync is requested for a completed download whose movie/series is not found in the user's current Radarr/Sonarr library
- **THEN** the system SHALL leave `download_path` unchanged and report that the item is not managed by Radarr/Sonarr
