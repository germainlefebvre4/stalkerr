## Why

Today, clicking "Force Download" starts the file transfer immediately in a background goroutine inside the API request handler, bypassing the scheduler's shared worker pool and concurrency limit entirely. This lets an arbitrary number of forced downloads run concurrently alongside the scheduled `download` cron's own pool, with no shared throttling. Deferring the transfer to the next `download` cron run lets it flow through the existing scheduler (concurrency limit, stream/tier machinery, resume-on-crash handling) instead of running out-of-band.

## What Changes

- **BREAKING**: `force-download` no longer starts the file transfer synchronously in-process. It only validates eligibility, persists a `pending` `DownloadInfo` record, and returns; the actual transfer now happens on the next scheduled `download` cron run, picked up through the existing incomplete-download resume mechanism (`media-download-scheduling`'s "Interrupted downloads resume within their own stream" requirement) — no changes to that mechanism itself.
- The live Radarr/Sonarr existence check performed at request time is extended to also require the movie/series (or episode) to be `monitored`. A forced-download request for existing-but-unmonitored media is now refused at the click, instead of silently persisting a `DownloadInfo` that the next cron run's resume logic would skip anyway (that logic already excludes confirmed-unmonitored items).
- No change to response shape: the endpoint still returns `202 {status: "queued", processed_line_id}` once eligibility, existence, and monitored checks pass and the `DownloadInfo` is persisted.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `force-download-media-occurrence`: the "Forced download executes asynchronously" requirement changes from "runs in-process immediately" to "deferred to the next `download` cron run"; the existence-check requirement gains a monitored-status condition.

## Impact

- `internal/api/force_download.go`: remove the goroutine that calls `s.downloader.Download(...)` immediately after persisting the `DownloadInfo`; add a monitored check to `resolveForceDownloadMoviePath`/`resolveForceDownloadEpisodePath` using the `Monitored` field already present on the Radarr/Sonarr API responses these functions fetch (no extra external calls).
- No changes to `internal/scheduler/build.go` or `internal/downloader/state_manager.go`: the existing `mergeIncompleteDownloads` resume/dedup mechanism already handles a `pending` `DownloadInfo` correctly (verified: it reuses an existing tier-1 stream for the same movie/series rather than duplicating it, and `attachResume` skips re-adding a candidate already present).
- Frontend (`frontend/src/components/MediaOccurrenceDrawer.tsx`, `frontend/src/locales/*/playlist.json`): no code change required — the existing "queued" confirmation badge shown immediately after a successful request already matches the new semantics ("queued", not "started"). A new `not_monitored` error case surfaces through the existing generic error-message path.
- Worst-case latency for a forced download becomes bounded by the `download` cron's schedule (every 2 hours, see `charts/stalkerr/values.yaml`), rather than starting immediately.
