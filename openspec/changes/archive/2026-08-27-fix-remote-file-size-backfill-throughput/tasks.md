## 1. Configuration

- [x] 1.1 Add `Concurrency int` (`mapstructure:"concurrency"`) and `RetryCooldownHours int` (`mapstructure:"retry_cooldown_hours"`) fields to `RemoteFileSizeConfig` in `internal/config/config.go`, and verify the struct compiles with `go build ./...`
- [x] 1.2 Set Viper defaults: `remote_file_size.concurrency` = 10, `remote_file_size.retry_cooldown_hours` = 168, and raise `remote_file_size.per_run_cap` default from 200 to 2000 (near the existing defaults block in `internal/config/config.go`), and verify with a config test that loads defaults with no overrides and asserts these three values
- [x] 1.3 Document the two new keys (and the raised `per_run_cap` default) in `config.yml.example`, and verify by diffing against the current file to confirm only the `remote_file_size` block changed

## 2. Eligibility query and schema

- [x] 2.1 Add a GORM index tag for a composite index `idx_processed_lines_remote_file_size_probe` on `(content_type, remote_file_size, remote_file_size_checked_at)` to `ProcessedLine` in `internal/models/processed_line.go`, and verify the index is created by running auto-migration against a test database and inspecting `information_schema.indexes` (or `\d processed_lines` in psql)
- [x] 2.2 Update the eligibility query in `BackfillRemoteFileSize` (`internal/processor/remote_file_size.go`) to select rows where `content_type IN (...) AND line_url IS NOT NULL AND (remote_file_size_checked_at IS NULL OR (remote_file_size IS NULL AND remote_file_size_checked_at < now() - cooldown))`, threading the cooldown duration in as a parameter, and verify with a unit test asserting a row with `remote_file_size_checked_at` older than the cooldown and `remote_file_size` NULL is selected
- [x] 2.3 Verify a row with `remote_file_size` already set is never selected regardless of how old `remote_file_size_checked_at` is, and a row that failed recently (within the cooldown window) is not selected — add/extend unit tests for both cases

## 3. Concurrent probing

- [x] 3.1 Replace the sequential `for i := range lines` loop in `BackfillRemoteFileSize` with a bounded worker pool sized by the new `concurrency` parameter (semaphore channel or equivalent), keeping each worker's probe-then-persist logic (including the `remote_file_size_checked_at` update on both success and failure) intact per line, and verify `go build ./...` succeeds
- [x] 3.2 Guard `RemoteFileSizeBackfillStats` (`Checked`/`Found`/`Errors`) and the per-line DB update against concurrent access (mutex or atomic counters, since multiple goroutines write results), and verify with `go test -race ./internal/processor/...`
- [x] 3.3 Add a unit test asserting no more than `concurrency` probes are in flight simultaneously (e.g., using a test HTTP server that tracks concurrent in-flight requests via a counter/semaphore) and that a single line's probe failure does not abort or block sibling in-flight probes

## 4. Update existing tests for new behavior

- [x] 4.1 Update `TestBackfillRemoteFileSize_BothProbesFail` (and any other test asserting the old "checked once, never retried" behavior) in `internal/processor/remote_file_size_test.go` to reflect that a failed row becomes eligible again only after the cooldown elapses, not immediately and not never
- [x] 4.2 Run the full existing suite `go test ./internal/processor/...` and confirm all tests (including `TestBackfillRemoteFileSize_HeadSuccess`, `TestBackfillRemoteFileSize_HeadFailureFallsBackToRangeGet`, `TestBackfillRemoteFileSize_ChannelsExcluded`, `TestBackfillRemoteFileSize_PerRunCapRespected`) pass unmodified in behavior other than the deliberately updated retry test

## 5. Verification against the reported symptom

- [x] 5.1 Run `stalkeer process` locally (or in the dev environment) against a playlist with a large eligible backlog and verify via logs/DB query that more than 200 lines are probed in a single run, with concurrent HTTP activity observable (e.g., via request timing overlap in logs)
- [x] 5.2 Manually verify in the M3U playlist side panel that a previously "Unavailable" item shows a resolved size after its line is probed successfully, confirming the fix is visible end-to-end and not just at the DB layer
