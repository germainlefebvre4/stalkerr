## Why

The Configuration page's "Intégrations" tab lets a user edit Radarr, Sonarr, TMDB, and Jellyfin connection settings, but gives no feedback on whether those values actually work until a background sync job fails later (often silently, or only visible in logs). The user has to save blind, then separately visit the "Système" tab to see aggregated reachability — which only reflects the last *saved* config, not what they just typed, and covers Radarr/Sonarr/TMDB but not Jellyfin.

## What Changes

- Each of the Radarr, Sonarr, and Jellyfin cards in the Intégrations tab gains a "Tester la connexion" action, appearing in the card's action bar next to Enregistrer/Annuler whenever that card has an unsaved change (same trigger as the existing save bar).
- The test action sends the card's current in-progress field values (url, api_key, enabled) exactly as shown in the form — not the previously saved values, and with no fallback to a saved secret. If a sensitive field (api_key) has not been retyped since it was masked, it is sent empty; validating that field with its real value is the user's responsibility.
- A new backend endpoint performs a live, on-demand reachability-and-authentication check against the target service using those caller-supplied values and returns a result without persisting anything.
- The result distinguishes: reachable and authenticated (OK), unreachable (network/DNS failure), invalid credentials (HTTP 401), timed out, or another failure — reusing the OK/KO + reason vocabulary already established by the system status feature (`system-status-api`), so the user sees e.g. "KO: Identifiants invalides" instead of a generic failure.
- The check bypasses that service's shared circuit breaker: a manual test always attempts a live call regardless of unrelated recent outages, and never counts as a production failure toward that breaker.
- Jellyfin gains a lightweight authenticated reachability check at the client level (new capability — Radarr and Sonarr already have one via their existing `SystemStatus` methods; Jellyfin's client has none today).
- TMDB is explicitly out of scope: its client is instantiated once at server boot and a save does not take effect until restart, which would make a test result misleading; that gap is left for separate consideration.
- **BREAKING**: none.

## Capabilities

### New Capabilities
- `integration-connectivity-test`: on-demand backend check that a caller-supplied Radarr, Sonarr, or Jellyfin URL + API key combination is reachable and authenticates, independent of any saved configuration, returning an OK/KO+reason result without touching settings storage.

### Modified Capabilities
- `frontend-app-settings-management`: the Radarr, Sonarr, and Jellyfin cards gain a "Test connection" action, shown only while the card has an unsaved change, that tests the in-progress (not necessarily saved) field values and displays a detailed OK/KO+reason result badge on the card.
- `radarr-sonarr-resilience`: the on-demand connectivity-test call is explicitly exempted from the shared circuit breaker requirement (every other call to that service continues to pass through it), so a manual test is never blocked by, and never itself trips, that breaker.

## Impact

- Backend: one new generic endpoint (dispatches by service via an internal switch) in `internal/api`; a new `SystemStatus`-equivalent check + `StatusError` type added to `internal/external/jellyfin/jellyfin.go`, mirroring the existing pattern in `radarr.go`/`sonarr.go`; reuses `internal/api/system_status.go`'s `classifyReachabilityError` reason vocabulary rather than inventing a new one.
- Frontend: the Radarr/Sonarr/Jellyfin card component gains a "Test connection" action and result badge; `frontend/src/services/api.ts` gains one new call; locale files gain new strings for the action label and result reasons.
- No changes to settings storage, override persistence, the `/api/v1/system/status` endpoint or response shape, or the Configuration page's tab structure.
- Out of scope: TMDB connectivity testing, any change to how overrides are saved/cleared, any change to the "Système" tab.
