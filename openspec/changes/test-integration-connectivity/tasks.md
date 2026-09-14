## 1. Jellyfin connectivity check

- [ ] 1.1 Add `StatusError{Code, Body}` (with `StatusCode() int`) and `SystemStatus(ctx context.Context) error` to `internal/external/jellyfin/jellyfin.go`, calling `GET /System/Info` with the `X-Emby-Token` header, mirroring `radarr.go:232-279`'s shape (unretried, returns `*StatusError` on non-200) — verify with a new unit test in `jellyfin_test.go` covering a 200 (nil error), a 401 (`StatusError{Code:401}`), and a connection failure.

## 2. Backend connectivity-test endpoint

- [ ] 2.1 Add a request/response type and handler (e.g. `internal/api/integration_test_handlers.go`): `POST` body `{"service": "radarr"|"sonarr"|"jellyfin", "url": string, "api_key": string}`, response `{"status": "ok"|"ko", "reason"?: string}` — verify the file compiles and the types match `ServiceStatus` in `system_status.go`.
- [ ] 2.2 Validate the request: reject an empty `url` and any `service` other than `radarr`/`sonarr`/`jellyfin` with a 400, without making any outbound call — verify with unit tests for both rejection cases (no HTTP call attempted).
- [ ] 2.3 Implement the `radarr`/`sonarr` branches: build a throwaway client from the request's `url`/`api_key` with `Breaker: nil` and `RetryConfig{MaxAttempts: 1}`, call `SystemStatus(ctx)`, classify the error with the existing `classifyReachabilityError` — verify with unit tests covering OK, 401→`unauthorized`, connection-refused→`unreachable`, and timeout→`timeout`.
- [ ] 2.4 Implement the `jellyfin` branch using the new client method from task 1.1, same classification — verify with an equivalent unit test.
- [ ] 2.5 Register the route (e.g. `v1.POST("/settings/integrations/test", s.testIntegrationConnectivity)`) next to the system status route in `internal/api/api.go` — verify with an integration-style test hitting the route through the router.
- [ ] 2.6 Confirm no settings read/write occurs in this path (no call to `settings.Effective()`, `SetOverride`, or `ClearOverride`) — verify by inspection and by a test asserting stored settings are unchanged after a test call.

## 3. Frontend API client and shared card support

- [ ] 3.1 Add `api.testIntegration(service, url, apiKey)` to `frontend/src/services/api.ts`, posting to the new endpoint and returning the typed `{status, reason?}` result (add the type to `types.ts`, reusing `ServiceStatusState`/`reason` shape) — verify with a unit test mocking `fetch`.
- [ ] 3.2 Add an optional `testConfig?: {service: 'radarr'|'sonarr'|'jellyfin'; urlKey: string; apiKeyKey: string}` prop to `SettingsGroupCard`, resolving the current `url`/`api_key` values the same way `handleSave` resolves pending values (pending edit if present, else the field's current draft) — verify with a component test that the resolved values match what `handleSave` would submit.
- [ ] 3.3 Render a "Tester la connexion" button in the card's action row when `testConfig` is set and `hasPending` is true, disabled when the resolved `url` is empty or a test is in flight — verify with a component test for each visibility/disabled condition.
- [ ] 3.4 On click, call `api.testIntegration`, track `{state: 'idle'|'testing'|'ok'|'ko', reason?}` in local state, render it as a status badge, and clear it on the next field edit in that card — verify with a component test simulating a click, an OK response, a KO+reason response, and a subsequent edit clearing the badge.

## 4. Wiring the three integration cards

- [ ] 4.1 Pass `testConfig` for Radarr (`radarr.url`/`radarr.api_key`), Sonarr (`sonarr.url`/`sonarr.api_key`), and Jellyfin (`jellyfin.url`/`jellyfin.api_key`) from `IntegrationsSection.tsx`; leave the TMDB card without `testConfig` — verify with a test asserting the Test button appears only on the three wired cards.

## 5. Localization

- [ ] 5.1 Add "Tester la connexion" / "Test en cours…" (and English equivalents) to the relevant locale file(s); reuse the existing `dialogs.json` `systemStatus.reasons.*` / `systemStatus.states.*` keys for the badge text (cross-namespace `useTranslation`) rather than duplicating them — verify both `en` and `fr` locale files stay valid JSON and the new keys render in a component test.

## 6. Verification

- [ ] 6.1 Run the full backend test suite (`go test ./...`) and frontend test suite, and confirm both pass with the new code included.
- [ ] 6.2 Manually exercise the feature against a real or stubbed Radarr/Sonarr/Jellyfin instance: correct URL+key (OK), correct URL+wrong key (`unauthorized`), wrong URL (`unreachable`), and confirm the "Système" tab's existing behavior is unaffected.
