## 1. Fix the flaky test

- [x] 1.1 In `internal/external/tmdb/tmdb_test.go` (`TestRetryAfterHTTPDateFormat`, around line 222), replace `retryAt := time.Now().Add(1 * time.Second).UTC().Format(http.TimeFormat)` with a version that truncates the base time to the second boundary before adding a safety margin (e.g. `time.Now().Truncate(time.Second).Add(2 * time.Second)`), so the HTTP-date's inherent loss of sub-second precision can never push the encoded time below 1 full second in the future. Verify by reading the diff: the base time must be truncated *before* the offset is added, not after.
- [x] 1.2 Update the assertion at line 242-243 (`if elapsed < 1*time.Second`) to check both a lower and upper bound consistent with the new margin (e.g. `elapsed < 1*time.Second || elapsed >= 3*time.Second`), so the test still catches a broken/no-op retry (too fast) or a runaway sleep (too slow), not just the lower bound. Verify by reading the updated assertion covers both directions.
- [x] 1.3 Run `go test -run TestRetryAfterHTTPDateFormat -race -count=20 ./internal/external/tmdb/...` and verify all 20 runs pass with no flakes (previously this test could fail nondeterministically based on sub-second timing).

## 2. Regression check

- [x] 2.1 Run `go test -race -v ./internal/external/tmdb/...` and verify the full package suite passes, including `TestRetryAfterSecondsFormat` (the sibling numeric-seconds test), to confirm no unrelated regression.
- [x] 2.2 Run `go test -race ./...` for the full repo and verify it exits 0, confirming the fix doesn't affect other packages before pushing to the PR branch.
