## Context

`internal/external/radarr/radarr.go` and `internal/external/sonarr/sonarr.go` (~330 and ~510 lines) each define their own `Client`/`Config` with identical fields (`baseURL`, `apiKey`, `httpClient`, `retryConfig`, `logger`), an identical `New()` bootstrap (default timeout/retry fill-in), an identical `newRequest` (builds the request, sets `X-Api-Key`/`Content-Type`/`Accept` headers, marshals a JSON body), and repeat the same `if resp.StatusCode != want { body, _ := io.ReadAll(...); return fmt.Errorf("unexpected status code %d: %s", ...) }` block at 8 call sites per file, each followed by an identical `json.NewDecoder(resp.Body).Decode(&x)` for GET, or none for PUT. Sonarr additionally has two paginated GETs (`getPagedMovies`-equivalent `getEpisodes`, and radarr's `getPagedMovies`) that decode a `{totalRecords, records}` envelope, which is also duplicated shape (Movie vs Episode).

`internal/api` has one `ErrorResponse{Error, Message}` type (`dto.go:6-9`) and 136 call sites across `handlers.go`, `handlers_frontend.go`, `radarr_sonarr_handlers.go`, `force_download.go`, `resync_path.go` that each hand-write `c.JSON(status, ErrorResponse{Error: "code", Message: "..."})`. Codes in use today are dominated by `database_error` (44), `not_found` (15), `invalid_request` (14), `validation_error` (5), plus a long tail of one-off codes (`fs_error`, `tmdb_error`, `invalid_sort_order`, ...).

See proposal.md for why this is worth doing now and what is explicitly out of scope.

## Goals / Non-Goals

**Goals:**
- One shared implementation for request building, status-code checking, and JSON GET/decode used by both `radarr.go` and `sonarr.go`, including the paginated-envelope variant.
- One shared `respondError` helper in `internal/api`, used everywhere `ErrorResponse` is constructed today.
- Zero observable behavior change: identical error strings, identical status codes, identical JSON response bodies, identical request headers/bodies sent to Radarr/Sonarr.

**Non-Goals:**
- Not introducing `internal/apperrors` into `internal/api`'s error responses (separate decision, out of scope per proposal).
- Not adding a circuit breaker to radarr/sonarr (separate decision).
- Not changing any *currently inconsistent* status code or error code — if two handlers disagree today on what code to use for a similar situation, both keep their existing (different) code after this refactor. Only the plumbing is shared, not the policy.
- Not touching `tmdb.go` (it already has its own cache/rate-limiter/circuit-breaker shape and isn't part of the duplication being addressed here).

## Decisions

**1. New package `internal/external/httpclient` shared by radarr and sonarr (not sonarr importing radarr, or vice versa).**
Neither client should depend on the other's package — they're peers. A small new leaf package avoids a false hierarchy and is trivially unit-testable in isolation (`httptest.Server`, no Radarr/Sonarr-specific types).

Shape:
```go
type Client struct {
    BaseURL string
    APIKey  string
    HTTP    *http.Client
    Retry   retry.Config
    Logger  *logger.Logger
}

func New(baseURL, apiKey string, timeout time.Duration, retryCfg retry.Config, log *logger.Logger) *Client
func (c *Client) NewRequest(ctx context.Context, method, endpoint string, body any) (*http.Request, error)
func CheckStatus(resp *http.Response, want ...int) error
func Get[T any](ctx context.Context, c *Client, endpoint string) (T, error)
func GetPage[T any](ctx context.Context, c *Client, endpoint string) (records []T, total int, err error)
func Put(ctx context.Context, c *Client, endpoint string, body any) error
```
`radarr.Client` and `sonarr.Client` keep their existing exported `Config`/`New()` signatures (external contract unchanged) but hold a `*httpclient.Client` internally instead of the four duplicated fields, and their private `getMovies`/`getSeries`/etc. become one-line calls into `httpclient.Get[Movie]`/`httpclient.GetPage[Episode]`. `retry.Do(...)` and `apperrors.ExternalServiceError(...)` wrapping stay exactly where they are today, at the public-method level in `radarr.go`/`sonarr.go` — `httpclient` does not know about retry or `apperrors`, keeping it a dumb, reusable leaf.

*Alternative considered*: put the shared helper directly in one of the two existing packages and have the other import it (e.g. sonarr imports radarr's helper). Rejected — creates an arbitrary dependency between two peer integrations and makes it look like one is "primary."

*Alternative considered*: a shared base struct both `Client` types embed by value/pointer instead of a field. Rejected in favor of a plain field (`http *httpclient.Client`) — embedding would promote `httpclient.Client`'s methods onto `radarr.Client`'s public API surface, which is not desired (`radarr.Client` should only expose `GetMissingMovies`, `UpdateMovie`, etc., not `NewRequest`/`Get`).

**2. Use Go generics (module is on Go 1.25) for `Get[T]`/`GetPage[T]` rather than `interface{}` + manual type assertion, or per-type wrapper functions.**
Keeps each call site (`httpclient.Get[Movie](ctx, c.http, endpoint)`) as a single readable line and preserves compile-time type safety, matching how the rest of a Go-1.25 codebase should use generics for this exact "decode a JSON GET response into T" shape.

**3. `respondError(c *gin.Context, status int, code, message string)` in `internal/api`, plain wrapper only — no new error taxonomy.**
```go
func respondError(c *gin.Context, status int, code, message string) {
    c.JSON(status, ErrorResponse{Error: code, Message: message})
}
```
Every existing `c.JSON(status, ErrorResponse{Error: "...", Message: "..."})` call site is replaced with `respondError(c, status, "...", "...")` — same arguments, same output, mechanical find-and-replace per file followed by a build+test check. No convenience wrappers for specific codes (e.g. a `respondDBError` shortcut for the 44 `database_error` sites) are added in this change, to keep the diff mechanical and low-risk; that consolidation is a natural, separate follow-up once this helper exists everywhere.

*Alternative considered*: also centralize the *choice* of code/status per situation (e.g. one helper per error category). Rejected for this change — that changes behavior in edge cases (different handlers may map a similar failure to different codes today) which is explicitly out of scope per the proposal; this change only extracts the response-writing plumbing.

## Risks / Trade-offs

- **[Risk]** A behavior-preserving refactor across ~150 call sites and 2 full external-client files is easy to get subtly wrong (e.g. reordering an `Error`/`Message` argument, or picking the wrong `want` status codes for `CheckStatus`). → **Mitigation**: migrate one file at a time (radarr, then sonarr, then each `internal/api` handler file), run `go test ./...` after each file, and diff-review that error strings/status codes are byte-identical to before.
- **[Risk]** `httpclient.Get[T]`/`GetPage[T]` must reproduce today's exact error message text (`"unexpected status code %d: %s"`, `"failed to decode response: %w"`, `"failed to create request: %w"`, `"failed to marshal request body: %w"`) since any test or caller pattern-matching on error strings would otherwise break. → **Mitigation**: keep the exact same format strings in the shared helper; add a table-driven test in `httpclient` asserting the exact error text for each failure branch.
- **[Trade-off]** `radarr.Client`/`sonarr.Client` now hold a `*httpclient.Client` field instead of four inline fields — a small indirection for readers unfamiliar with the new package, offset by removing ~90 lines of duplicated plumbing per file.

## Migration Plan

1. Add `internal/external/httpclient` with `Client`, `NewRequest`, `CheckStatus`, `Get[T]`, `GetPage[T]`, `Put`, plus unit tests (using `httptest.Server`) covering success, non-2xx status, and decode-failure paths for each helper.
2. Refactor `radarr.go` to use it; run `go test ./internal/external/radarr/...`.
3. Refactor `sonarr.go` to use it; run `go test ./internal/external/sonarr/...`.
4. Add `respondError` to `internal/api` (e.g. in `middleware.go` or a new small `errors.go` in that package); run the full `internal/api` test suite once to confirm the helper itself compiles and is wired correctly (no call sites migrated yet).
5. Migrate call sites file by file (`handlers.go` → `handlers_frontend.go` → `radarr_sonarr_handlers.go` → `force_download.go` → `resync_path.go`), running that file's test suite after each.
6. Run the full test suite (`go test ./...`) once all files are migrated.

Rollback is a plain `git revert` of the relevant commit(s) — no data migration, no config change, no deployed-contract change is involved.

## Open Questions

None — the scope, shape, and file-by-file migration order above are settled; nothing here needs to be revisited during `tasks`/implementation.
