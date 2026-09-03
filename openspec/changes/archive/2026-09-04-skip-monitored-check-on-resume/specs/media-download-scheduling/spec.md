## MODIFIED Requirements

### Requirement: Interrupted downloads resume within their own stream
An episode or movie with an incomplete or interrupted download SHALL be treated as the next item to attempt within its own stream (the series-season stream it belongs to, or its movie stream), rather than requiring a separate command invocation to resume it — unless its movie/series is confirmed to no longer be monitored in Radarr/Sonarr, in which case it SHALL NOT be resumed this run. Monitored status SHALL be checked using data already available to the run; when that status cannot be confirmed either way (no matching Radarr/Sonarr entry for the movie/series, or Radarr/Sonarr unreachable this run), the incomplete download SHALL still be resumed as before, since only a confirmed "no longer monitored" state is grounds for skipping it.

#### Scenario: Partially-downloaded episode resumes as part of its season stream
- **WHEN** an episode in an in-progress season stream has an incomplete download from a previous run, and its series is still monitored in Sonarr
- **THEN** the scheduler SHALL attempt to resume that episode's download as part of processing that stream, without a separate resume command being invoked

#### Scenario: Incomplete movie download is not resumed once unmonitored
- **WHEN** a movie has an incomplete/interrupted download from a previous run, and that movie is confirmed unmonitored in Radarr for this run
- **THEN** the scheduler SHALL NOT include that movie's incomplete download in this run's stream set

#### Scenario: Incomplete series-episode download is not resumed once unmonitored
- **WHEN** a TV episode has an incomplete/interrupted download from a previous run, and its series is confirmed unmonitored in Sonarr for this run
- **THEN** the scheduler SHALL NOT include that episode's incomplete download in this run's stream set

#### Scenario: Monitored status cannot be confirmed
- **WHEN** an incomplete download's movie/series has no matching entry in Radarr/Sonarr this run, or Radarr/Sonarr cannot be reached to check
- **THEN** the scheduler SHALL still resume that incomplete download as it would have before this change
