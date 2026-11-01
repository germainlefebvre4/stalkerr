# Spec Delta

## MODIFIED Requirements

### Requirement: Interrupted downloads resume within their own stream
An episode or movie with an incomplete or interrupted download SHALL be treated as the next item to attempt within its own stream (the series-season stream it belongs to, or its movie stream), rather than requiring a separate command invocation to resume it — unless its movie/series is confirmed to no longer be monitored in Radarr/Sonarr, or the occurrence itself has exhausted its retry budget or was manually cancelled (capability `cancel-download-occurrence`), in which case it SHALL NOT be resumed this run. Monitored status SHALL be checked using data already available to the run; when that status cannot be confirmed either way (no matching Radarr/Sonarr entry for the movie/series, or Radarr/Sonarr unreachable this run), the incomplete download SHALL still be resumed as before, since only a confirmed "no longer monitored" state is grounds for skipping it. A forced download (capability `force-download-media-occurrence`) is resumed as the next item of its stream like any other incomplete download, but its occurrence SHALL be the only candidate attempted for that item.

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

#### Scenario: Retry-exhausted incomplete download is not resumed
- **WHEN** an incomplete download's occurrence has reached its configured `max_retry_attempts` budget
- **THEN** the scheduler SHALL NOT include that occurrence in this run's stream set, even though its movie/series is still monitored

#### Scenario: Cancelled incomplete download is not resumed
- **WHEN** an incomplete download's occurrence has been manually cancelled
- **THEN** the scheduler SHALL NOT include that occurrence in this run's stream set, even though its movie/series is still monitored

#### Scenario: Forced incomplete download attempts only its own occurrence
- **WHEN** a movie's incomplete download is a pending forced download and the movie has other untried occurrences
- **THEN** the scheduler SHALL attempt only the forced occurrence for that item and SHALL NOT attempt the other occurrences in the same stream item

## ADDED Requirements

### Requirement: Non-forced incomplete downloads of already-downloaded media are cancelled
When an incomplete download that is not forced belongs to a movie or TV episode that already has a different occurrence in state `downloaded`, the scheduler SHALL NOT resume it and SHALL transition it to the terminal cancelled status, so it is not re-evaluated on every later run.

#### Scenario: Non-forced incomplete download with a downloaded sibling is cancelled
- **WHEN** an incomplete download that is not forced belongs to a movie that already has another occurrence in state `downloaded`
- **THEN** the scheduler SHALL NOT resume it and SHALL mark it and its occurrence cancelled

### Requirement: Forced incomplete downloads are exempt from the already-downloaded cancellation
An incomplete download that is forced (capability `force-download-media-occurrence`) SHALL NOT be cancelled because another occurrence of the same movie or TV episode is already `downloaded`, and SHALL remain eligible for resume, since forcing a sibling of already-downloaded media is the purpose of that capability.

#### Scenario: Forced incomplete download with a downloaded sibling is kept
- **WHEN** a pending forced download belongs to a movie that already has another occurrence in state `downloaded`
- **THEN** the scheduler SHALL NOT cancel it and SHALL include it in this run's stream set, subject to the monitored check

#### Scenario: Forced incomplete download of an unmonitored movie is still skipped
- **WHEN** a pending forced download belongs to a movie confirmed unmonitored in Radarr for this run
- **THEN** the scheduler SHALL NOT include it in this run's stream set, and SHALL NOT cancel it
