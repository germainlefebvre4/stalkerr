## Why

A download occurrence whose source link is permanently dead (e.g. a truncated
transfer reported as `unexpected EOF`) is retried forever, roughly every
cronjob cycle. Two gaps combine to cause this: `retry_count` is only
incremented for a narrow class of "retryable" transport errors used to drive
in-call sub-retries, so a hard failure barely moves the counter; and neither
the normal missing-content matcher nor the incomplete-download resume path
actually checks any retry budget before selecting a candidate to attempt
again. `media-download-scheduling` already documents a "permanently failed
after exhausting retries" terminal state as if it were enforced — it isn't.
On top of that, there is no way for a user to stop a stuck occurrence today
short of the existing full media reset, which erases the entire movie/show's
download history (and, since it frees up the line's unique hash, can let the
exact same dead line come back on the next playlist ingest).

## What Changes

- Every attempt to download a specific occurrence — whether triggered by the
  normal missing-content matcher or by the incomplete-download resume path —
  counts once against that occurrence's retry budget (`max_retry_attempts`),
  regardless of whether the underlying failure was itself classified as
  retryable for in-call sub-retries.
- Once an occurrence's retry budget is exhausted, it is permanently excluded
  from future candidate selection (both the matcher's candidate queries and
  the resume path), reaching the terminal state `media-download-scheduling`
  already describes but that isn't currently enforced.
- A new manual "cancel" action lets a user immediately move one specific
  occurrence into that same terminal excluded state, without waiting for its
  retry budget to exhaust naturally, and without affecting any sibling
  occurrence of the same movie/episode.
- Cancel is only available for an occurrence in `pending`, `failed`, or
  `retrying` status — not one whose transfer is actively `downloading` (no
  live-transfer-interrupt mechanism is introduced by this change) and not one
  already `completed`.
- The only way back for a cancelled or retry-exhausted occurrence remains the
  existing full media reset (`POST /movies/:id/reset` /
  `/tvshows/:id/reset`); this change does not introduce a targeted
  "reactivate" action.
- The Downloads tab's details sidepanel gains an "Annuler" action alongside
  the existing Move/Rename actions, shown only when the selected download is
  eligible, and reflects the new terminal status.

## Capabilities

### New Capabilities
- `cancel-download-occurrence`: the manual cancel action — eligibility rules,
  API endpoint, and its effect of permanently excluding one specific
  occurrence from future automatic attempts without touching its siblings.

### Modified Capabilities
- `media-download-scheduling`: candidate matching for tier-1 streams and the
  incomplete-download resume path both exclude an occurrence whose retry
  budget is exhausted or that was manually cancelled, instead of retrying it
  indefinitely.
- `download-url-monitoring`: `retry_count` increments once per external
  download attempt on any path (not only for in-call retryable sub-attempts),
  and reaching `max_retry_attempts` transitions the occurrence to the new
  terminal excluded state.
- `downloads-details-sidepanel`: adds the "Annuler" action and the new
  terminal status to the sidepanel.

## Impact

- `internal/matcher/matcher.go` — candidate queries gain the exclusion.
- `internal/scheduler/build.go` (`mergeIncompleteDownloads`) — resume path
  gains the same exclusion.
- `internal/downloader/downloader.go`, `internal/downloader/state_manager.go`
  — uniform retry accounting, new terminal status transition.
- `internal/models/download.go`, `internal/models/processed_line.go` — new
  terminal status/state values.
- `internal/api/*` — new cancel endpoint.
- `frontend/src/components/DownloadsTab.tsx`, `frontend/src/services/api.ts`,
  `frontend/src/types.ts`, i18n locales — "Annuler" action and status
  display.
