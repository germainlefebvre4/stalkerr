## Why

Stalkeer's `download`, `process`, and `m3u-download` commands run unattended on a cron schedule with no persistent daemon watching them. Today the only way to learn a run failed, a playlist fetch broke, or Radarr/Sonarr went unreachable is to open the dashboard or dig through pod logs. A lightweight push notification on failure gives visibility without requiring anyone to check the dashboard proactively.

## What Changes

- Add an `internal/notifier` package with a channel-agnostic `Notifier` interface (`Notify(Event) error`) so future channels (Discord, Telegram, generic webhook) can be added without touching call sites.
- Add an `ntfy` notifier implementation that posts to a configured ntfy server/topic.
- Add a `notifications` configuration section (`internal/config`) following the existing per-integration pattern (`enabled` flag, nested `ntfy` block with server URL, topic, optional auth token).
- Wire notifications into exactly the run outcomes already distinguishable today:
  - `download` run summary, sent only when at least one item failed in that run.
  - `process` run summary, sent only when the run completed with errors.
  - `download` aborting because Radarr/Sonarr could not be reached while building the stream set (already a fatal `os.Exit(1)` today).
  - `process` aborting because the M3U file could not be parsed (already a fatal `os.Exit(1)` today).
  - `m3u-download` aborting because the playlist could not be fetched (already a fatal `os.Exit(1)` today).
- A successful run with zero failures sends nothing (silence is the expected signal).
- A failed notification delivery (e.g. ntfy server unreachable) is logged and swallowed — it never changes a command's exit code or masks the original failure.

Explicitly out of scope for this change (see exploration discussion):
- Per-item permanent-failure notifications — the existing data model doesn't cleanly distinguish "this item will never succeed" from "this run's attempt failed," and no clean signal exists to notify on.
- Disk-space alerts — no code path currently checks disk space before/during a download; wiring that check in is a separate change.
- Distinguishing "TMDB is down" from "one title failed to enrich" — would require inspecting the TMDB client's circuit-breaker state directly, not just aggregate error counts.
- Discord, Telegram, and generic-webhook channels — the `Notifier` interface accommodates them, but only `ntfy` ships now.

## Capabilities

### New Capabilities
- `run-notifications`: notifier abstraction, ntfy channel, notification configuration, and the failure/summary triggers wired into the `download`, `process`, and `m3u-download` commands.

### Modified Capabilities
(none — no existing capability's requirements change; wiring happens inside the commands' existing error/summary paths)

## Impact

- New package `internal/notifier` (interface, `Event` type, `ntfy` implementation, reused `internal/retry` for delivery attempts).
- New config fields in `internal/config/config.go` (`NotificationsConfig`), plus `config.yml.example` and `.env.example` documentation.
- Touches `cmd/download.go`, `cmd/process.go`, `cmd/m3u.go` at their existing summary/error-handling points — no change to command flags, arguments, or exit codes.
- Follow-up (tracked as implementation tasks, not a spec change): surface the new config fields through `charts/stalkerr` (values.yaml/ConfigMap) and `docker-compose.yml` so the feature is configurable in both deployment paths.
