## Context

See `proposal.md` - Why. Relevant current-state facts that shape this design:

- `internal/api/force_download.go` (`forceDownloadItem`) already persists a `DownloadInfo{Status: pending}` row and links it to the target `ProcessedLine` via `persistForceDownloadPath` *before* it kicks off the transfer. Only the immediately-following goroutine (`go s.downloader.Download(...)`) needs to be removed for the deferred behavior; the durable state this change relies on is already written today.
- `internal/scheduler/build.go`'s `mergeIncompleteDownloads` already resumes any `DownloadInfo` in status `pending`/`downloading`/`paused`/`failed`/`retrying` that isn't currently locked (`internal/downloader/state_manager.go` `GetIncompleteDownloads`), as part of the `media-download-scheduling` capability's "Interrupted downloads resume within their own stream" requirement. This was built for crash recovery, but a freshly-created `pending` `DownloadInfo` satisfies the same query with no code change.
- That same mechanism skips resuming a `DownloadInfo` whose movie/series is *confirmed* unmonitored in Radarr/Sonarr (`MonitoredMovieTMDBIDs`/`MonitoredSeriesTVDBIDs`). The existing force-download existence check (`resolveForceDownloadMoviePath`/`resolveForceDownloadEpisodePath`) only confirms the movie/series *exists*, not that it's monitored — so today's immediate-execution path bypasses that gate, but a deferred one would silently hit it.
- `radarr.Movie` and `sonarr.Series` (returned by the same `GetMovieByTMDBID`/`FindEpisodeByTVDBID` calls the existence check already makes) both carry a `Monitored bool` field, so confirming monitored status costs no additional external call.

## Goals / Non-Goals

**Goals:**
- Route the actual file transfer for a forced download through the same scheduled `download` run and worker pool as automatic downloads, instead of an unbounded in-request goroutine.
- Fail the request at click time, with a clear error, for the one case (unmonitored media) where deferring silently would drop the request with no feedback.

**Non-Goals:**
- Changing `internal/scheduler/build.go` or the resume/dedup mechanism itself — analysis below confirms it already does the right thing for this new source of `pending` records.
- Reducing the `download` cron's schedule or adding a way to trigger an off-cycle run. The accepted worst-case latency is the existing cron interval (`0 */2 * * *`, see `charts/stalkerr/values.yaml`).
- Any durable "this was force-downloaded" marker in the UI's general playlist/downloads list. The existing per-drawer ephemeral "queued" confirmation (local React state in `MediaOccurrenceDrawer.tsx`) is the only feedback in scope; it already reads correctly under the new semantics ("queued", not "started").

## Decisions

**Decision: Delete the goroutine, keep everything else in `forceDownloadItem` as-is.**
`persistForceDownloadPath` already writes the exact `DownloadInfo`/`ProcessedLine` state the resume mechanism expects. No new column, status value, or "force" marker is introduced — a force-originated pending download is indistinguishable, once persisted, from one left behind by a crash mid-transfer, and that's intentional: it lets the existing, already-tested resume path carry it with zero new scheduler code.
- *Alternative considered*: add an explicit `origin`/`forced` flag on `DownloadInfo` so `mergeIncompleteDownloads` could treat it specially (e.g., bypass the monitored gate). Rejected: no requirement calls for different treatment once the monitored check happens at click time (see next decision), so the extra column and branching would add complexity with no behavioral payoff.

**Decision: Gate on `Monitored` at click time instead of changing the resume mechanism.**
Two ways existed to prevent a force-downloaded-but-unmonitored item from being silently dropped: (a) check `Monitored` in the existence check and refuse the request, or (b) special-case force-originated records in `mergeIncompleteDownloads` to skip the monitored gate. (a) was chosen because it gives the user an immediate, actionable error ("not monitored") instead of a request that appears accepted but silently never completes; it also requires touching only `internal/api/force_download.go`, reusing data already fetched for the existence check.
- *Alternative considered*: (b), bypassing the monitored gate for forced downloads. Rejected: it would mean a forced download for unmonitored media keeps retrying indefinitely across cron runs with no path to success (Radarr/Sonarr won't organize an unmonitored item's file), which is a worse outcome than refusing it up front.

**Decision: No change to the API response shape or the frontend.**
The endpoint still returns `202 {status: "queued", processed_line_id}`, and the frontend's existing `queued` badge text ("✓ Téléchargement mis en file d'attente") already describes "queued", not "started" — it needed no wording change to remain accurate. The new refusal case (`not_monitored`) surfaces through the existing generic API-error-to-badge path used by every other refusal (`not_matched`, `already_downloaded`, `media_not_found`, etc.), so no new frontend branch is needed either.

## Risks / Trade-offs

- **[Risk] Worst-case latency jumps from "immediate" to up to one `download` cron interval (2h).** → Accepted explicitly for this change; no mitigation needed (see Non-Goals).
- **[Risk] A forced download's transfer no longer starts the instant the request is accepted, so a user could force-download the same occurrence twice before the first pass runs.** → Already handled by existing eligibility checks: `persistForceDownloadPath` reuses the existing `DownloadInfo` when `item.DownloadInfoID` is already set, and the state guard (`item.State == StateDownloading` -> 409) still exists for a transfer already in flight. No new duplicate-request risk is introduced beyond what the current code already guards against once a first request has been accepted (its `ProcessedLine.State` stays whatever it was pre-request, e.g. `processed`, until the cron run actually starts the transfer — so a second click before that point re-enters `persistForceDownloadPath`'s reuse branch rather than creating a second record).
- **[Risk] `resolveForceDownload*Path`'s existence+monitored check reflects Radarr/Sonarr state at click time, which can drift before the cron run picks it up up to ~2h later** (e.g., monitoring gets turned off in between). → Not a new risk: `mergeIncompleteDownloads` already re-checks monitored status at run time and will skip resuming it then, consistent with how any other incomplete download is treated.

## Migration Plan

No data migration. This is a behavior-only change to one request handler; rollback is a plain revert (re-adding the goroutine and removing the monitored check) with no cleanup needed, since `DownloadInfo` rows created under either version of the code are structurally identical.
