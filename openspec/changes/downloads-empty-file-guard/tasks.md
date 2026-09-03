## 1. Config

- [ ] 1.1 Add `MinFileSizeMB int64` (`mapstructure:"min_file_size_mb"`) to `DownloadsConfig` in `internal/config/config.go`, defaulting to `1`, mirroring `M3UDownloadConfig.MaxFileSizeMB`'s pattern. Verify with a config test asserting the default applies when unset and an explicit value overrides it.
- [ ] 1.2 Document the new setting in `config.yaml` (and Helm chart values, if download settings are templated there per `configuration-management`) with its default. Verify by grepping the config file(s) for the new key.

## 2. Transfer-Level Guard

- [ ] 2.1 Add a sentinel error (e.g. `ErrFileTooSmall`) in `internal/downloader` carrying the observed byte count, and return it from `downloadFileWithResume` (`internal/downloader/downloader.go`) when a completed, error-free transfer's `bytesRead` is below the configured minimum. Verify with a unit test simulating a small HTTP response body and asserting the sentinel error is returned instead of a nil error.
- [ ] 2.2 Wire the minimum size threshold into `Downloader` (read from `DownloadsConfig.MinFileSizeMB` at construction in `New()`, converted to bytes) so `downloadFileWithResume` has it available without a new parameter threaded through every call. Verify by asserting a `Downloader` built with a given config value rejects a transfer under that size and accepts one at/above it.
- [ ] 2.3 Update `isRetryableError`/`apperrors.IsRetryable` (whichever gates `retry.Do` in `Download()`) to treat `ErrFileTooSmall` as retryable, sharing the existing `retryConfig.MaxAttempts` budget. Verify with a test asserting `retry.Do` attempts again after this sentinel and stops once `MaxAttempts` is exhausted.

## 3. Failure Path

- [ ] 3.1 Confirm (or adjust) `Download()`'s existing failure branch produces an `ErrorMessage` that clearly states the file was empty/undersized and includes the last attempt's byte count when the final error is `ErrFileTooSmall`, reusing the existing `stateManager.UpdateState(ctx, downloadInfoID, models.DownloadStatusFailed, &errMsg)` call. Verify with a test asserting the persisted `error_message` mentions the undersized condition and the byte count.
- [ ] 3.2 Verify a download that never exceeds the minimum size across all retry attempts ends in `DownloadStatusFailed`, never `DownloadStatusCompleted`, and that no file is moved to its final destination path. Cover with an integration-style test in `internal/downloader/downloader_test.go` using a test HTTP server returning an undersized body on every attempt.
- [ ] 3.3 Verify a download that returns an undersized body on its first attempt(s) but a valid-size body on a later attempt within the retry budget completes normally (`DownloadStatusCompleted`, correct `file_size`). Cover with a test HTTP server returning a small body once, then a full body.

## 4. Shared Path Coverage

- [ ] 4.1 Verify both `force_download.go`'s call to `Downloader.Download` and `resume_helper.go`'s batch path (`parallel.go` -> `Downloader.Download`) exhibit the same undersized-then-fail behavior without any changes to those files, by exercising the existing `downloader` package tests rather than adding call-site-specific tests (per the design's single-choke-point decision).

## 5. Verification

- [ ] 5.1 Run `go test ./internal/downloader/... ./internal/config/...` and confirm all tests pass, including the new undersized-transfer cases.
- [ ] 5.2 Run `openspec validate downloads-empty-file-guard --strict` and resolve any reported issues.
