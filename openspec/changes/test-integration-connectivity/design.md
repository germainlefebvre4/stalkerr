## Context

The reachability-and-auth classification this feature needs already exists, built for the aggregated "Système" tab:

- `internal/api/system_status.go:257` `classifyReachabilityError(err error) string` maps an error to `unreachable` / `unauthorized` / `timeout` / `unavailable` / `circuit_open`, checking (in order) circuit-breaker error, `net.Error.Timeout()`, a `StatusCoder`-shaped error (401 → `unauthorized`, else → `unavailable`), then `net.OpError`/`net.DNSError` → `unreachable`.
- `internal/external/radarr/radarr.go:259` and `internal/external/sonarr/sonarr.go:365` each already expose `SystemStatus(ctx) error`, an unretried, breaker-aware ping against `/api/v3/system/status`, returning a `*StatusError{Code, Body}` satisfying the `statusCoder` interface (`StatusCode() int`) the classifier checks for.
- `internal/external/jellyfin/jellyfin.go` has no equivalent today: its client only wraps `NotifyPathsUpdated` (`/Library/Media/Updated`), uses a raw `*http.Client` with no circuit breaker, and has no `StatusError`/`SystemStatus` method.
- `internal/api/system_status.go:191-233` builds a fresh Radarr/Sonarr client per request from `settings.Effective()`, passing the server's long-lived `s.radarrBreaker`/`s.sonarrBreaker` (`internal/api/api.go:80-81`) — shared with every other Radarr/Sonarr call in the app.
- `SettingsGroupCard.tsx` (frontend) already tracks a per-card `pending` draft map and a dirty-state save bar (`hasPending`, `handleSave`/`handleDiscard`, `SettingsGroupCard.tsx:56-101`); sensitive fields draft as `''` until the user retypes them (`draftFromField`, `SettingsGroupCard.tsx:23-25`).

See `proposal.md` for why this needs to become an on-demand, per-card, pre-save action for Radarr/Sonarr/Jellyfin.

## Goals / Non-Goals

**Goals:**
- One new backend endpoint, dispatching by service internally, that checks caller-supplied `(url, api_key)` values live and returns the existing `ok`/`ko`+reason shape.
- A "Test connection" action on the Radarr, Sonarr, and Jellyfin cards that uses this endpoint with the card's current in-form values.
- A new `SystemStatus`-equivalent check on the Jellyfin client, matching the Radarr/Sonarr shape closely enough to share the same classifier.

**Non-Goals:**
- TMDB is not touched by this change (see proposal.md).
- The endpoint never returns `not_configured` or `circuit_open`: input validation requires a non-empty URL (so "not configured" can't occur), and the check deliberately bypasses the shared breaker (so "circuit open" can't occur either). The existing frontend `systemStatus.reasons` translation set already covers `unreachable`/`unauthorized`/`timeout`/`unavailable`; no new translation keys are needed for reasons this endpoint can produce, and the pre-existing gap where `circuit_open` has no French translation (used only by `/api/v1/system/status`) is out of scope here.
- No change to `/api/v1/system/status`, its response shape, or the "Système" tab.
- No change to how settings overrides are stored, read, or cleared.

## Decisions

**Single generic endpoint with an internal switch**, e.g. `POST /api/v1/settings/integrations/test` with body `{"service": "radarr"|"sonarr"|"jellyfin", "url": string, "api_key": string}` and response `{"status": "ok"|"ko", "reason"?: string}` (the existing `ServiceStatus` shape). A handler in `internal/api` validates `service` and non-empty `url`, then dispatches to one of three unexported check functions mirroring `checkRadarrStatus`/`checkSonarrStatus`/`checkTMDBStatus` (`system_status.go:191-252`), but built from the request body instead of `settings.Effective()`, and reusing the same `classifyReachabilityError`.
Alternative considered: three separate routes (`/radarr/test`, `/sonarr/test`, `/jellyfin/test`). Rejected per explicit direction — one endpoint, less duplication, same internal switch either way.

**No circuit breaker on the per-request client.** The throwaway `radarr.New(...)`/`sonarr.New(...)` client built for a test call is constructed with `Breaker: nil` (already documented as "nil means unprotected" on `radarr.Config.Breaker`) instead of `s.radarrBreaker`/`s.sonarrBreaker`. This satisfies the `integration-connectivity-test` "independent of the shared breaker" requirement and the `radarr-sonarr-resilience` delta with no new bypass logic — it falls directly out of the existing `Config.Breaker` nil-check in `httpclient.Client`. `RetryConfig: retry.Config{MaxAttempts: 1}` and the existing `systemStatusCheckTimeout` (5s) constant are reused so a test call fails fast like the aggregated status check does.

**Jellyfin check hits `/System/Info` (not `/System/Info/Public`).** Jellyfin's `/System/Info/Public` is deliberately unauthenticated (used for pre-login server discovery) and would validate reachability but never credentials. `/System/Info` requires the `X-Emby-Token` header and returns 401 on an invalid key — the same ping-plus-auth shape Radarr/Sonarr's `/api/v3/system/status` already gives us. Add `SystemStatus(ctx) error` and a `StatusError{Code, Body}` (with `StatusCode() int`) to `internal/external/jellyfin/jellyfin.go`, matching `radarr.go`'s method shape closely enough that `checkJellyfinConnectivity` can reuse `classifyReachabilityError` unchanged. Since the Jellyfin client has no circuit breaker at all today, there is nothing to bypass.

**No fallback to the saved API key.** The test request body carries exactly what the frontend form currently shows: for `url`, that's the pending edit or the original value (always visible); for `api_key`, that's whatever the user retyped, or empty if they didn't. The backend does not merge in the currently-saved secret for an omitted `api_key`. This was an explicit product decision (validating the real saved secret is the user's responsibility, by retyping it) and keeps the endpoint stateless with respect to settings storage — it never needs to read `settings.Effective()` at all.

**Frontend: `SettingsGroupCard` takes an optional `testConfig` prop.** `IntegrationsSection.tsx` passes `testConfig={{service: 'radarr', urlKey: 'radarr.url', apiKeyKey: 'radarr.api_key'}}` (and similarly for `sonarr`/`jellyfin`); the TMDB card passes nothing. When present and `hasPending` is true, the card renders a "Tester la connexion" button next to Annuler/Enregistrer, disabled when the resolved `url` value is empty. Clicking it resolves the current value of `urlKey`/`apiKeyKey` the same way `handleSave` already resolves pending values (pending edit if present, else the field's current draft, where a sensitive field's untouched draft is `''`), calls a new `api.testIntegration(service, url, apiKey)`, and stores `{state: 'idle'|'testing'|'ok'|'ko', reason?}` in local state, rendered as a badge. The result clears on the next edit to that card (same lifecycle as the existing `error` state already does on `handleChange`).

## Risks / Trade-offs

- **Sending an untyped/blank API key can read as a false "unauthorized."** If the user only changed the URL and clicks Test without retyping the key, the check runs with no credentials and will likely report `unauthorized` even though the saved key is fine. → Accepted per explicit product direction; mitigated with a short hint under the button ("La clé API doit être ressaisie pour être testée") so the result isn't misread as "the saved key is broken."
- **A blank/garbage URL from a fast typist could still slip through before the disable check re-renders.** → The backend's own input validation (non-empty URL required) is the real backstop; the frontend disable is a UX nicety, not the safety net.
- **Reusing `classifyReachabilityError` requires the new handler to live in (or import from) `internal/api`.** It is currently unexported. → Keep the new handler in `internal/api` (same package) rather than exporting the classifier; no package boundary to cross.
- **Manual testing bypasses the breaker, so a determined user could hammer a genuinely-down Radarr/Sonarr instance with test clicks with no backoff.** → Acceptable: it's an explicit, user-initiated, single-shot diagnostic action (button disables itself while a test is in flight), not background/automatic traffic; the shared breaker still protects every other (automatic) call path unchanged.

## Migration Plan

Purely additive: one new backend route, one new Jellyfin client method, new frontend UI on existing cards. No data migration, no settings schema change, no rollback complexity beyond reverting the commit(s).
