## Context

See `proposal.md` - Why. Relevant current state in `internal/downloader/downloader.go`:

- `Download()` runs the transfer inside `retry.Do(ctx, retryConfig, func() error { ... d.downloadFile(...) ... }, apperrors.IsRetryable)`. `retryConfig.MaxAttempts` defaults to 3 (`New()`, when `retryAttempts == 0`) and is shared by every call site — `force_download.go` and `resume_helper.go`/`parallel.go` both construct their `Downloader` through the same `New()`.
- `downloadFile` always calls `downloadFileWithResume(ctx, url, destPath, 0, onProgress)` — i.e. every attempt inside the `retry.Do` loop starts at byte 0. The `startByte > 0` branch of `downloadFileWithResume` is only reached via its own internal recursive fallback (line ~366, when a `Range` request is rejected mid-attempt), not from `Download()`'s retry loop. So "every retry restarts from scratch" is already the existing behavior for the case this change adds — no new plumbing needed for that part.
- `apperrors.IsRetryable(err)` is the sole gate deciding whether `retry.Do` tries again; it currently only recognizes transport/HTTP-level errors (see `internal/apperrors`).
- On success, `downloadFileWithResume` returns `&DownloadResult{FileSize: totalBytes, BytesRead: totalBytes}` with `err == nil` unconditionally — there is no path today where a technically-successful HTTP transfer produces a non-nil error, regardless of `totalBytes`.
- `Download()`'s failure branch (transfer path) already does the right thing once it receives an error: `stateManager.UpdateState(ctx, downloadInfoID, models.DownloadStatusFailed, &errMsg)` plus `updateProcessedLineState(..., models.StateFailed)`. Reusing this existing branch is enough to get failure handling right, once the transfer step actually returns an error for the undersized case.
- Existing precedent for a configurable size threshold: `internal/config.M3UDownloadConfig.MaxFileSizeMB` (`mapstructure:"max_file_size_mb"`), read by `m3udownloader.Downloader`. `internal/config.DownloadsConfig` (`mapstructure:"downloads"`) is the equivalent struct for the media downloader and does not currently have a size field.

## Goals / Non-Goals

**Goals:**
- Make an undersized transfer (below a configurable minimum) return a distinct, retryable error from the transfer step, so it flows through the retry loop and failure path that already exist for network errors — no new retry loop, no new failure-handling branch.
- Give the eventual `failed` state a message that clearly names this cause (vs. a generic network error) and includes the last attempt's byte count.
- Cover both callers (`force_download.go`, `resume_helper.go`) through the single shared `Download()`/`downloadFileWithResume` code path, per the proposal's "un fix central" scoping.

**Non-Goals:**
- No change to `retryConfig.MaxAttempts` itself, its backoff, or per-call-site overrides — the undersized case is folded into the existing budget as-is, not given its own counter.
- No change to the `moveFile` cross-filesystem fallback's existing size-match check, or to any resume-from-partial-bytes logic used elsewhere (`ResumeSupport`/`downloadFileWithResume`'s internal `startByte > 0` recursion) — those are unrelated to a same-`Download()`-call retry restarting from scratch.
- No new DB column: the existing `DownloadInfo.ErrorMessage *string` already carries free-form failure text; the undersized case just writes a specific message into it.

## Decisions

**Detect the undersized condition in `downloadFileWithResume`, right after `io.Copy` returns, comparing `bytesRead` (for a fresh, non-resumed attempt) against the configured minimum.**
This is the one place that already knows the exact byte count for the attempt and already returns `(*DownloadResult, string, error)`. Returning a new sentinel error here — e.g. `ErrFileTooSmall`, wrapping the byte count — means `Download()`'s existing retry/failure plumbing needs no new branching to discover the condition; it just needs to route the existing error through unchanged.
Alternative considered: check `result.FileSize` back in `Download()` after `retry.Do` returns successfully — rejected because `retry.Do` would then have no way to retry the case at all (success already ended the loop with `err == nil`); the check has to happen before the loop decides "success".

**Extend `apperrors.IsRetryable` (or the `isRetryable` callback passed to `retry.Do`) to treat the new sentinel as retryable, sharing `retryConfig.MaxAttempts`.**
Matches the decision to homogenize with the existing network-retry count (3) rather than add a second, independent retry allowance — one budget, one place that decides "give up".

**On final failure, build the error message at the point `Download()` already formats `errMsg := err.Error()` for the `DownloadStatusFailed` update.**
The sentinel error's `Error()` string carries "empty or undersized file" plus the last byte count, so it naturally becomes the `ErrorMessage` persisted by the existing `stateManager.UpdateState(..., DownloadStatusFailed, &errMsg)` call — no new update call, no new column.

**New config field on `DownloadsConfig` (`internal/config`), e.g. `MinFileSizeMB int64` with `mapstructure:"min_file_size_mb"`, default `1`.**
Mirrors `M3UDownloadConfig.MaxFileSizeMB`'s existing shape and mapstructure convention rather than introducing a differently-named or differently-typed setting; keeps the two downloaders' config surfaces consistent for anyone reading `config.yaml`. Read once when constructing `Downloader` (`New()`), same lifecycle as `retryConfig.MaxAttempts`.
Alternative considered: hardcode 1 MB as an unexported constant — rejected per explicit product decision to keep this adjustable, consistent with how `m3udownloader` already exposes its own size guard as config rather than a constant.

**Do not special-case the resume path's own `startByte > 0` recursive branch.**
Within a single `Download()` call, every `retry.Do` iteration re-enters at `startByte = 0` already (see Context), so "retry repart de zéro" falls out of existing behavior once the undersized case is wired into the same retryable-error path — the internal resume recursion (used only for a rejected mid-attempt `Range` request) is a different mechanism and stays untouched.

## Risks / Trade-offs

- [A legitimately tiny real media file (e.g. a very short clip) below 1 MB would now be misclassified as "undersized" and fail after 3 attempts] -> Mitigation: the threshold is configurable per the design above; 1 MB is the requested default and is far below any real movie/episode encode, so this is an accepted trade-off for the catalog this system targets.
- [Folding the undersized case into the shared retry budget means a download that already retried once for a network blip has fewer remaining attempts left to recover from a dead link, and vice versa] -> Mitigation: explicit product decision ("changer pour 3 retry, histoire d'homogénéiser") to keep one unified, simpler budget rather than reason about two independent counters.
- [`IsRetryable`/`isRetryableError` is a shared classification point; adding a new case there means every future error type funnelled through the same path must remember it now covers "successful transfer, bad content" in addition to "transport failure"] -> Mitigation: keep the new sentinel and its check colocated and clearly named (`ErrFileTooSmall` or equivalent) so the distinction stays legible in code, not just in behavior.
