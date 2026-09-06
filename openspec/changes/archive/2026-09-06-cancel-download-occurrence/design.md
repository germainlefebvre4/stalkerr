## Context

See proposal.md - Why. Two existing mechanisms are relevant and stay as-is:

- `DownloadInfo.Status` and `ProcessedLine.State` are plain `varchar` columns
  (GORM auto-migrate, no DB-level enum/CHECK constraint — confirmed via
  `init-db.sql`, which defers all table creation to GORM). Adding a new
  string value needs no schema migration.
- Both existing "what to attempt next" queries are **allow-lists**, not
  exclusion lists:
  - `matcher.FindMovieDownloadCandidates` / `FindTVShowDownloadCandidates`:
    `WHERE state IN ('processed', 'failed')`
  - `StateManager.GetIncompleteDownloads` (the resume path):
    `WHERE status IN ('pending', 'downloading', 'paused', 'failed', 'retrying')`

A new status value that is simply **absent** from both allow-lists is
therefore excluded from both paths automatically, with no new WHERE clause
needed anywhere.

## Goals / Non-Goals

**Goals:**
- One shared terminal status value reached either by exhausting the retry
  budget or by a manual cancel request.
- Retry accounting that counts uniformly across both attempt paths and
  regardless of whether the failure was internally "retryable".
- A minimal, synchronous cancel endpoint mirroring the existing
  force-download handler's shape.

**Non-Goals:**
- Interrupting an actively running transfer (`downloading` status). No
  `context.CancelFunc` registry is introduced; today's downloads run to
  completion or failure once started (`go func(){ ... }` with
  `context.Background()` in `force_download.go`, and the worker pool loop in
  `cmd/download.go`).
- A targeted "reactivate this one occurrence" action. Reversal stays the
  existing full `POST /movies/:id/reset` / `/tvshows/:id/reset`.
- Redesigning force-download into a queued/prioritized cron-driven flow (see
  proposal's discussion of the follow-up change).
- Any DB schema migration (see Context — none is needed).

## Decisions

### One shared status value, not two
Both triggers (retry exhaustion, manual cancel) transition the occurrence to
the same new status value, `cancelled`, on both `DownloadInfo.Status` and
`ProcessedLine.State`.

**Alternative considered**: two distinct values (`exhausted` vs `cancelled`).
Rejected — the two allow-list queries above would need to exclude both
values identically, so nothing behavioral is gained; the existing
`error_message` field already lets a curious user distinguish "last attempt
failed with X after N retries" from a message set at manual-cancel time (see
below), without a second status value to keep in sync everywhere status is
switched on (frontend badge, API filters, matcher).

### retry_count increments once per external attempt, at terminal failure
Move the `retry_count` increment out of `UpdateState`'s `retrying` case (fired
between in-call sub-attempts by `retry.Config.OnRetry`, and blind to whether
the error was retryable) and into the point where a whole `Download()` call
concludes without success — i.e. the existing `failed` branch in
`downloader.go` after `retry.Do` returns an error. This is a single-line move
in `state_manager.go`'s `UpdateState` (the `case models.DownloadStatusFailed`
branch gains `retry_count + 1`; the `case models.DownloadStatusRetrying`
branch drops it) plus removing the increment from the retrying case.

`retrying` still exists as a transient status for live UI progress (a
`downloading` item mid-backoff between sub-attempts still shows as
"retrying"); it just no longer double-counts against the budget.

Immediately after this increment, the same code path checks the new
`retry_count` against `cfg.Downloads.MaxRetryAttempts`; when reached, it sets
status to `cancelled` instead of `failed` (and mirrors that onto the linked
`ProcessedLine.State`) in the same update. This keeps the exhaustion
transition in one place, exercised identically whether the call originated
from the tier-1 matcher path or the resume path (both call the same
`Downloader.Download()` → `UpdateState`).

**Alternative considered**: check the budget in `downloadItem()`
(`cmd/download.go`) before each candidate attempt instead. Rejected — that
function has no equivalent on the resume path (`ResumeHelper`/
`mergeIncompleteDownloads` build jobs differently), so the check would need
duplicating; centralizing in `state_manager.go` covers both for free,
consistent with `download-transfer-integrity`'s existing "guard applies
uniformly to every path" precedent.

### Cancel is a thin, synchronous status transition
`POST /api/v1/downloads/:id/cancel` (id = `DownloadInfo.id`, mirroring the
`downloads` resource the rest of that path already addresses) loads the
`DownloadInfo` and its linked `ProcessedLine`(s), checks status is one of
`pending`/`failed`/`retrying` (else `409`), and — in one transaction — sets
`DownloadInfo.Status = "cancelled"` and each linked `ProcessedLine.State =
"cancelled"`. No goroutine, no external call, unlike `force_download.go`'s
handler: there is no I/O to wait on.

A cancel request against a `downloading` occurrence is refused by the same
status check with no extra locking logic needed: `UpdateState` already sets
`status = downloading` for the whole duration of an attempt (lock held,
released only at the end), so a request arriving mid-transfer simply sees
`downloading` and is refused — no race window requiring a separate lock
check.

### Manual cancel sets a distinguishing error message
To keep the two triggers distinguishable without a second status value, the
cancel handler sets `error_message` to a fixed, clearly-labeled string (e.g.
"Cancelled by user") when transitioning to `cancelled`. A retry-exhaustion
transition leaves whatever real failure message the last attempt already
set. The sidepanel's existing error-message section (gated on status, per
`downloads-details-sidepanel`) surfaces this without new UI plumbing beyond
recognizing the `cancelled` status.

## Risks / Trade-offs

- **[Risk]** Changing what `retry_count` counts (attempts, not sub-retries)
  changes the meaning of any existing dashboard/query built against the old
  semantic. → **Mitigation**: `download-url-monitoring`'s spec delta
  documents the new semantic; no other in-repo consumer depends on the old
  one (checked: only `GetIncompleteDownloads`'s budget comparison and the
  frontend's "failedWithRetry" label read it, both fine under either
  semantic).
- **[Risk]** Existing rows already stuck in `failed` with a low `retry_count`
  (like the Nobody (2021) case that prompted this change) are not
  retroactively migrated to `cancelled` — they would still be retried a few
  more times under the old semantic's leftover count before this fix's
  accounting takes over. → **Mitigation**: this is intentional, not a gap —
  it's exactly why the manual cancel action ships in the same change: the
  user can immediately cancel any already-stuck occurrence rather than
  waiting on the corrected counter to catch up.
- **[Risk]** Reusing one status value for two different root causes could
  make future analytics ("how many were manually cancelled vs. gave up on
  their own?") harder. → **Mitigation**: the distinguishing error message
  keeps that information recoverable if ever needed; a dedicated field can
  be added later without another status value if it becomes worth it.

## Migration Plan

No DB migration required (new string value on existing `varchar` columns,
confirmed no CHECK constraint exists). Single deploy: backend first (new
status value understood by matcher/resume/API), frontend picks up the new
status value and the "Annuler" action in the same release. No feature flag —
both sides degrade safely if deployed slightly out of order (old frontend
against new backend: an unrecognized `cancelled` status falls back to the
default badge styling; new frontend against old backend: the "Annuler"
button would 404 until the backend deploy lands, acceptable for a same-day
paired rollout).
