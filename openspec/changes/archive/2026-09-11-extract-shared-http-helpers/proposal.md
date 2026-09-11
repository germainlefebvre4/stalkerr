## Why

A full code-quality audit of the codebase (2026-09-04) found that HTTP request/response boilerplate is duplicated extensively in two places: `internal/external/radarr` and `internal/external/sonarr` share ~90% identical request-building, status-code-checking, and client-bootstrap code, and `internal/api` hand-writes the same "build an error JSON response" literal at roughly 150 call sites across its handler files. This duplication means a bugfix or behavior tweak (e.g. how a non-2xx response is handled, or what an error payload looks like) has to be applied correctly at every copy, and it makes adding the next external client or handler slower and more error-prone than it should be. Consolidating both into shared helpers now, while both areas are simple and well-tested, is low-risk and removes the single largest source of repetition identified in the audit.

## What Changes

- Add a small shared internal HTTP client helper used by both `internal/external/radarr` and `internal/external/sonarr`: common request construction, status-code checking, and a generic JSON-decoding GET (and PUT, where both clients need it). Refactor `radarr.go` and `sonarr.go` to call it instead of their independent, near-identical implementations.
- Add a shared `respondError(c, status, code, message)`-style helper in `internal/api` and route the existing ad-hoc `ErrorResponse{...}` + `c.JSON(...)` literals in `handlers.go`, `handlers_frontend.go`, `radarr_sonarr_handlers.go`, `force_download.go`, and `resync_path.go` through it.
- **No behavior change**: every existing response status code, error code string, and error message stays exactly as it is today. Every existing Radarr/Sonarr request shape and response handling stays exactly as it is today. This is a pure internal refactor — existing tests should continue to pass unmodified, plus new unit tests for the extracted helpers themselves.
- Out of scope for this change (left for later, separate changes): adding a circuit breaker to radarr/sonarr, adopting `internal/apperrors` as the error taxonomy behind `respondError`, normalizing any *currently inconsistent* status codes/shapes across handlers, and injecting the DB dependency instead of using the `database.Get()` global. Those are real follow-ups from the same audit but are separate, larger-blast-radius decisions.

## Capabilities

This is a pure internal refactor: no user-observable or API-observable behavior changes, so no capability specs are added or modified (`skip_specs: true` is set for this change).

### New Capabilities
(none)

### Modified Capabilities
(none)

## Impact

- **Code**: `internal/external/radarr/radarr.go`, `internal/external/sonarr/sonarr.go` (new shared helper, likely a small new file/package such as `internal/external/httpclient`), `internal/api/handlers.go`, `internal/api/handlers_frontend.go`, `internal/api/radarr_sonarr_handlers.go`, `internal/api/force_download.go`, `internal/api/resync_path.go` (new shared `respondError` helper).
- **Tests**: `internal/external/radarr/radarr_test.go`, `internal/external/sonarr/sonarr_test.go`, and the `internal/api/*_test.go` files must keep passing unchanged, since response bodies/status codes are not changing; new unit tests are added for the extracted helpers.
- **No** DB migrations, no config changes, no changes to the Radarr/Sonarr/TMDB external contracts, no frontend changes required (API responses are byte-for-byte the same).
