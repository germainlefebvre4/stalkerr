## Context

`download`, `process`, and `m3u-download` are three separate short-lived CLI invocations (`cmd/download.go`, `cmd/process.go`, `cmd/m3u.go`), each triggered independently by cron (Helm `CronJob` or docker-compose sidecar). There is no shared long-lived process between them apart from the API server, which is not involved in any of these runs. Each command already computes the exact data needed for the triggers in the spec at a single, well-defined point:

- `cmd/download.go`: `stats := runDownloadWorkerPool(...)` (run summary) and the `scheduler.BuildStreams` error branch (structural failure).
- `cmd/process.go`: `stats, err := proc.Process(opts)` (both the error branch and the `stats.Errors > 0` branch).
- `cmd/m3u.go`: the `dl.DownloadAndArchive(...)` error branch.

See proposal.md for motivation and the exact trigger list.

## Goals / Non-Goals

**Goals:**
- A minimal, channel-agnostic notification abstraction wired into the 5 trigger points from the spec.
- A working `ntfy` channel end to end: configuration in → HTTP delivery out.
- Delivery failure never changes a command's exit code, output, or side effects.

**Non-Goals:**
- Multi-channel fan-out (Discord/Telegram/generic webhook) — the interface must not preclude adding these later, but none ship in this change.
- Cross-run state: deduplication, rate-limiting, or "don't repeat this alert" logic. Each run's notification (or silence) stands alone.
- Configurable/templated message formats. Message content is fixed in code for v1.
- Helm chart / docker-compose wiring of the new config fields — tracked as an implementation task, not a design concern (no new resource types or branching logic in the charts).

## Decisions

**Notifier interface stays single-method.**
```go
type Notifier interface {
    Notify(ctx context.Context, event Event) error
}
type Event struct {
    Severity Severity // Warning | Critical
    Title    string
    Message  string
}
```
Each implementation is responsible for its own formatting (ntfy priority/tags, a future Discord embed, etc.) — call sites never branch on channel type. Considered giving `Notify` richer, channel-aware parameters (e.g. structured fields per trigger type); rejected because it would leak per-channel formatting concerns into `cmd/*.go`, and the current triggers only ever need a title + short human-readable body.

**A single `notifier.FromConfig(cfg.Notifications) Notifier` constructor, used identically by all three commands.**
Each command already loads `cfg := config.Get()` near the top of its `Run` func. `FromConfig` centralizes the "is this enabled and correctly configured" check once, instead of duplicating it three times. When notifications are disabled, or enabled but missing required ntfy fields (server/topic), it returns a `noopNotifier` whose `Notify` is a no-op returning `nil` — so call sites in `cmd/*.go` never need a nil check or an `if notificationsEnabled` branch; they always just call `notif.Notify(...)`.

**Call sites log-and-discard the returned error; they never branch on it.**
```go
if err := notif.Notify(ctx, event); err != nil {
    log.WithFields(map[string]interface{}{"error": err}).Warn("failed to send notification")
}
```
This is the only place delivery errors surface, matching the spec requirement that delivery failure never changes command behavior. The `ntfy` implementation itself performs its own bounded retry (via the existing `internal/retry` package — same `Config` shape already used by the Radarr/Sonarr/TMDB clients) before giving up and returning an error; call sites do not retry.

**Plain `net/http` POST, no ntfy SDK.**
ntfy's publish API is a single HTTP POST with the message as the body and `Title`/`Priority`/`Tags`/`Authorization` as headers. This matches how `internal/external/{radarr,sonarr,tmdb}` already talk to their APIs (hand-rolled `net/http` clients, no SDK dependency), so the ntfy implementation follows the same shape instead of introducing a new dependency.

**Severity maps to two ntfy priority/tag pairs.**
`Warning` (run-with-failures summaries) → default priority, `warning` tag. `Critical` (structural failures: Radarr/Sonarr unreachable, playlist fetch failure, M3U parse failure) → high priority, `rotating_light` tag. This is a cosmetic mapping with no bearing on the spec and can be adjusted freely during implementation.

**Secrets follow the existing API-key convention.**
The optional ntfy auth token is treated like `RadarrConfig.APIKey`/`SonarrConfig.APIKey`/`TMDBConfig.APIKey`: no real default in `config.yml.example`, overridden via `STALKEER_NOTIFICATIONS_NTFY_AUTH_TOKEN`. Consistent handling, no new secret-management pattern introduced.

## Risks / Trade-offs

- [Risk] A slow or unreachable ntfy server adds latency at the very end of a cron job → Mitigation: short HTTP timeout (a few seconds) and a small bounded retry count via `internal/retry`; `Notify` is always the last thing a command does, after its real work is already complete or already failed.
- [Risk] A typo'd topic or server URL silently means "no notification ever arrives" even though config says enabled → Mitigation: `FromConfig` only checks that required fields are *present*, not that they're *correct* — logging a one-line "notifications enabled" confirmation at startup (with server/topic, never the auth token) at least lets an operator sanity-check the target from the job's log output. Validating the topic actually exists is not attempted.
- [Risk] Introducing a shared `internal/notifier` package touched by three otherwise-independent commands is a small coupling increase → Mitigation: the package has zero dependencies on `cmd/download.go`/`cmd/process.go`/`cmd/m3u.go` internals; it only depends on `internal/config` and `internal/retry`, both already shared across the codebase the same way.

## Migration Plan

Purely additive and opt-in (`notifications.enabled` defaults to `false`). No existing behavior changes until an operator sets it to `true` and provides ntfy server/topic. No data migration, no rollback beyond reverting the config flag.
