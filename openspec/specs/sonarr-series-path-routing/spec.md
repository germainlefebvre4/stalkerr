# Spec: sonarr-series-path-routing

## Purpose

Define how Stalkeer routes download destinations for TV show episodes and movies by using the `series.Path` / `movie.Path` values returned by the Sonarr and Radarr APIs, which encode the root folder chosen in each application's Media Management settings.

## Requirements

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
- **THEN** the system SHALL fall back to constructing the path from `cfg.Downloads.MoviesPath` and the movie's title (previous behavior)

#### Scenario: Tier-2 upgrade resolves the same destination as the original download
- **WHEN** a tier-2 stream is built for a movie whose earlier tier-1 download used `movie.Path` of `/downloads/radarr/The Matrix (1999)`
- **THEN** the tier-2 download SHALL resolve to the same `/downloads/radarr/The Matrix (1999)/The Matrix (1999)` destination, so it replaces the previously downloaded file instead of creating a duplicate elsewhere

#### Scenario: Tier-2 live lookup finds no matching movie
- **WHEN** a tier-2 stream is built for a movie and the live Radarr lookup returns no matching movie
- **THEN** the system SHALL fall back to constructing the path from `cfg.Downloads.MoviesPath` and the movie's title (the same fallback used for tier-1)

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
During each run of the scheduled `download` command, the system SHALL fetch the current Sonarr library once (a single `GetAllSeries` call, distinct from the run's existing missing-episode fetch) and compare the series root directory implied by each eligible completed TV episode download's stored `download_path` against that series' current `Path`, and SHALL update `download_path` when they differ, preserving the `Season NN/<file>` segment below the root exactly as it already was. This comparison SHALL NOT require any Sonarr API call beyond that single per-run library fetch. The comparison SHALL be insensitive to a trailing path separator on either side, so that a series' current `Path` differing from the stored root only by a trailing separator is treated as unchanged, not as a drift requiring correction.

When the stored root already matches the series' current `Path` (no directory correction needed), the system SHALL additionally verify that the file still exists at the stored `download_path`. When it does not, the system SHALL leave `download_path` unchanged but report a distinct "file missing" outcome for that download (surfaced the same way a `rename_failed` outcome is today) instead of treating it as fully reconciled.

#### Scenario: Series root renamed since the original download
- **WHEN** a completed episode's `download_path` is `/media/tvshows/Series Name (2020)/Season 01/Series Name (2020) - S01E01.mkv` and the matching series' current Sonarr `Path` is `/media/tvshows/Series Name (2020) - Corrected`
- **THEN** the scheduled run SHALL update `download_path` to `/media/tvshows/Series Name (2020) - Corrected/Season 01/Series Name (2020) - S01E01.mkv`

#### Scenario: Series root unchanged
- **WHEN** a completed episode's stored root directory already matches the series' current Sonarr `Path`, and the file still exists at the stored `download_path`
- **THEN** the scheduled run SHALL leave `download_path` unchanged and report no issue for that download

#### Scenario: Series root unchanged but file missing
- **WHEN** a completed episode's stored root directory already matches the series' current Sonarr `Path`, but no file exists at the stored `download_path`
- **THEN** the scheduled run SHALL leave `download_path` unchanged and report a distinct "file missing" outcome for that download, so the discrepancy is surfaced instead of silently treated as up to date

#### Scenario: Series path differs only by a trailing separator
- **WHEN** a completed episode's `download_path` is `/media/tvshows/Series Name (2020)/Season 01/Series Name (2020) - S01E01.mkv` and the matching series' current Sonarr `Path` is `/media/tvshows/Series Name (2020)/` (identical root, with a trailing separator)
- **THEN** the scheduled run SHALL treat the root as unchanged, leave `download_path` unchanged, and SHALL NOT attempt to move the file or log a correction-failure warning (the file-existence check above still applies)

### Requirement: Scheduled Correction of Completed Radarr Movie Paths
During each run of the scheduled `download` command, the system SHALL fetch the current Radarr library once (a single `GetAllMovies` call, distinct from the run's existing missing-movie fetch) and compare the movie root directory implied by each eligible completed movie download's stored `download_path` against that movie's current `Path`, and SHALL update `download_path` when they differ, preserving the file name below the root exactly as it already was. This comparison SHALL NOT require any Radarr API call beyond that single per-run library fetch. The comparison SHALL be insensitive to a trailing path separator on either side, so that a movie's current `Path` differing from the stored root only by a trailing separator is treated as unchanged, not as a drift requiring correction.

When the stored root already matches the movie's current `Path` (no directory correction needed), the system SHALL additionally verify that the file still exists at the stored `download_path`. When it does not, the system SHALL leave `download_path` unchanged but report a distinct "file missing" outcome for that download (surfaced the same way a `rename_failed` outcome is today) instead of treating it as fully reconciled.

#### Scenario: Movie root renamed since the original download
- **WHEN** a completed movie's `download_path` is `/media/movies/Movie Title (2019)/Movie Title (2019).mkv` and the matching movie's current Radarr `Path` is `/media/movies/Movie Title (2019) - Corrected`
- **THEN** the scheduled run SHALL update `download_path` to `/media/movies/Movie Title (2019) - Corrected/Movie Title (2019).mkv`

#### Scenario: Movie root unchanged
- **WHEN** a completed movie's stored root directory already matches the movie's current Radarr `Path`, and the file still exists at the stored `download_path`
- **THEN** the scheduled run SHALL leave `download_path` unchanged and report no issue for that download

#### Scenario: Movie root unchanged but file missing
- **WHEN** a completed movie's stored root directory already matches the movie's current Radarr `Path`, but no file exists at the stored `download_path`
- **THEN** the scheduled run SHALL leave `download_path` unchanged and report a distinct "file missing" outcome for that download, so the discrepancy is surfaced instead of silently treated as up to date

#### Scenario: Movie path differs only by a trailing separator
- **WHEN** a completed movie's `download_path` is `/media/movies/Movie Title (2019)/Movie Title (2019).mkv` and the matching movie's current Radarr `Path` is `/media/movies/Movie Title (2019)/` (identical root, with a trailing separator)
- **THEN** the scheduled run SHALL treat the root as unchanged, leave `download_path` unchanged, and SHALL NOT attempt to move the file or log a correction-failure warning (the file-existence check above still applies)

### Requirement: On-Demand Correction of a Single Completed Download's Path
The system SHALL support correcting a single completed download's stored path against Radarr's/Sonarr's current data immediately, independent of the scheduled `download` run, by performing a live lookup of the associated movie/series and applying the same root-correction logic as the scheduled reconciliation, including the same trailing-separator-insensitive comparison and the same file-existence verification when the root already matches.

#### Scenario: On-demand lookup finds a divergent path and corrects it
- **WHEN** an on-demand resync is requested for a completed download whose eligible movie/series has a current Radarr/Sonarr `Path` different from the download's stored root
- **THEN** the system SHALL update that download's `download_path` to reflect the current root, preserving its existing sub-path segment

#### Scenario: On-demand lookup finds the path already correct
- **WHEN** an on-demand resync is requested for a completed download whose stored root already matches the current Radarr/Sonarr `Path`, and the file still exists at the stored `download_path`
- **THEN** the system SHALL leave `download_path` unchanged and report that no correction was needed

#### Scenario: On-demand lookup finds the path already correct but the file missing
- **WHEN** an on-demand resync is requested for a completed download whose stored root already matches the current Radarr/Sonarr `Path`, but no file exists at the stored `download_path`
- **THEN** the system SHALL leave `download_path` unchanged and report a distinct "file missing" outcome, rather than reporting that no correction was needed

#### Scenario: On-demand lookup finds no Radarr/Sonarr match
- **WHEN** an on-demand resync is requested for a completed download whose movie/series is not found in the user's current Radarr/Sonarr library
- **THEN** the system SHALL leave `download_path` unchanged and report that the item is not managed by Radarr/Sonarr

#### Scenario: On-demand path differs only by a trailing separator
- **WHEN** an on-demand resync is requested for a completed download whose current Radarr/Sonarr `Path` differs from the stored root only by a trailing separator
- **THEN** the system SHALL leave `download_path` unchanged and report that no correction was needed, rather than reporting a rename failure or collision (the file-existence check above still applies)

### Requirement: Filename Title Sourced From the Same Live Lookup as the Destination Directory
For any automatically-triggered download — tier-1 (missing content), tier-2 (quality upgrade), or the incomplete-download resume mechanism — the title text used to compose the destination *filename* SHALL be sourced from the same live Radarr/Sonarr lookup used to compose the destination *directory* (`movie.Title`/`series.Title` as returned by the Radarr/Sonarr API for the matched movie/series), not from the locally stored `Movie.TMDBTitle`/`TVShow.TMDBTitle` database field. This applies regardless of which language TMDB metadata is configured in locally, so a movie or series with a differently-localized title in the local database does not produce a differently-named file from one download to the next.

#### Scenario: Tier-2 upgrade names the file using the live Radarr title
- **WHEN** a tier-2 upgrade is downloaded for a movie whose live Radarr `Title` is `"Star Wars"` but whose local `Movie.TMDBTitle` is `"La Guerre des étoiles"`
- **THEN** the upgraded file's name SHALL use `"Star Wars"`, matching the filename the original tier-1 download would have produced

#### Scenario: Incomplete-download resume names the file using the live Sonarr title
- **WHEN** the incomplete-download resume mechanism resumes an episode whose live Sonarr `Title` is `"Star Wars: The Clone Wars"` but whose local `TVShow.TMDBTitle` is `"Star Wars : The Clone Wars"`
- **THEN** the resumed file's name SHALL use `"Star Wars: The Clone Wars"`, matching the filename a tier-1 download of the same episode would have produced

### Requirement: Incomplete-Download Resume Performs Its Own Live Lookup When Synthesizing a New Stream
When the incomplete-download resume mechanism must synthesize a new stream for a movie or series-season because no tier-1 or tier-2 stream already exists for it this run, it SHALL perform the same live Radarr/Sonarr lookup used by tier-1 and tier-2 to obtain the destination directory (`movie.Path`/`series.Path`) and the filename title, instead of falling back to a locally-constructed path built from `cfg.Downloads.MoviesPath`/`cfg.Downloads.TVShowsPath` and the local database title. The existing empty-path fallback (used when tier-1/tier-2 also cannot find a live match) SHALL still apply when the live lookup itself finds no matching movie/series.

#### Scenario: Resume synthesis finds a live match
- **WHEN** the incomplete-download resume mechanism synthesizes a new stream for a movie that has no tier-1/tier-2 stream this run, and a live Radarr lookup for that movie succeeds
- **THEN** the synthesized stream's destination directory and filename title SHALL be derived from that live lookup's `movie.Path` and `movie.Title`, the same as a tier-1 stream for that movie would use

#### Scenario: Resume synthesis finds no live match
- **WHEN** the incomplete-download resume mechanism synthesizes a new stream for a movie that has no tier-1/tier-2 stream this run, and a live Radarr lookup for that movie finds no match
- **THEN** the synthesized stream SHALL fall back to constructing the path from `cfg.Downloads.MoviesPath` and the local database title, the same fallback already used elsewhere when no live match is found
