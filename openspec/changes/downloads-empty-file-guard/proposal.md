## Why

A forced download can complete an HTTP transfer with a `200 OK` response and an empty (or near-empty) body — typical of a dead/expired IPTV stream link — without that ever surfacing as an error. `Downloader.Download()` (`internal/downloader/downloader.go`) never checks how many bytes were actually written before finalizing state, so the item is marked `completed` with `file_size: 0`, and the UI reports success for a file that doesn't exist.

## What Changes

- Add a minimum downloaded-size guard to the shared HTTP transfer step (`downloadFileWithResume` in `internal/downloader/downloader.go`), used by both the force-download and resume-incomplete-downloads flows through the same `Downloader.Download()` entry point.
- A transfer that finishes with fewer bytes written than the configured minimum is treated as a retryable failure, sharing the existing retry budget already used for network errors (currently 3 attempts total, unified rather than adding a separate counter).
- Every retry restarts the transfer from byte 0 (no resume-from-partial for an undersized attempt).
- If every attempt still ends under the minimum, the download SHALL be marked `failed` (not `completed`) with an explicit error message stating the file was empty/undersized and how many bytes were written, so it reads clearly wherever `DownloadInfo.ErrorMessage` is surfaced (sidepanel, logs).
- Add a configurable minimum file size setting (default 1 MB), following the existing `MaxFileSizeMB`-style pattern already used by `m3udownloader`'s config.

## Capabilities

### New Capabilities
- `download-transfer-integrity`: minimum-size validation on the shared media download transfer, with retry-then-fail semantics and an explicit failure reason, so a download is never marked `completed` with an unusably small file.

### Modified Capabilities
(none — `DownloadInfo.status`/`error_message` and the retry mechanism already exist; this change adds a new failure cause within their existing contract, not a new API shape or behavior visible to `api-downloads-monitoring`.)

## Impact

- Backend: `internal/downloader/downloader.go` (`downloadFileWithResume`, `isRetryableError`, `Download`'s failure/completion handling), `internal/config` (new minimum-size setting), `internal/models/download.go` if a new failure detail is needed beyond the existing `ErrorMessage` string.
- Both callers of `Downloader.Download()` inherit the fix without their own changes: `internal/api/force_download.go` and `internal/downloader/resume_helper.go` (via `internal/downloader/parallel.go`).
- Config surface: `config.yaml` / Helm chart values gain a new field mirroring `m3udownloader`'s `MaxFileSizeMB`.
