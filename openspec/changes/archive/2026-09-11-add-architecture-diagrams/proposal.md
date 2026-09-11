## Why

New contributors and operators can't tell what Stalkeer *is* or what it talks to without reading Go source across a dozen packages. `README.md` documents every CLI flag and API route in detail, but has zero diagrams: no system-context view (what external services it connects to), no deployment topology (what the docker-compose services do and how they relate), and no view of the two core pipelines (`process` and `download`). This makes onboarding unnecessarily slow, and the repo name (`TorrentTracker/stalkeer`) is actively misleading about what the tool does (it downloads via direct HTTP links from an IPTV/M3U source, not torrents).

## What Changes

- Add a new `## Architecture` section to `README.md`, placed after `## Features` and before `## Quick Start`.
- Add four Mermaid diagrams to that section:
  1. **System context** — Stalkeer and every external system it connects to (M3U/IPTV source, TMDB, TVDB, Radarr, Sonarr, Jellyfin, PostgreSQL, Prometheus, the browser/React dashboard), each optional integration marked as such.
  2. **Deployment topology** — the docker-compose services (`server`, `process`, `download`, `m3u-download`, `postgres`, `ofelia`, `frontend`, dev-only `radarr`/`sonarr`) and how they relate, including the `cron` profile.
  3. **`process` pipeline** — playlist ingestion flow: parse -> filter -> classify -> dedupe -> TMDB enrich -> batch save.
  4. **`download` pipeline** — Radarr/Sonarr missing-item lookup -> Tier1/Tier2 stream building -> scheduler worker pool -> resumable HTTP download -> Jellyfin rescan notify.
- Add one link from the new section to `docs/DATABASE.md#entity-relationships` for the data-model diagram, instead of duplicating it.
- No other files change; no existing prose is removed.

## Capabilities

No spec-level behavior changes — this is a documentation-only addition (`skip_specs: true`, no new/modified capabilities).

## Impact

- **Affected files**: `README.md` only.
- **Systems documented (not changed)**: PostgreSQL, TMDB, TVDB, Radarr, Sonarr, Jellyfin, Prometheus, the M3U/IPTV source, the React frontend, docker-compose services, Ofelia/Helm CronJob scheduling.
- **Risk**: none — purely additive documentation; no code, config, or schema is touched.
