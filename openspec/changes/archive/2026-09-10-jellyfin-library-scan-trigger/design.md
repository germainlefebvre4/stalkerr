## Context

See proposal.md - Why. Two facts from the existing codebase shape this design:

- `internal/downloader/path.go` (`buildMovieBasePath`/`buildTVShowBasePath`) always writes a movie's files directly under `MoviesPath/Title (Year)/` and an episode's files directly under `TVShowsPath/Series (Year)/Season NN/`. This means `filepath.Dir(finalFilePath)` is already the correct Jellyfin-library-relevant folder for both content types — no content-type branching is needed to compute "the folder to report".
- The two call sites that finish downloads have different concurrency shapes:
  - `cmd/download.go`'s `runDownloadWorkerPool` calls `downloadItem` concurrently from N worker goroutines; today it only maintains a plain counter (`downloadStats.recordItem`).
  - `internal/downloader/resume_helper.go`'s `ResumeDownloads` already consumes results sequentially via `for result := range results` (a channel drained one at a time), so no additional synchronization is needed there.
- `internal/external/radarr` and `internal/external/sonarr` establish the existing pattern for a small external HTTP client (`Config{BaseURL, APIKey, Timeout, RetryConfig, Logger}`, `retry.Do` + `apperrors.IsRetryable`), and `RadarrConfig`/`SonarrConfig` in `internal/config/config.go` establish the pattern for a per-integration config block. This change follows both patterns rather than inventing new ones.
- `internal/api/resync_path.go`'s `computeReconciledPath` recently shipped a fix for comparing paths without `filepath.Clean` first (commit `eb3d251`). The path de-duplication in this change reuses that lesson directly.

## Goals / Non-Goals

**Goals:**
- Reuse the Radarr/Sonarr client and config conventions for a new, independent Jellyfin client.
- Make the notification a thin, best-effort side effect that cannot regress `download`/`resume-downloads` behavior, timing, or exit codes.
- Collect exactly one de-duplicated path per changed movie/season per run, and send exactly one HTTP call per run.

**Non-Goals:**
- Plex or Emby support (proposal.md - Impact: explicitly out of scope; the client is Jellyfin-specific, not behind a multi-provider abstraction, since there is only one provider to support).
- A dashboard/API surface to configure or manually trigger the integration — config.yml/env vars only, matching Radarr/Sonarr.
- Wiring the new config keys into the Helm chart's ConfigMap/Secret (`openspec/specs/configuration-management`) — left for a follow-up change; nothing here prevents someone from setting the env vars manually on a Helm deployment today.
- Notifying on the manual dashboard "move" actions (`moveMovieFolder`/`moveTVShowFolder`) — explicitly deferred per the exploration; only the two CLI commands are in scope.

## Decisions

### 1. New package `internal/external/jellyfin`, modeled on `internal/external/radarr`
A `Client` with `New(Config) *Client` (`BaseURL`, `APIKey`, `Timeout`, `RetryConfig`, `Logger`) and one method:

```go
func (c *Client) NotifyPathsUpdated(ctx context.Context, paths []string) error
```

It builds `POST {BaseURL}/Library/Media/Updated` with body `{"Updates": [{"Path": p, "UpdateType": "Created"}, ...]}` and header `X-Emby-Token: {APIKey}` (the token header Jellyfin accepts for backward compatibility with the Emby API it forked from; confirm against the target Jellyfin version during implementation — see Risks). Uses `retry.Do`/`apperrors.IsRetryable` like the Radarr client, but with a short default timeout (see Decision 4) since this call must never meaningfully delay a CLI run.

Alternative considered: a generic `MediaServerNotifier` interface with a Jellyfin implementation, anticipating Plex/Emby later. Rejected as premature — there is exactly one implementation needed today (per the exploration, Plex was explicitly excluded for its extra section-ID requirement); introducing the interface now would be speculative generality with no second caller to validate its shape.

### 2. `JellyfinConfig` in `internal/config/config.go`, disabled by default
```go
type JellyfinConfig struct {
    URL     string `mapstructure:"url"`
    APIKey  string `mapstructure:"api_key"`
    Enabled bool   `mapstructure:"enabled"`
}
```
Added as `Jellyfin JellyfinConfig `mapstructure:"jellyfin"`` next to `Radarr`/`Sonarr` on the root config struct, with `Enabled` defaulting to `false`. A helper (e.g. `cfg.Jellyfin.IsConfigured()`, or an inline check) gates every call site: enabled AND non-empty `URL` before any network attempt, matching the spec's "Jellyfin integration is opt-in" requirement including its warn-and-skip case for enabled-but-unconfigured.

### 3. Path collection: targeted addition to each command's existing result-handling, not a shared framework
- **`resume-downloads`**: inside the existing `for result := range results` loop in `ResumeHelper.ResumeDownloads`, on success append `filepath.Clean(filepath.Dir(result.Result.FilePath))` to a local `[]string`. No mutex needed (already single-threaded consumption).
- **`download`**: extend `downloadStats` (in `cmd/download.go`) with a mutex-guarded `[]string` (or equivalent thread-safe collector) and have `downloadItem` append `filepath.Clean(filepath.Dir(result.FilePath))` on success, alongside the existing `stats.recordItem(success)` call.
- Both then de-duplicate the collected slice into a set (`map[string]struct{}`) before calling the Jellyfin client, so a run with multiple episodes of the same season reports that season's path exactly once (spec: "One targeted notification per run, not one per item").

Alternative considered: centralizing collection in a shared helper called by both commands. Rejected for this change — the two call sites have different enough shapes (worker pool vs. drained channel) that a shared collector would need its own synchronization abstraction for no real duplication savings beyond the ~5-line dedup-and-notify tail, which is small enough to keep in each command alongside its existing stats handling.

### 4. Notification call site and failure handling
In each command, after the run's work completes (after `wg.Wait()` in `download`; after the `for result := range results` loop in `resume-downloads`) and only if the deduplicated path set is non-empty:
```go
if cfg.Jellyfin.Enabled && cfg.Jellyfin.URL != "" {
    if err := jellyfinClient.NotifyPathsUpdated(ctx, paths); err != nil {
        log.WithFields(...).Warn("failed to notify Jellyfin of library changes", err)
    }
}
```
A short client timeout (proposed default: 10s, configurable only via the client's internal default — not exposed as a new config key, keeping the config surface minimal) bounds the worst case so an unreachable Jellyfin cannot hang the command. The error is logged and swallowed; it never changes `stats`/`ResumeStats` or the process exit code.

### 5. `UpdateType` is always `"Created"`
Jellyfin's `Library/Media/Updated` payload requires an `UpdateType` per path (`Created`, `Modified`, or `Deleted`). Both fresh downloads and resumed/completed downloads are reported as `"Created"` — Jellyfin's own library scan determines whether the item is genuinely new or an update to an existing one; the distinction is not meaningful to compute on the Stalkeer side and Jellyfin behaves correctly either way.

## Risks / Trade-offs

- **[Risk]** Jellyfin's expected auth header/exact payload shape could differ slightly across Jellyfin versions (header name, or the API requiring `/Items/{id}` style refresh instead for some versions) → Mitigation: verify against the target Jellyfin version's API docs or a live instance during implementation/testing (tasks.md includes a verification task); the client isolates this to one file (`internal/external/jellyfin`), so adjusting it later is a small, contained change.
- **[Risk]** A path reported to Jellyfin might not exactly match one of Jellyfin's configured library root folders (e.g. a differently-mounted path inside the Jellyfin container vs. the Stalkeer container) → Mitigation: out of scope to solve generically here; this is a deployment/configuration concern the user already manages for the shared media volume (see `openspec/specs/storage-configuration`), same as it is for Radarr/Sonarr path reconciliation today.
- **[Trade-off]** Collecting paths only from `download` and `resume-downloads` misses library changes made through the dashboard's manual move endpoints (`moveMovieFolder`/`moveTVShowFolder`) → Accepted: explicitly deferred per the exploration; those moves are rarer and can be covered by a later change if needed.
