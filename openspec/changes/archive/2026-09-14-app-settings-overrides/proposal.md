## Why

Stalkeer can only be configured today via `config.yml`, environment variables, or Docker/Helm-injected env — any change requires editing a file or redeploying. The app already proves a better model for one narrow case: runtime filter overrides (`filter-override-policy`) let an operator override `config.yml`-defined filter patterns from the UI, with the origin (file vs runtime) always visible. This change generalizes that same override-with-origin pattern to the rest of the application's configuration (Radarr, Sonarr, TMDB, Jellyfin, notifications, downloads tuning, logging levels, and M3U sources), so the app can be fully configured from its own interface, and so it boots successfully even when none of that configuration exists in `config.yml` or environment variables.

## What Changes

- New generic per-field settings-override mechanism: for every applicative config key across Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads, and logging levels/format, a value set from the UI overrides the `config.yml`/env value for that field. Effective value = UI override if present, else the existing file/env-resolved value (including built-in defaults), never merged partially within a single override.
- New read-only exposure, with origin, of the bootstrap-tier config that stays file/env-only because the app needs it before it can reach its own database or bind its ports: `database.*` (host/port/user/password/dbname/sslmode), `api.port`, `metrics.port`, `metrics.enabled`. These are visible in the UI but not editable from it.
- New dedicated DB-backed CRUD for M3U sources (`m3u.sources`), keyed by each source's `name`, mirroring the filter-override mechanism: a stored override for a given `name` replaces that source's entire block (never merges field-by-field); a `name` with no override keeps using its file-defined entry; a `name` that exists only as a stored override is a source created entirely from the UI.
- **BREAKING**: `m3u.sources` is no longer required to be a non-empty list for the app to load its configuration and boot. The effective source list is the union of file-defined and UI-created/overridden sources and may be empty until at least one exists (in the file or the UI).
- The existing Settings drawer is replaced by a dedicated Configuration page, reachable from the same header icon, so the growing set of backend-settable configuration has room to be presented. Every section the drawer used to hold (Apparence, Langue, Filtres, Préférences, Système, À propos) moves to this page unchanged, alongside new sections to view and edit the settings above, each field showing a two-state origin badge ("Interface" vs "Config" — file and env are not distinguished from each other) and, for the handful of settings still baked into long-lived server state at process start (TMDB client, downloader tuning, metrics listener), a "restart required" indicator rather than a false "applied immediately" claim.

## Capabilities

### New Capabilities
- `app-settings`: generic backend override store and per-field origin resolution ("Interface" vs "Config") for Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads, and logging config, plus read-only exposure of the bootstrap-tier config (`database.*`, `api.port`, `metrics.*`) with its origin.
- `m3u-source-overrides`: DB-backed CRUD for M3U sources, keyed by `name`, with whole-block override and origin exposure — the same pattern `filter-override-policy` already established, applied to sources instead of filters.
- `frontend-app-settings-management`: Configuration-page UI to view and edit `app-settings`-backed fields and manage M3U sources, including origin badges and restart-required indicators.
- `frontend-configuration-page`: the dedicated page (replacing the Settings drawer) reachable from the header settings icon, carrying every section the drawer used to hold plus the new ones this change adds (Sources M3U, Intégrations, Notifications, Avancé).

### Modified Capabilities
- `m3u-multi-source`: relax "Configurable list of M3U sources" — `m3u.sources` is no longer required to be non-empty for configuration to load successfully; the effective source list now comes from `m3u-source-overrides` (file-defined entries plus DB overrides/additions) and may be empty.
- `frontend-settings-drawer`: every requirement removed, superseded by `frontend-configuration-page` — the drawer no longer exists.
- `frontend-filters-management`: "Filters List View" now references the Configuration page instead of the Settings drawer as its container; the filter management behavior itself is unchanged.

## Impact

- `internal/config/config.go`: `validate()` no longer fails when `m3u.sources` is empty; bootstrap validation (`database.user`, `database.dbname`) is unchanged.
- New DB models/migrations: an app-settings override store (shape detailed in `design.md`) and `m3u_source_configs`.
- New API endpoints mirroring the existing `/filters` pattern: origin/effective/CRUD for app settings, and for M3U sources.
- `internal/config` cannot import `internal/database` (the reverse import already exists), so resolving "effective config" (file/env + DB overrides) needs a new merge point usable consistently by both the API server and the CLI one-shot commands (`cmd/process.go`, `cmd/download.go`, `cmd/m3u.go`, etc.) — detailed in `design.md`.
- Frontend: `SettingsDrawer.tsx` is retired; a new Configuration page/route (and its navigation from the header icon) replaces it, hosting every existing section plus new components/hooks per settings domain; locale updates in `src/locales/{en,fr}`.
- Secrets (Radarr/Sonarr/TMDB API keys, ntfy auth token) become storable in the application database when overridden from the UI, in addition to the existing K8s Secret/env path — masking and access handling detailed in `design.md`.
- Out of scope: the Helm chart's own config delivery and validation (`configuration-management`, ConfigMap/Secret, `values.schema.json` requiring non-empty `m3u.sources`) is unchanged by this proposal; loosening chart-level validation is a separate, later change.
