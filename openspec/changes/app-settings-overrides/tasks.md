## 1. Data model and migration

- [ ] 1.1 Add `models.SettingsOverride` (`settings_overrides`: `key` unique text primary key, `value` text, `updated_at`) in `internal/models/settings.go`, and verify `go build ./...` succeeds
- [ ] 1.2 Add `models.M3USourceConfig` (`m3u_source_configs`: `name` unique, `file_path`, every `download.*` field, `is_runtime`, timestamps) in `internal/models/m3u_source.go`, mirroring `internal/models/filter.go`, and verify `go build ./...` succeeds
- [ ] 1.3 Register both new models in the `AutoMigrate` call in `internal/database/database.go`, and verify a fresh test database migrates cleanly (existing DB integration tests still pass)

## 2. Relax config boot requirements

- [ ] 2.1 Remove the `m3u.sources` non-empty check from `validate()` in `internal/config/config.go`, and update/add a `config_test.go` case asserting `Load()` succeeds with `m3u.sources` empty or absent
- [ ] 2.2 Verify `database.user`/`database.dbname` (bootstrap) validation is unchanged, via existing `config_test.go` cases

## 3. `internal/settings` package: field registry and effective resolution

- [ ] 3.1 Define the overridable-field registry (dotted key → Go type, `sensitive`, `restart_required`) covering every Radarr/Sonarr/TMDB/Jellyfin/Notifications/Downloads/logging field, per `design.md`'s Decisions, and verify a unit test enumerates the registry and asserts it matches the expected key list
- [ ] 3.2 Add a reflection-based test over `config.Config` asserting every scalar leaf field is either in the registry or on an explicit Tier-0/list-shaped exclusion list, so a future config field can't silently go unregistered
- [ ] 3.3 Implement override storage: get/set/delete a `SettingsOverride` row by key, with values JSON-encoded, and verify unit tests for round-tripping each registered type (string, int, bool, float64)
- [ ] 3.4 Implement `settings.Effective() *config.Config`: start from `config.Get()`, apply every stored override from the registry on top, and verify a unit test that an override changes the returned field while unset fields keep their `config.Get()` value
- [ ] 3.5 Implement per-field origin lookup (`"interface"` vs `"config"`) and reject writes targeting a Tier-0 key, and verify unit tests for both

## 4. `internal/settings` package: M3U source overrides

- [ ] 4.1 Implement origin M3U source listing (parsed from `config.Get().M3U.Sources`) exposed independently of stored rows, and verify a unit test with an empty and a non-empty `config.yml` sources list
- [ ] 4.2 Implement effective M3U source list resolution (stored row wins per `name`, else origin, plus runtime-only names), and verify unit tests for: origin-only, override-replaces-origin, runtime-only-addition, and empty-both-sides
- [ ] 4.3 Implement create/update/delete for a stored M3U source row, including delete-reverts-to-origin and delete-of-runtime-only-removes-entirely, and verify unit tests for each

## 5. API endpoints

- [ ] 5.1 Add `GET /api/v1/settings` (effective values + origin, masking sensitive fields per `design.md`) and `GET /api/v1/settings/origin` (bootstrap + config-only view), following the response shape of the existing `/api/v1/filters` handlers, and verify handler tests
- [ ] 5.2 Add `PUT /api/v1/settings/:key` and `DELETE /api/v1/settings/:key` for setting/clearing one field's override, rejecting unknown or Tier-0 keys with a distinct machine-readable error code, and verify handler tests covering success, unknown key, and Tier-0 key rejection
- [ ] 5.3 Add `GET /api/v1/m3u/sources/origin`, `GET /api/v1/m3u/sources` (effective), `POST /api/v1/m3u/sources`, `PUT /api/v1/m3u/sources/:name`, `DELETE /api/v1/m3u/sources/:name`, mirroring the existing `/api/v1/filters` CRUD handlers in `internal/api/handlers.go`, and verify handler tests for each

## 6. Wire Tier-1 call sites to the effective config

- [ ] 6.1 Switch Radarr/Sonarr handlers in `internal/api/radarr_sonarr_handlers.go` from `config.Get()` to `settings.Effective()`, and verify existing handler tests still pass with zero overrides, plus a new test that an override changes handler behavior
- [ ] 6.2 Switch TMDB/Jellyfin/Notifications config reads (`internal/api/system_status.go`, `internal/notifier`, relevant `cmd/*.go`) to `settings.Effective()`, and verify existing tests pass and a new test shows an override taking effect
- [ ] 6.3 Re-invoke `logger.InitializeLoggersWithFormat` whenever a `logging.*` override is set or cleared, so log level/format changes apply live, and verify a test that setting a `logging.app.level` override changes subsequent log output without a restart
- [ ] 6.4 Flag `tmdb.*` and `downloads.timeout`/`downloads.retry_attempts`/`downloads.min_file_size_mb` overrides as requiring a restart (per the registry's `restart_required` flag) in the settings API response, and verify a handler test asserts the flag is present for exactly this field set

## 7. Frontend: settings drawer structure

- [ ] 7.1 Add "Sources M3U", "Intégrations", "Notifications", and "Avancé" collapsible sections to `SettingsDrawer.tsx` in the order fixed by the `frontend-settings-drawer` delta, collapsed by default, and verify `SettingsDrawer` tests assert the full ten-section order and collapsed states
- [ ] 7.2 Add `en`/`fr` locale entries for the new sections in `src/locales/{en,fr}/settings.json`, and verify the app renders with no missing-translation warnings in either locale

## 8. Frontend: app settings management

- [ ] 8.1 Add an `useAppSettings` hook (fetch/set/clear an override, fetch bootstrap config) and generic field components showing the "Interface"/"Config" origin badge, and verify unit tests for the hook's fetch/set/clear behavior
- [ ] 8.2 Build the sensitive-field input (masked, shows only "set"/"not set", submits only on explicit edit) and verify a component test that leaving it untouched sends no update
- [ ] 8.3 Build the "restart required" indicator driven by the backend's per-field flag, and verify a component test that it appears only for flagged fields
- [ ] 8.4 Build the "Intégrations" (Radarr/Sonarr/TMDB/Jellyfin), "Notifications", and "Avancé" (Downloads + logging) forms using the above components, and verify component tests for each domain's fields
- [ ] 8.5 Build the read-only bootstrap configuration display (database/API port/metrics), and verify a component test that no edit control is rendered

## 9. Frontend: M3U sources management

- [ ] 9.1 Add a `useM3uSources` hook (fetch effective + origin, create/update/delete) and the "Sources M3U" list view distinguishing origin-only, overridden, and runtime-only sources, and verify unit/component tests for each state
- [ ] 9.2 Add create/edit dialog and delete action for M3U sources, reusing the `CreateFilterDialog.tsx` pattern where applicable, and verify component tests for create, edit-an-origin-source (creates an override), and delete-reverts-to-origin

## 10. End-to-end verification

- [ ] 10.1 Add an integration test (or extend `cmd/server_test.go`) that starts the server with only bootstrap config set (no `m3u.sources`, no Radarr/Sonarr/TMDB/Jellyfin/Notifications config anywhere) and verify it boots successfully and responds on `/health`
- [ ] 10.2 Add an integration test that sets an override via the settings API and confirms the corresponding CLI command (e.g. `process` or `download`) picks it up through `settings.Effective()` without a restart
- [ ] 10.3 Run the full test suite (`go test ./...` and the frontend test suite) and verify all tests pass
