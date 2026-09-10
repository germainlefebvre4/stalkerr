## 1. Backend: reachability checks per integration

- [ ] 1.1 Add a `SystemStatus`/reachability method to `internal/external/radarr.Client` calling `GET /api/v3/system/status` under a caller-supplied short context timeout; verify with a unit test covering success, timeout, connection-refused, and 401 cases (table-driven, following `radarr_test.go` conventions)
- [ ] 1.2 Add the equivalent reachability method to `internal/external/sonarr.Client` calling `GET /api/v3/system/status`; verify with the same set of unit test cases
- [ ] 1.3 Add a reachability method to `internal/external/tmdb.Client` calling `GET /3/configuration`; verify with a unit test covering success, timeout, and 401 cases
- [ ] 1.4 Introduce the small fixed set of KO reason codes (`unreachable`, `unauthorized`, `timeout`, `unavailable`) and a shared classifier mapping each client's error into one of them; verify with unit tests covering each mapped error type

## 2. Backend: disk usage deduplication

- [ ] 2.1 Add a function that resolves a configured path to its nearest existing ancestor and returns that ancestor's device id (`os.Stat(...).Sys().(*syscall.Stat_t).Dev`), reusing the walk-up logic already in `GetDiskSpace`; verify with a unit test asserting two paths on the same filesystem yield the same device id
- [ ] 2.2 Add a grouping function that takes the app's configured storage paths (movies path, TV shows path, temp dir if set, M3U archive dir) and returns deduplicated disk usage entries (one `GetDiskSpace` call per distinct device id, each entry listing the paths it backs); verify with a unit test covering both a shared-volume case and a distinct-volumes case
- [ ] 2.3 Handle a missing/unreadable configured path as an unavailable entry with a short reason rather than an error, and verify with a unit test

## 3. Backend: aggregation endpoint

- [ ] 3.1 Add `GET /api/v1/system/status` route and handler in `internal/api` that runs the DB health check, Radarr check, Sonarr check, TMDB check, and disk usage grouping concurrently, each independently, following the existing per-request client construction convention from `radarr_sonarr_handlers.go` (checking `cfg.Radarr.URL`/`cfg.Radarr.APIKey` etc. for "not configured") and reusing `Server.tmdbClient` for TMDB
- [ ] 3.2 Verify with an integration-style test (`internal/api` test conventions) that one dependency failing (e.g. Radarr unreachable) still returns HTTP 200 with correct status for the other sections
- [ ] 3.3 Verify with a test that a disabled integration reports `not configured` without an outbound call being attempted (e.g. via a client/mux that fails the test if hit)
- [ ] 3.4 Verify with a test that the endpoint's total response time stays bounded when a dependency's check exceeds the per-check timeout

## 4. Frontend: data fetching

- [ ] 4.1 Add `api.getSystemStatus()` to `frontend/src/services/api.ts` following the existing service method conventions, and add the corresponding response type to `frontend/src/types`
- [ ] 4.2 Add a `useSystemStatus` hook (`frontend/src/hooks`) exposing an explicit `fetchStatus`/`refresh` action and loading/error state, with no automatic fetch on mount and no polling interval; verify with a hook test asserting no request fires until the action is invoked

## 5. Frontend: header entry point and dialog

- [ ] 5.1 Add the system status icon button to `FloatingHeader.tsx`, positioned next to the language selector, rendered identically regardless of viewport width, with no badge/indicator element
- [ ] 5.2 Create `SystemStatusDialog.tsx` following the existing `*Dialog.tsx` (Radix Dialog) pattern, triggered by the header icon, calling `useSystemStatus`'s fetch action on open
- [ ] 5.3 Render the DB/Radarr/Sonarr/TMDB rows with visually distinct OK/KO/not-configured states, showing the translated short reason on KO rows
- [ ] 5.4 Render the deduplicated disk usage entries (used/available space per volume)
- [ ] 5.5 Add a refresh action inside the dialog that re-invokes the fetch and updates the displayed rows without closing the dialog
- [ ] 5.6 Verify with a component test (following `RadarrSonarrTab.test.tsx`/`DownloadsTab.test.tsx` conventions) that opening the dialog triggers exactly one fetch, reopening after close triggers a new one, and no fetch occurs before first open

## 6. i18n

- [ ] 6.1 Add English and French translation keys for the header icon's label, the dialog's title/rows/states/reasons, and the refresh action to `frontend/src/locales/{en,fr}`; verify no missing-key warnings when the dialog is opened in both languages

## 7. Verification

- [ ] 7.1 Run `make test` and confirm all Go tests pass, including the new backend tests
- [ ] 7.2 Run `npm test` in `frontend/` and confirm all tests pass, including the new frontend tests
- [ ] 7.3 Manually exercise the dialog against a local stack with at least one integration disabled and one misconfigured (bad API key), confirming the not-configured and KO-with-reason states render as expected on both desktop and mobile widths
