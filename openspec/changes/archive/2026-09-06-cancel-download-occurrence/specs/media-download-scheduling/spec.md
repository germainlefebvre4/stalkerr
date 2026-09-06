## MODIFIED Requirements

### Requirement: Interrupted downloads resume within their own stream
An episode or movie with an incomplete or interrupted download SHALL be treated as the next item to attempt within its own stream (the series-season stream it belongs to, or its movie stream), rather than requiring a separate command invocation to resume it — unless its movie/series is confirmed to no longer be monitored in Radarr/Sonarr, or the occurrence itself has exhausted its retry budget or was manually cancelled (capability `cancel-download-occurrence`), in which case it SHALL NOT be resumed this run. Monitored status SHALL be checked using data already available to the run; when that status cannot be confirmed either way (no matching Radarr/Sonarr entry for the movie/series, or Radarr/Sonarr unreachable this run), the incomplete download SHALL still be resumed as before, since only a confirmed "no longer monitored" state is grounds for skipping it.

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

## ADDED Requirements

### Requirement: Retry-exhausted or cancelled occurrences are excluded from missing-content candidate matching
A `ProcessedLine` occurrence whose retry budget (`max_retry_attempts`) has been exhausted, or that has been manually cancelled (capability `cancel-download-occurrence`), SHALL NOT be returned as a download candidate for its movie or TV-show episode by the missing-content matcher, regardless of whether that movie/episode is still reported missing by Radarr/Sonarr. This reaches the terminal state ("permanently failed after exhausting retries") already described by the "Series seasons are attempted in ascending order" requirement, which this exclusion enforces.

This exclusion applies identically whether the occurrence's most recent attempt came from the missing-content matcher or from the incomplete-download resume path: reaching the retry budget on either path excludes the occurrence from both going forward.

#### Scenario: Retry-exhausted candidate is not offered again
- **WHEN** a movie's only download candidate has reached its `max_retry_attempts` budget, and the movie is still reported missing by Radarr
- **THEN** the missing-content matcher SHALL NOT return that candidate, and no stream SHALL be built for that movie from it

#### Scenario: Cancelled candidate is not offered again
- **WHEN** a TV episode's only download candidate has been manually cancelled, and the episode is still reported missing by Sonarr
- **THEN** the missing-content matcher SHALL NOT return that candidate for that episode

#### Scenario: Sibling candidates remain eligible
- **WHEN** a movie has one retry-exhausted or cancelled candidate and another candidate that has not exhausted its retry budget and was not cancelled
- **THEN** the missing-content matcher SHALL still return the other, eligible candidate, unaffected by its sibling's exclusion

#### Scenario: A movie with only excluded candidates yields no stream
- **WHEN** every download candidate for a movie is either retry-exhausted or cancelled
- **THEN** no tier-1 stream SHALL be built for that movie this run, and it SHALL NOT be retried again until a full media reset (capability `media-reset`) makes new candidates eligible
