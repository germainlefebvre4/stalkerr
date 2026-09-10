## 1. Notification core

- [ ] 1.1 Add `internal/notifier` package with `Severity` (`Warning`, `Critical`), `Event{Severity, Title, Message}`, and the `Notifier` interface (`Notify(ctx context.Context, event Event) error`); verify it compiles with no implementations yet.
- [ ] 1.2 Add a `noopNotifier` implementing `Notifier` with a `Notify` that always returns `nil`; add a unit test asserting it never errors and does nothing observable.
- [ ] 1.3 Add the `ntfy` `Notifier` implementation: plain `net/http` POST to `{server}/{topic}` with `Title`, `Priority`, `Tags` headers set from `Event.Severity` (Warning → default priority/`warning` tag, Critical → high priority/`rotating_light` tag) and an optional `Authorization` header when an auth token is configured; verify with a unit test using `httptest.Server` covering a successful post and the request headers/body sent.
- [ ] 1.4 Wrap the ntfy HTTP call with `internal/retry` (reuse `retry.Config`, small `MaxAttempts`, short backoff) so a transient failure retries before `Notify` returns an error; verify with a unit test that a server failing once then succeeding results in a delivered notification, and one where all attempts fail returns a non-nil error.
- [ ] 1.5 Add `notifier.FromConfig(cfg config.NotificationsConfig) Notifier` that returns the `ntfy` implementation when enabled with server+topic set, and `noopNotifier` otherwise (disabled, or enabled but missing required fields); verify with unit tests covering all three cases.

## 2. Configuration

- [ ] 2.1 Add `NotificationsConfig{Enabled bool; Ntfy NtfyConfig{Enabled bool; ServerURL string; Topic string; AuthToken string}}` to `internal/config/config.go` following the existing `mapstructure` tag style used by `RadarrConfig`/`TMDBConfig`; verify `internal/config/config_test.go` still passes and add a case covering defaults (disabled) when the section is absent.
- [ ] 2.2 Register `viper.BindEnv` calls for the new keys in `Load()` alongside the existing bindings (e.g. `notifications.enabled`, `notifications.ntfy.server_url`, `notifications.ntfy.topic`, `notifications.ntfy.auth_token`), matching the `STALKEER_NOTIFICATIONS_NTFY_AUTH_TOKEN`-style env var naming already used for other API keys; verify with a config test that env vars override file values.
- [ ] 2.3 Document the new section in `config.yml.example` and `.env.example` with `enabled: false` and empty/placeholder values (no real token committed), matching how the Radarr/Sonarr/TMDB sections are documented.

## 3. Wire into `download`

- [ ] 3.1 In `cmd/download.go`, construct `notif := notifier.FromConfig(cfg.Notifications)` once near the top of `downloadCmd.Run`, alongside the existing config/logger setup.
- [ ] 3.2 Extract a small helper (e.g. `notifyBuildStreamsFailure(notif, err)`) called from the existing `scheduler.BuildStreams` error branch (before `os.Exit(1)`), sending a `Critical` event identifying the Radarr/Sonarr reachability failure; verify with a unit test on the helper (fake `Notifier` capturing the event) that it builds the expected `Event` for a given error.
- [ ] 3.3 Extract a small helper (e.g. `notifyDownloadRunResult(notif, stats *downloadStats)`) called after `runDownloadWorkerPool` returns, sending a `Warning` event only when `stats.Failed > 0` (never when 0), including total/downloaded/failed counts in the message; verify with unit tests covering both the "no failures → no call" and "failures → event sent with correct counts" cases.

## 4. Wire into `process`

- [ ] 4.1 In `cmd/process.go`, construct `notif := notifier.FromConfig(cfg.Notifications)` near the top of `processCmd.Run`.
- [ ] 4.2 Extract a small helper (e.g. `notifyProcessParseFailure(notif, err)`) called from the existing `proc.Process(opts)` error branch (before `os.Exit(1)`), sending a `Critical` event identifying the M3U parse failure; verify with a unit test on the helper.
- [ ] 4.3 Extract a small helper (e.g. `notifyProcessRunResult(notif, stats *processor.Statistics)`) called after a successful `Process()` call, sending a `Warning` event only when `stats.Errors > 0`, including processed/error counts; verify with unit tests covering the zero-errors and errors-present cases.

## 5. Wire into `m3u-download`

- [ ] 5.1 In `cmd/m3u.go`, construct `notif := notifier.FromConfig(cfg.Notifications)` near the top of the `m3u-download` command's `Run`.
- [ ] 5.2 Extract a small helper (e.g. `notifyPlaylistFetchFailure(notif, err)`) called from the existing `dl.DownloadAndArchive(...)` error branch (before `os.Exit(1)`), sending a `Critical` event identifying the playlist fetch failure; verify with a unit test on the helper.

## 6. Verification

- [ ] 6.1 Run the full test suite (`go test ./...`) and confirm it passes, including all new notifier/config/cmd tests.
- [ ] 6.2 Run `golangci-lint run` (or the project's configured lint command) and confirm no new findings.
- [ ] 6.3 Manually verify end-to-end delivery against a real ntfy topic (e.g. `ntfy.sh` test topic): configure `notifications.enabled: true` with that server/topic, force one failure in each of `download`, `process`, and `m3u-download` (e.g. temporarily point Radarr/Sonarr/the playlist URL at an unreachable address), and confirm a push notification arrives for each, then confirm a clean run of each command produces no notification.
