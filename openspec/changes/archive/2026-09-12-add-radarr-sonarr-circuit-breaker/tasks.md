## 1. Shared httpclient: breaker + error classification

- [x] 1.1 Add a `Breaker *circuitbreaker.CircuitBreaker` field to `httpclient.Client` and a parameter to `httpclient.New(...)`; verify `go build ./internal/external/httpclient/...` succeeds and existing `httpclient_test.go` cases still pass with a nil breaker.
- [x] 1.2 Wrap the `c.HTTP.Do(req)` call in `Get[T]`, `GetPage[T]`, and `Put` in `c.Breaker.Execute(...)` when `Breaker` is non-nil (call `c.HTTP.Do` directly when nil); add a unit test in `httpclient_test.go` that a breaker opened via repeated failures makes a subsequent `Get`/`GetPage`/`Put` call return the breaker's open-state error without hitting the test server.
- [x] 1.3 Classify `httpclient` errors into `apperrors.AppError` with a retryable code: network/timeout errors from `c.HTTP.Do` -> `CodeServiceTimeout`/`CodeServiceUnavailable` as appropriate, and non-2xx responses in `CheckStatus` -> `CodeRateLimited` for 429, `CodeServiceUnavailable` for 5xx, left unclassified for other 4xx; verify with unit tests in `httpclient_test.go` asserting `apperrors.IsRetryable(err)` is true for timeout/connection/429/5xx cases and false for other 4xx.
- [x] 1.4 Add an `IsSuccessful` helper (used when constructing each service's breaker in task 3.1) that returns true for `err == nil`, for any error where `apperrors.IsRetryable(err)` is false, and treats an error exposing `StatusCode() int` as a failure only for 429/5xx codes; cover with a table-driven unit test including a `*radarr.StatusError`-shaped 401 (must count as success/not-open-circuit) and a 503 (must count as failure).

## 2. Radarr and Sonarr clients

- [x] 2.1 Add `Breaker *circuitbreaker.CircuitBreaker` to `radarr.Config` and `sonarr.Config`, threaded into `httpclient.New(...)` in each package's `New(cfg Config)`; verify `go build ./...` succeeds.
- [x] 2.2 Route `radarr.Client.SystemStatus` and `sonarr.Client.SystemStatus` through `c.http.Breaker.Execute(...)` (single call, no retry, unchanged otherwise); verify existing `SystemStatus` tests in `radarr_test.go`/`sonarr_test.go` still pass, and add a case where a pre-opened breaker makes `SystemStatus` return `circuitbreaker.ErrOpenState` without an HTTP call reaching the test server.

## 3. Breaker ownership and wiring into call sites

- [x] 3.1 Add `radarrBreaker`/`sonarrBreaker *circuitbreaker.CircuitBreaker` fields to `api.Server`, constructed once in `NewServer()` using the same defaults as the existing TMDB breaker (`MaxFailures: 5`, `Timeout: 60s`, `MaxHalfOpenRequests: 1`) and the `IsSuccessful` helper from 1.4; verify `go build ./internal/api/...` succeeds.
- [x] 3.2 Pass `Breaker: s.radarrBreaker` / `Breaker: s.sonarrBreaker` at every `radarr.New`/`sonarr.New` call site in `internal/api/radarr_sonarr_handlers.go`, `internal/api/force_download.go`, `internal/api/resync_path.go`, and `internal/api/system_status.go`; verify with `grep -rn "radarr.New(\|sonarr.New(" internal/api` that every call site now includes `Breaker:`.
- [x] 3.3 Build one breaker per service in `cmd/download.go` before constructing `radarrFullClient`/`sonarrFullClient`, sharing it across the run's worker pool exactly as the client itself is already shared; verify `go build ./cmd/...` succeeds and `stalkeer download --dry-run` still runs cleanly against a test/staging Radarr+Sonarr.

## 4. System status short-circuit

- [x] 4.1 Add a `reasonCircuitOpen` reason constant and a branch in `classifyReachabilityError` (checked before the timeout/network branches) that maps `circuitbreaker.ErrOpenState` and `circuitbreaker.ErrTooManyRequests` to it; verify with a unit test in `system_status_test.go` asserting the reason for each error.
- [x] 4.2 Add an integration-style test in `system_status_test.go` that pre-opens a Radarr (and Sonarr) breaker and asserts the status endpoint reports KO with the circuit-open reason without making an HTTP call to the (failing-if-called) test server, and without waiting out `systemStatusCheckTimeout`.

## 5. Cross-cutting verification

- [x] 5.1 Run the full test suite (`go test ./...`) and confirm all existing Radarr/Sonarr/httpclient/system-status tests pass unchanged alongside the new ones.
- [x] 5.2 Manually or via a scripted check, confirm one service's forced failures do not affect the other's breaker state (independent Radarr/Sonarr circuits), matching the `radarr-sonarr-resilience` spec's independence requirement.
- [x] 5.3 Run `openspec validate add-radarr-sonarr-circuit-breaker --strict` and fix any reported issues.
