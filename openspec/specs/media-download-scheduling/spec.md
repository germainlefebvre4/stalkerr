# Capability: Media Download Scheduling

## Purpose

Coordinates missing-content downloads from Radarr and Sonarr through a single shared worker pool, so overall download concurrency stays within the IPTV backend's limits while still finishing a show's season once started and preferring earlier seasons first.

## Requirements

### Requirement: Unified fetch across both catalogs
The `download` command SHALL fetch missing movies from Radarr and missing episodes from Sonarr within the same invocation, so a single scheduled run considers both catalogs together instead of requiring two separate invocations.

#### Scenario: Both catalogs fetched in one run
- **WHEN** the `download` command runs
- **THEN** it SHALL query both the configured Radarr instance for missing movies and the configured Sonarr instance for missing episodes before building any streams

### Requirement: Streams are the atomic unit of scheduling
The scheduler SHALL group downloadable items into atomic streams: exactly one stream per movie, and one stream per TV series representing that series' earliest season that still has at least one non-terminal (missing, incomplete, or resumable) episode. A stream SHALL NOT be split across more than one worker at a time.

#### Scenario: A movie is its own stream
- **WHEN** a missing movie is matched and has a downloadable candidate
- **THEN** it SHALL be represented as a single-item stream independent of any series stream

#### Scenario: A series contributes at most one claimable stream at a time
- **WHEN** a series has missing episodes in more than one season
- **THEN** only one stream for that series SHALL be claimable at any given time

### Requirement: Series seasons are attempted in ascending order
For a given TV series, the scheduler SHALL make only its earliest incomplete season available as a claimable stream. A later season of the same series SHALL NOT become claimable until every episode of the earlier season has reached a terminal state (downloaded or permanently failed after exhausting retries/candidates).

#### Scenario: Later season blocked while earlier season is incomplete
- **WHEN** a series has missing episodes in both season 1 and season 2
- **THEN** the season 2 stream SHALL NOT be claimable while season 1 still has a non-terminal episode

#### Scenario: Later season unlocked once earlier season completes
- **WHEN** every episode of season 1 for a series reaches a terminal state
- **THEN** the season 2 stream for that series SHALL become claimable

### Requirement: Episode order within a season stream
Within a series-season stream, episodes SHALL be attempted in ascending episode-number order.

#### Scenario: Episodes attempted lowest-numbered first
- **WHEN** a season stream contains missing episodes 3, 5, and 6
- **THEN** episode 3 SHALL be attempted before episode 5, and episode 5 before episode 6

### Requirement: A claimed stream is drained to completion
Once a worker claims a stream, it SHALL process every remaining item of that stream, in order, before releasing it back as available, regardless of whether other workers become idle in the meantime. A worker SHALL only return to the pool of claimable streams once its current stream has no remaining item to attempt.

#### Scenario: Worker does not abandon a season mid-way
- **WHEN** a worker has claimed a series-season stream with 3 remaining episodes and completes the first episode
- **THEN** the worker SHALL attempt the next remaining episode of that same stream rather than claiming a different stream, even if another worker is idle

### Requirement: Random selection among claimable tier-1 streams
When a worker becomes free and more than one tier-1 stream (see the tier-selection requirement) is claimable, the scheduler SHALL select the next stream to claim uniformly at random among the eligible, not-yet-claimed tier-1 streams, so consecutive runs do not deterministically process one service's catalog before the other or in a fixed alphabetical order.

#### Scenario: Selection is not source-ordered
- **WHEN** both movie streams and series-season streams are claimable
- **THEN** the scheduler SHALL NOT guarantee that all movie streams are claimed before all series streams, or vice versa

### Requirement: Configurable shared concurrency limit
The number of concurrently claimed streams SHALL be bounded by a single configurable limit shared across both Radarr- and Sonarr-originated streams. The limit SHALL NOT be applied independently per service.

#### Scenario: Limit applies across both services combined
- **WHEN** the configured limit is 2
- **THEN** at most 2 streams, drawn from either service in any combination, SHALL be in flight at the same time

### Requirement: Weighted tier selection between new and force content
The scheduler SHALL classify claimable streams into tier 1 (content not yet downloaded) and tier 2 (already-downloaded content eligible for forced re-download or quality upgrade). Each time a worker selects its next stream, it SHALL draw from tier 2 with a small, fixed, configurable probability whenever at least one tier-2 stream is claimable, and from tier 1 otherwise. Tier 2 SHALL NOT be selected only after tier 1 is exhausted, and SHALL NOT be selected on every draw.

#### Scenario: Tier 2 is drawn occasionally while tier 1 is non-empty
- **WHEN** tier-1 streams remain claimable for the entire run and at least one tier-2 stream is also claimable
- **THEN** over the course of the run, some tier-2 streams SHALL still be claimed rather than being deferred until tier 1 is empty

#### Scenario: No tier-2 candidates means only tier 1 is drawn
- **WHEN** no tier-2 stream is claimable
- **THEN** every draw SHALL select from tier 1

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

### Requirement: Stale lock cleanup before scheduling
At the start of a run, the scheduler SHALL clean up stale download locks left by a previous, abnormally-terminated run before constructing streams.

#### Scenario: Stale lock from a crashed run is cleared
- **WHEN** a download lock exists past its timeout because a previous run crashed without releasing it
- **THEN** the scheduler SHALL clear that lock before the corresponding item can be claimed by a stream in the current run

### Requirement: A work unit yields at most one claimable stream
For a given `SourceKey` (a single movie, or a single series-season), `BuildStreams` SHALL produce at most one stream per run. When a work unit qualifies for both tier-1 (still missing) and tier-2 (already-downloaded, upgrade-eligible) classification in the same run, the tier-1 stream SHALL take precedence and no tier-2 stream SHALL be built for that same `SourceKey`.

#### Scenario: Movie missing per Radarr but already marked downloaded locally
- **WHEN** a movie is still reported as missing by Radarr, and the local database already has a `downloaded` `ProcessedLine` for it plus another eligible candidate
- **THEN** `BuildStreams` SHALL produce only the tier-1 stream for that movie, and SHALL NOT also produce a tier-2 stream for it

#### Scenario: Series-season missing per Sonarr but already marked downloaded locally
- **WHEN** a series-season has at least one episode still reported as missing by Sonarr, and the local database already has a `downloaded` `ProcessedLine` for another episode of that same season plus an eligible candidate
- **THEN** `BuildStreams` SHALL produce only the tier-1 stream for that series-season, and SHALL NOT also produce a tier-2 stream for it

#### Scenario: No destination path is targeted by two independently-scheduled downloads
- **WHEN** streams are claimed and drained in the same run
- **THEN** no two claimable streams SHALL resolve to the same destination path, so a completed download is never silently overwritten by another download of the same work unit later in the same run

### Requirement: Tier-2 stream requires a strictly better candidate
A tier-2 stream SHALL only be built for a movie or series-season when its best untried candidate — ranked by the `m3u-quality-selection` ordering (language, then resolution) — is strictly better than the candidate already downloaded for that same movie or series-season. A leftover untried candidate that is equal to or worse than the already-downloaded candidate SHALL NOT cause a tier-2 stream to be built.

#### Scenario: Untried candidate with a worse language is not offered as an upgrade
- **WHEN** a movie's already-downloaded candidate has `language = "VF"` and its only untried candidate has `language = "VOSTFR"`
- **THEN** no tier-2 stream SHALL be built for that movie

#### Scenario: Untried candidate with a worse resolution in the same language is not offered as an upgrade
- **WHEN** a movie's already-downloaded candidate has `language = "VF"`, `resolution = "720p"` and its only untried candidate has `language = "VF"`, `resolution = "4K"`
- **THEN** no tier-2 stream SHALL be built for that movie, since `720p` already outranks `4K` in the resolution preference order

#### Scenario: Untried candidate with a strictly better language triggers an upgrade
- **WHEN** a movie's already-downloaded candidate has `language = "MULTI"` and an untried candidate has `language = "VF"`
- **THEN** a tier-2 stream SHALL be built for that movie, targeting the `VF` candidate

#### Scenario: Untried candidate with a strictly better resolution in the same language triggers an upgrade
- **WHEN** a movie's already-downloaded candidate has `language = "VF"`, `resolution = "4K"` and an untried candidate has `language = "VF"`, `resolution = "720p"`
- **THEN** a tier-2 stream SHALL be built for that movie, targeting the `720p` candidate

#### Scenario: A satisfied tier-2 upgrade replaces the original file rather than duplicating it
- **WHEN** a tier-2 stream is built and its download completes successfully
- **THEN** the resulting file SHALL be written to the same destination path as the movie's or episode's original download, replacing it, rather than being written to a different path alongside it
