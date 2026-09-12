## Context

See proposal.md - Why/What Changes for motivation. Relevant current-state facts that shape the approach below:

- `internal/external/httpclient` (added by `extract-shared-http-helpers`) is the single execution chokepoint for every Radarr and Sonarr call: `radarr.go`/`sonarr.go` each call `retry.Do(ctx, c.http.Retry, func() error { return c.getXxx(...) }, apperrors.IsRetryable)`, and every `getXxx` delegates to `httpclient.Get[T]` / `GetPage[T]` / `Put`, which all end at `c.HTTP.Do(req)`.
- `apperrors.IsRetryable(err)` only returns `true` for a `*apperrors.AppError` carrying one of a fixed set of codes. Errors from `httpclient` (network failures from `c.HTTP.Do`, non-2xx responses from `CheckStatus`) are plain errors, never wrapped as `AppError` - so `retry.Do`'s classifier rejects them on the first attempt and the configured retry policy never actually triggers today.
- `radarr.New`/`sonarr.New` are called fresh at 14 call sites: 4 files under `internal/api` (one new: `system_status.go`) construct a client per incoming HTTP request, and `cmd/download.go` constructs one client per `download` run, shared across its worker pool. Any breaker created *inside* `New()` would reset every request on the API-server call sites, making it a no-op there.
- `Client.SystemStatus` (used only by the system-status reachability checks) deliberately bypasses `httpclient.Get`/`GetPage`/`Put` and does its own single unretried request, building a `*StatusError` (exposing `StatusCode() int`) so `internal/api/system_status.go`'s `classifyReachabilityError` can distinguish a 401 from a plain connectivity failure.
- `tmdb.go`'s existing breaker is a useful precedent for defaults (`MaxFailures: 5`, `Timeout: 60s`) and for the `Server`-owns-the-long-lived-client pattern (`Server.tmdbClient`, built once in `NewServer()`).

## Goals / Non-Goals

**Goals:**
- One shared circuit breaker per service (Radarr, Sonarr), reused across every call for the life of the process (API server) or the life of a run (`download` command).
- Retry actually retries transient failures (timeout, connection failure, HTTP 429, 5xx) before they count against the breaker; non-transient failures (4xx other than 429) are neither retried nor counted against the breaker.
- The system status endpoint's Radarr/Sonarr checks reflect an open breaker instantly, without an outbound call.
- No duplication: the breaker and the error classification live once, in the shared `httpclient` package, not copy-pasted into `radarr.go` and `sonarr.go`.

**Non-Goals:**
- Exposing breaker state as a Prometheus metric or a new dedicated API field - the existing status-endpoint reason string is enough for this change; metrics can be a follow-up.
- Making breaker thresholds user-configurable - reuses fixed defaults, same as the existing TMDB/M3U breakers.
- Changing TMDB's or the M3U downloader's existing breaker wiring.
- Persisting breaker state across process restarts - in-memory only, exactly like the existing usages.

## Decisions

### D1: The breaker executes inside `httpclient.Get`/`GetPage`/`Put`, not in `radarr.go`/`sonarr.go`
`httpclient.Client` gets a `Breaker *circuitbreaker.CircuitBreaker` field (nil-safe: `Get`/`GetPage`/`Put` call `c.HTTP.Do` directly when it's nil). `Get`, `GetPage`, and `Put` wrap their `c.HTTP.Do(req)` call in `c.Breaker.Execute(...)` when non-nil.
**Alternative considered**: wrap each of the ~10 `retry.Do(...)` call sites in `radarr.go`/`sonarr.go` individually. Rejected - this is exactly the duplication `extract-shared-http-helpers` just removed; one chokepoint means one place to get right and one place to test.

### D2: Composition order is retry (outer) wrapping the breaker (inner) - a consequence of D1, not a separate choice
`retry.Do` stays where it is today, in `radarr.go`/`sonarr.go`, calling into `httpclient.Get`/`GetPage`/`Put` where the breaker now lives. So each retry attempt passes through the breaker individually - the same composition TMDB already uses. This only becomes meaningful once D4 (below) makes retry actually retry; until then the distinction was moot because retry never ran more than once.
**Alternative considered**: breaker wrapping the whole retry loop (the M3U downloader's style). Rejected - it would require moving the breaker back up into `radarr.go`/`sonarr.go` around each `retry.Do` call, reintroducing the per-call-site duplication D1 avoids. Per-attempt accounting is also the more useful signal here: repeated attempts against the same failing instance are exactly what should count toward opening the circuit.

### D3: Breaker ownership and lifetime
A new `Breaker *circuitbreaker.CircuitBreaker` field on `radarr.Config`/`sonarr.Config`, threaded into `httpclient.New(...)`. Each service gets exactly one breaker instance, built once and passed to every `radarr.New`/`sonarr.New` call:
- `api.Server` gains `radarrBreaker`/`sonarrBreaker *circuitbreaker.CircuitBreaker` fields, built once in `NewServer()` (same pattern as `tmdbClient`), and passed as `Config.Breaker` at all 14 API-handler call sites.
- `cmd/download.go` builds one breaker per service once per run, before constructing `radarrFullClient`/`sonarrFullClient`, shared by the worker pool exactly as the client itself already is.
**Alternative considered**: a package-level singleton inside `radarr`/`sonarr` (e.g. `var breaker = circuitbreaker.New(...)`). Rejected - harder to isolate in tests (shared mutable state across parallel test runs), and inconsistent with the existing `Server`-owns-its-long-lived-clients pattern.

### D4: Error classification - one predicate shared by retry and the breaker
`httpclient.CheckStatus` and the `c.HTTP.Do` error path classify failures into `apperrors.AppError` with a retryable code: network/timeout errors -> `CodeServiceTimeout`; connection failures -> `CodeServiceUnavailable`; HTTP 429 -> `CodeRateLimited`; HTTP 5xx -> `CodeServiceUnavailable`. Other 4xx responses are left unclassified (not retryable, matching today's `IsRetryable` default of `false` for anything that isn't one of those codes).

The breaker's `IsSuccessful` config reuses the same signal instead of the default `err == nil`: an error counts as a breaker failure when `apperrors.IsRetryable(err)` is true, **or** when it exposes `StatusCode() int` (the `StatusError` shape used by `SystemStatus`, see D5) with a 429/5xx code. This keeps "is this the service's fault" as a single policy: a bad API key (401) or a bad request (400/404) fails the caller immediately and is surfaced with its real reason, but does **not** open the circuit - retrying or fail-fasting doesn't help a bad credential, and letting it trip the breaker would mask "invalid credentials" behind a generic "circuit open" for every subsequent call.
**Alternative considered**: classify in `radarr.go`/`sonarr.go` per client. Rejected for the same duplication reason as D1.

### D5: `SystemStatus` is routed through the breaker without going through `Get`/`GetPage`/`Put`
`SystemStatus`'s body is wrapped in `c.http.Breaker.Execute(...)` (single call, no retry - unchanged from today's "fail fast" intent), but keeps building its own `*StatusError` rather than switching to `CheckStatus`, preserving the existing 401-vs-other-failure distinction `classifyReachabilityError` relies on. When the breaker is already open, `Execute` returns `circuitbreaker.ErrOpenState` **without invoking the call at all** (see `beforeRequest` in `internal/circuitbreaker`) - so "no live call while open" falls out of the breaker's existing behavior for free, no separate pre-check needed.
`system_status.go`'s `classifyReachabilityError` gains one more `errors.As` branch, checked first: `circuitbreaker.ErrOpenState` (and `ErrTooManyRequests`, the half-open equivalent) maps to a new `reasonCircuitOpen` reason string.

### D6: Breaker defaults match the existing TMDB/M3U breakers
`MaxFailures: 5`, `Timeout: 60s`, `MaxHalfOpenRequests: 1` - no new tuning surface for this change (see Non-Goals).

## Risks / Trade-offs

- **[Risk]** A shared breaker means a burst of unrelated failures (e.g. a brief network blip hitting several concurrent handler requests at once) could open the circuit and fail-fast legitimate subsequent requests for up to the timeout window, where today each request independently retried/timed out on its own. -> **Mitigation**: this is the intended trade-off of a circuit breaker (bounded fail-fast vs. unbounded independent waits) and matches the behavior already accepted for TMDB and the M3U downloader; the 60s open window is short relative to the handler timeouts it replaces.
- **[Risk]** If the breaker's `IsSuccessful` predicate were left at its default (`err == nil`), repeated invalid-credentials (401) responses would open the circuit and the status endpoint would report "circuit open" instead of "invalid credentials" for every subsequent check, hiding the actionable reason. -> **Mitigation**: D4 makes 401/400/404 count as "successful" for breaker accounting (they're not the service's fault), so they never trip the circuit; the real reason always reaches the caller.
- **[Risk]** Sharing one breaker per service across every handler means a single misbehaving endpoint's failures can fail-fast an unrelated endpoint's calls to the same service. -> **Mitigation**: this is the explicit goal (protect Radarr/Sonarr as a whole from being hammered while unhealthy) rather than a side effect; per-endpoint isolation was never a requirement here.

## Migration Plan

Purely additive and in-process (no persisted state, no config surface, no API shape change). Ship behind normal code review/tests; rollback is a plain revert. No phased rollout needed - the breaker starts closed on every process start, so behavior is unchanged until real consecutive failures occur.

## Open Questions

- Should `cmd/download.go`'s breaker use a higher `MaxFailures` than the API server's, since a single `download` run can make far more Radarr/Sonarr calls in quick succession via its worker pool than interactive API traffic typically does? Can be tuned during implementation/testing without changing the spec, the approach, or the task breakdown.
