# Tasks

## 1. Upstream time budgets

- [x] 1.1 Define `arrListTimeout` (30 s) and `arrStatsTimeout` (60 s) in `internal/api/radarr_sonarr_handlers.go` with a comment tying them to `WriteTimeout`; verify `go build ./...` succeeds (edit already present in the working tree, not yet built)
- [x] 1.2 Use `arrListTimeout` for the request context and client timeout of the Radarr list, Sonarr list, Radarr match detail and Sonarr episodes handlers, and for both clients in the stats handler, with `arrStatsTimeout` as the stats request context; verify `grep -n existenceCheckTimeout internal/api/radarr_sonarr_handlers.go` returns nothing
- [x] 1.3 Raise `http.Server.WriteTimeout` from 15 s to 65 s in `internal/api/api.go`; verify it is greater than `arrStatsTimeout`
- [x] 1.4 Add a unit test in `internal/api/radarr_sonarr_handlers_test.go` asserting `arrListTimeout > existenceCheckTimeout` and `arrStatsTimeout >= arrListTimeout`, and verify `go test ./internal/api/ -run <test name>` passes

## 2. Failure logging

- [x] 2.1 Log the underlying error with `logger.AppLogger()` when `GetAllMovies` fails in the Radarr list handler and when `GetAllSeries` fails in the Sonarr list handler, keeping the generic 502 response (edit already present in the working tree, not yet built)
- [x] 2.2 Verify the existing upstream-failure handler tests still return the same 502 body, by running `go test ./internal/api/`

## 3. Integration check

- [x] 3.1 Run `go vet ./...` and `go test ./...`, and verify both pass with no change to force-download, resync or duplicate tests (they still use `existenceCheckTimeout`)
- [x] 3.2 Deploy and verify on the arr-suite page that Films, Séries and Résumé load against the NFS-backed instances, and that a forced failure logs the underlying error

## Workflow follow-up

- Archive the change once the review requirements are satisfied.
- The proper fix for the slow instances is infrastructure (move the Radarr/Sonarr config volume off NFS); this change only masks it.
