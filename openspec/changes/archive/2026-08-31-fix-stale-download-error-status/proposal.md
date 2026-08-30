## Why

A download that failed with a transient error (e.g. HTTP 429) and later succeeds on a retry keeps showing the old error message on the Downloads tab even though its status is `completed`. This happens because `DownloadInfo` records are reused across retry attempts (keyed by `ProcessedLine`), but `error_message` is only ever written on failure and is never cleared when a later attempt succeeds. The Downloads tab renders the error banner whenever `error_message` is non-empty, regardless of the current status, so users see a "complete" card that still looks broken and can't tell it actually finished.

## What Changes

- Clear `DownloadInfo.error_message` (and reset `retry_count`-derived staleness) when a download starts a fresh attempt (`status` transitions to `downloading`), so a record carried over from a prior failed attempt no longer reports an error once it starts progressing again.
- Update the Downloads tab to only render the error banner when the download's current `status` is `failed`, instead of rendering it whenever `error_message` is present — this matches the already-documented intent in `downloads-display-ui` ("GIVEN a failed download... show error message") which the current implementation does not actually enforce.
- No API contract or schema changes: `error_message` remains a nullable string field: the fix changes *when* it is written/cleared and *when* it is displayed, not its shape.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `download-url-monitoring`: add a requirement that `error_message` is cleared when a `DownloadInfo` record transitions to `downloading` for a new attempt, so it never outlives the failure it describes.
- `downloads-display-ui`: tighten the existing failed-download requirement so the error banner is gated on `status === 'failed'` rather than on the mere presence of `error_message`.

## Impact

- `internal/downloader/state_manager.go` (`UpdateState`): clear `error_message` on the `downloading` transition.
- `internal/downloader/downloader.go`: no contract change, but the fix relies on `Download()`'s existing `downloading` transition at the start of each invocation.
- `frontend/src/components/DownloadsTab.tsx`: gate the error banner render on `item.status === 'failed'`.
- No database migration needed (existing nullable `error_message` column); no API response shape change.
