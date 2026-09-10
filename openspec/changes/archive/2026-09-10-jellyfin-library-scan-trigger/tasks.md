## 1. Configuration

- [x] 1.1 Add `JellyfinConfig{URL, APIKey, Enabled}` to `internal/config/config.go` (mapstructure tags, `Enabled` defaults to `false`) and verify `make build` succeeds and a unit test confirms the zero-value/default when the section is unset in config.
- [x] 1.2 Add a documented `jellyfin` section (url/api_key/enabled) to `config.yml.example` and `.env.example`, matching the style of the existing `radarr`/`sonarr` entries; verify by visual diff against those existing sections.

## 2. Jellyfin client

- [x] 2.1 Create `internal/external/jellyfin` with a `Client`/`Config` (`BaseURL`, `APIKey`, `Timeout`, `RetryConfig`, `Logger`) modeled on `internal/external/radarr`; verify `go build ./internal/external/jellyfin/...` succeeds.
- [x] 2.2 Implement `NotifyPathsUpdated(ctx, paths []string) error` sending `POST {BaseURL}/Library/Media/Updated` with an `X-Emby-Token` header and a `{"Updates":[{"Path":p,"UpdateType":"Created"}, ...]}` body, using `retry.Do`/`apperrors.IsRetryable` and a short bounded timeout; verify with a unit test against an `httptest.Server` asserting method, path, header, and body shape.
- [x] 2.3 Add unit tests for `NotifyPathsUpdated` covering a non-2xx response, a network/timeout error, and an empty `paths` slice (should be a no-op, no request sent); verify via `go test ./internal/external/jellyfin/...`.

## 3. `download` command integration

- [x] 3.1 Extend `downloadStats` in `cmd/download.go` with a mutex-guarded collector that appends `filepath.Clean(filepath.Dir(result.FilePath))` for each item `downloadItem` completes successfully, alongside the existing `recordItem` counter; verify with `go test -race ./cmd/...` exercising concurrent successful completions.
- [x] 3.2 After `runDownloadWorkerPool` returns, deduplicate the collected paths into a set and, only when the set is non-empty and `cfg.Jellyfin.Enabled` with a non-empty `URL`, call `jellyfinClient.NotifyPathsUpdated` once, logging a warning (never failing the command) on error; verify against a mock Jellyfin `httptest.Server` asserting exactly one call carrying the expected deduplicated paths.
- [x] 3.3 Add tests confirming: a run with zero successful downloads sends no notification; a run with Jellyfin disabled or unconfigured attempts no network call; verify via `go test ./cmd/...`.

## 4. `resume-downloads` command integration

- [x] 4.1 In `ResumeHelper.ResumeDownloads`, collect `filepath.Clean(filepath.Dir(result.Result.FilePath))` for each successful result inside the existing `for result := range results` loop (no extra synchronization needed); verify via a unit test that the collected set contains only successfully-resumed items' paths, excluding failed ones.
- [x] 4.2 After that loop, deduplicate and send one `NotifyPathsUpdated` call when the set is non-empty and Jellyfin is enabled+configured, mirroring the `download` command's gating and logging-only failure handling; verify via a test with a mock Jellyfin server.
- [x] 4.3 Verify `stalkeer resume-downloads --dry-run` never attempts a Jellyfin call, since dry-run completes no items; add/confirm a test for this.

## 5. Cross-cutting verification

- [x] 5.1 Add a test (in `cmd/` and/or `internal/downloader/`) covering the spec's "multiple episodes of the same season in one run" scenario, asserting the season's folder appears exactly once in the single notification sent for that run.
- [x] 5.2 Run `make lint` and `make test` for the full suite and confirm no regressions.
