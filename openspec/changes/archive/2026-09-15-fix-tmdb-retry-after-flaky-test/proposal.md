## Why

`TestRetryAfterHTTPDateFormat` in `internal/external/tmdb/tmdb_test.go` is flaky and blocks CI (observed failing on PR #44, GitHub Actions run 34964410555). The test's mock server computes an HTTP-date `Retry-After` header as `time.Now().Add(1*time.Second)`, but the HTTP-date format (RFC1123, `Mon, 02 Jan 2006 15:04:05 GMT`) has only whole-second precision. Formatting silently drops the sub-second fraction, so the encoded retry time is anywhere from 0ms to ~999ms short of a full second, depending on where in the current second the header is generated. The client then sleeps for that truncated duration, and the test's assertion `elapsed >= 1*time.Second` fails whenever the truncation loses more than the small margin absorbed by request/response overhead — which is common and unrelated to actual CI load.

## What Changes

- Fix `TestRetryAfterHTTPDateFormat` so the encoded `Retry-After` HTTP-date is computed from a time already truncated to the second boundary, then offset by a safety margin, guaranteeing the client's measured wait is always within a known, non-flaky range.
- No production code changes: the TMDB client's existing `Retry-After` parsing and sleep behavior (`internal/external/tmdb/tmdb.go`) is correct as-is; only the test's time construction was flawed.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
(none — this is a test-only fix; no spec-level behavior changes. `skip_specs: true` is set in `.openspec.yaml`.)

## Impact

- **Affected code**: `internal/external/tmdb/tmdb_test.go` only.
- **Affected systems**: CI (`Go CI` workflow, `Test` job) — removes an intermittent red build unrelated to feature PRs.
- **Dependencies**: none.
