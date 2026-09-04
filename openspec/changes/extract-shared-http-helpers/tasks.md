## 1. Shared `httpclient` package

- [ ] 1.1 Create `internal/external/httpclient` with `Client`, `New`, `NewRequest`, `CheckStatus`, `Get[T]`, `GetPage[T]`, `Put` as designed in design.md, preserving the exact existing error format strings (`"unexpected status code %d: %s"`, `"failed to decode response: %w"`, `"failed to create request: %w"`, `"failed to marshal request body: %w"`) and verify `go build ./...` succeeds.
- [ ] 1.2 Add unit tests for the new package using `httptest.Server`, covering: successful GET decode, successful paginated GET decode, non-2xx status on GET/PUT, JSON decode failure, request-marshal failure — verify `go test ./internal/external/httpclient/...` passes and asserted error strings match the format above exactly.

## 2. Migrate `internal/external/radarr`

- [ ] 2.1 Refactor `radarr.Client`/`radarr.Config` to hold a `*httpclient.Client` internally instead of the four duplicated fields, and rewrite `getMovies`, `getPagedMovies`, `getMovie`, `putMovie`, `newRequest` to delegate to `httpclient.Get`/`GetPage`/`Put`/`NewRequest`, keeping every exported method's signature, `retry.Do` call, and `apperrors.ExternalServiceError` wrapping unchanged — verify `go build ./...` succeeds and `radarr.Client`/`radarr.Config`'s exported shape is unchanged (same public fields/methods as before).
- [ ] 2.2 Run `go test ./internal/external/radarr/...` and confirm every existing test passes unmodified (no test file edits in this step).

## 3. Migrate `internal/external/sonarr`

- [ ] 3.1 Apply the same refactor as 2.1 to `sonarr.Client`/`sonarr.Config` (`getSeries`, `getSingleSeries`, `getEpisodes`, `getEpisodeList`, `getEpisode`, `putEpisode`, `newRequest`) — verify `go build ./...` succeeds and `sonarr.Client`/`sonarr.Config`'s exported shape is unchanged.
- [ ] 3.2 Run `go test ./internal/external/sonarr/...` and confirm every existing test passes unmodified.

## 4. Add `respondError` helper in `internal/api`

- [ ] 4.1 Add `respondError(c *gin.Context, status int, code, message string)` (wrapping the existing `c.JSON(status, ErrorResponse{Error: code, Message: message})`) to `internal/api`, with no call sites migrated yet — verify `go build ./...` succeeds.
- [ ] 4.2 Run `go test ./internal/api/...` once before migrating any call site, to establish a clean baseline to diff test results against after each subsequent file migration.

## 5. Migrate `internal/api` error call sites, one file at a time

- [ ] 5.1 Replace every `c.JSON(status, ErrorResponse{Error: "...", Message: "..."})` in `handlers.go` with the equivalent `respondError(c, status, "...", "...")` call (same status/code/message, no behavior change) — verify `go test ./internal/api/...` passes with identical results to the 4.2 baseline.
- [ ] 5.2 Same migration for `handlers_frontend.go` — verify `go test ./internal/api/...` still matches the baseline.
- [ ] 5.3 Same migration for `radarr_sonarr_handlers.go` — verify `go test ./internal/api/...` still matches the baseline.
- [ ] 5.4 Same migration for `force_download.go` — verify `go test ./internal/api/...` still matches the baseline.
- [ ] 5.5 Same migration for `resync_path.go` — verify `go test ./internal/api/...` still matches the baseline.

## 6. Final verification

- [ ] 6.1 Run `go test ./...` and confirm the full suite passes.
- [ ] 6.2 Run `grep -rn "ErrorResponse{" internal/api --include=*.go | grep -v _test.go` and confirm the only remaining match is inside `respondError`'s own definition (i.e. every call site was migrated).
- [ ] 6.3 Run `grep -rn "func (c \*Client) newRequest" internal/external/radarr internal/external/sonarr` and confirm no duplicated `newRequest`/status-check/decode logic remains in either file (only calls into `httpclient`).
