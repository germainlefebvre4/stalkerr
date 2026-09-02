## Why

The IPTV backend used to fetch stream URLs cannot reliably sustain more than 2-3 simultaneous downloads (higher counts risk aborted transfers). Today, four independent CronJobs (`radarr-sync`, `radarr-sync-force`, `sonarr-sync`, `sonarr-sync-force`) each run a fully sequential, one-item-at-a-time download loop, with no coordination between the Radarr and Sonarr pipelines and no shared concurrency budget. This wastes available download slots, processes catalogs in a rigid alphabetical/API order, and offers no way to bound simultaneous downloads across both services at once. Consolidating both pipelines into a single scheduled process lets us enforce one shared concurrency limit while still respecting how a viewer actually consumes a show (earliest season first, finish a season once started) and without the crawl always going "all of Radarr, then all of Sonarr" in the same order.

## What Changes

- Introduce a single `stalkeer download` CLI command that fetches missing movies from Radarr and missing episodes from Sonarr in the same run, replacing the separate `stalkeer radarr` and `stalkeer sonarr` commands. **BREAKING**: `stalkeer radarr`, `stalkeer sonarr`, and the `sonarr --series-id` filter are removed; there is no direct replacement for targeting a single series from the CLI.
- Add a stream-based scheduler: work is grouped into atomic streams (one per movie, one per series' earliest incomplete season), and a configurable pool of N workers (`downloads.max_parallel`) each claim one stream at random and drain it fully (in episode order) before claiming another, so a season in progress is never abandoned mid-way and a series' seasons are always attempted in ascending order.
- Add a weighted tier system so previously-downloaded content eligible for re-download/upgrade (today's `--force` behavior) is folded into the same run instead of a separate cronjob: normal missing content is tier 1, force/upgrade candidates are tier 2, drawn with a small fixed probability so tier 2 is never fully starved but never crowds out new content either.
- Integrate incomplete/interrupted download resumption (currently only reachable via the standalone, unscheduled `stalkeer resume-downloads` command) directly into stream construction, so a partially-downloaded episode or movie is resumed as the next item in its own stream rather than left for a manual run.
- Replace the four Helm CronJob templates (`cronjob-radarr-sync.yaml`, `cronjob-radarr-sync-force.yaml`, `cronjob-sonarr-sync.yaml`, `cronjob-sonarr-sync-force.yaml`) with a single `cronjob-download.yaml`, carrying both the Radarr and Sonarr API key secrets and both media volume mounts.
- `stalkeer resume-downloads` remains available as a standalone manual/debug command; it is not removed, only no longer the only way to resume interrupted downloads.

## Capabilities

### New Capabilities
- `media-download-scheduling`: the stream-based scheduler that unifies Radarr and Sonarr missing-content downloads into one worker pool, with season-locality, ascending-season priority, weighted tier selection between new and force/upgrade content, and integrated resume of interrupted downloads.

### Modified Capabilities
- `scheduled-jobs`: removes the four `radarrSync`/`sonarrSync` (normal + force) CronJob requirements and replaces them with a single unified download CronJob requirement.
- `radarr-movie-matching`: the TVDB-primary matching requirement currently attributed to the `download radarr` command now applies to the Radarr-fetch stage of the unified `download` command; matching behavior itself is unchanged.
- `m3u-quality-selection`: the quality-candidate fallback loop currently attributed to `download radarr`/`download sonarr` now applies to the unified `download` command; the fallback behavior itself is unchanged.

## Impact

- **Code**: new `cmd/download.go` (replacing `cmd/radarr.go` and `cmd/sonarr.go`); new scheduler package (likely `internal/downloader` or a new `internal/scheduler`) implementing stream construction, the weighted tier draw, and per-stream worker draining on top of the existing `ParallelDownloader`; reuse of `internal/matcher`, `internal/external/radarr`, `internal/external/sonarr`, and `internal/downloader.StateManager`/`ResumeHelper`.
- **Helm chart**: `charts/stalkerr/templates/cronjob-radarr-sync.yaml`, `cronjob-radarr-sync-force.yaml`, `cronjob-sonarr-sync.yaml`, `cronjob-sonarr-sync-force.yaml` removed; new `cronjob-download.yaml` added; `values.yaml` `jobs.radarrSync`/`jobs.sonarrSync` keys replaced by a single `jobs.download` key.
- **Configuration**: `downloads.max_parallel` becomes the single shared concurrency budget across both services (previously each command read the same key independently but never ran concurrently with the other service).
- **Operational**: reduces the CronJob count for this pipeline from four to one; changes the observable download ordering (interleaved, partly randomized) from the current strictly sequential per-service behavior.
