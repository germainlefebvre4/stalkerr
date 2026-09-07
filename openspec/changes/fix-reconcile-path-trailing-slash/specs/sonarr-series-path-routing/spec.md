## MODIFIED Requirements

### Requirement: Scheduled Correction of Completed Sonarr Series Paths
During each run of the scheduled `download` command, the system SHALL fetch the current Sonarr library once (a single `GetAllSeries` call, distinct from the run's existing missing-episode fetch) and compare the series root directory implied by each eligible completed TV episode download's stored `download_path` against that series' current `Path`, and SHALL update `download_path` when they differ, preserving the `Season NN/<file>` segment below the root exactly as it already was. This comparison SHALL NOT require any Sonarr API call beyond that single per-run library fetch. The comparison SHALL be insensitive to a trailing path separator on either side, so that a series' current `Path` differing from the stored root only by a trailing separator is treated as unchanged, not as a drift requiring correction.

#### Scenario: Series root renamed since the original download
- **WHEN** a completed episode's `download_path` is `/media/tvshows/Series Name (2020)/Season 01/Series Name (2020) - S01E01.mkv` and the matching series' current Sonarr `Path` is `/media/tvshows/Series Name (2020) - Corrected`
- **THEN** the scheduled run SHALL update `download_path` to `/media/tvshows/Series Name (2020) - Corrected/Season 01/Series Name (2020) - S01E01.mkv`

#### Scenario: Series root unchanged
- **WHEN** a completed episode's stored root directory already matches the series' current Sonarr `Path`
- **THEN** the scheduled run SHALL leave `download_path` unchanged

#### Scenario: Series path differs only by a trailing separator
- **WHEN** a completed episode's `download_path` is `/media/tvshows/Series Name (2020)/Season 01/Series Name (2020) - S01E01.mkv` and the matching series' current Sonarr `Path` is `/media/tvshows/Series Name (2020)/` (identical root, with a trailing separator)
- **THEN** the scheduled run SHALL treat the root as unchanged, leave `download_path` unchanged, and SHALL NOT attempt to move the file or log a correction-failure warning

### Requirement: Scheduled Correction of Completed Radarr Movie Paths
During each run of the scheduled `download` command, the system SHALL fetch the current Radarr library once (a single `GetAllMovies` call, distinct from the run's existing missing-movie fetch) and compare the movie root directory implied by each eligible completed movie download's stored `download_path` against that movie's current `Path`, and SHALL update `download_path` when they differ, preserving the file name below the root exactly as it already was. This comparison SHALL NOT require any Radarr API call beyond that single per-run library fetch. The comparison SHALL be insensitive to a trailing path separator on either side, so that a movie's current `Path` differing from the stored root only by a trailing separator is treated as unchanged, not as a drift requiring correction.

#### Scenario: Movie root renamed since the original download
- **WHEN** a completed movie's `download_path` is `/media/movies/Movie Title (2019)/Movie Title (2019).mkv` and the matching movie's current Radarr `Path` is `/media/movies/Movie Title (2019) - Corrected`
- **THEN** the scheduled run SHALL update `download_path` to `/media/movies/Movie Title (2019) - Corrected/Movie Title (2019).mkv`

#### Scenario: Movie root unchanged
- **WHEN** a completed movie's stored root directory already matches the movie's current Radarr `Path`
- **THEN** the scheduled run SHALL leave `download_path` unchanged

#### Scenario: Movie path differs only by a trailing separator
- **WHEN** a completed movie's `download_path` is `/media/movies/Movie Title (2019)/Movie Title (2019).mkv` and the matching movie's current Radarr `Path` is `/media/movies/Movie Title (2019)/` (identical root, with a trailing separator)
- **THEN** the scheduled run SHALL treat the root as unchanged, leave `download_path` unchanged, and SHALL NOT attempt to move the file or log a correction-failure warning

### Requirement: On-Demand Correction of a Single Completed Download's Path
The system SHALL support correcting a single completed download's stored path against Radarr's/Sonarr's current data immediately, independent of the scheduled `download` run, by performing a live lookup of the associated movie/series and applying the same root-correction logic as the scheduled reconciliation, including the same trailing-separator-insensitive comparison.

#### Scenario: On-demand lookup finds a divergent path and corrects it
- **WHEN** an on-demand resync is requested for a completed download whose eligible movie/series has a current Radarr/Sonarr `Path` different from the download's stored root
- **THEN** the system SHALL update that download's `download_path` to reflect the current root, preserving its existing sub-path segment

#### Scenario: On-demand lookup finds the path already correct
- **WHEN** an on-demand resync is requested for a completed download whose stored root already matches the current Radarr/Sonarr `Path`
- **THEN** the system SHALL leave `download_path` unchanged and report that no correction was needed

#### Scenario: On-demand lookup finds no Radarr/Sonarr match
- **WHEN** an on-demand resync is requested for a completed download whose movie/series is not found in the user's current Radarr/Sonarr library
- **THEN** the system SHALL leave `download_path` unchanged and report that the item is not managed by Radarr/Sonarr

#### Scenario: On-demand path differs only by a trailing separator
- **WHEN** an on-demand resync is requested for a completed download whose current Radarr/Sonarr `Path` differs from the stored root only by a trailing separator
- **THEN** the system SHALL leave `download_path` unchanged and report that no correction was needed, rather than reporting a rename failure or collision
