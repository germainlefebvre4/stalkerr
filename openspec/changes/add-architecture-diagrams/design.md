## Context

See `proposal.md` - Why. `README.md` currently jumps straight from the feature list to install/config instructions with no picture of the system: what's optional (Radarr/Sonarr/Jellyfin are all `enabled: false` by default per `config.yml.example`), what's required (PostgreSQL, TMDB), and how the two core CLI pipelines (`process`, `download`) actually move data. `docs/DATABASE.md` already has a maintained entity-relationship tree (lines 295-320) - the design below deliberately does not duplicate it.

## Goals / Non-Goals

**Goals:**
- Give a first-time reader a system-context view (every external system Stalkeer talks to, and which are optional) before they read the setup steps.
- Show the docker-compose service topology (what `server`/`process`/`download`/`m3u-download`/`ofelia`/`frontend` each do and how they relate) so `docker-compose.yml` doesn't have to be reverse-engineered.
- Show the two pipelines (`process`, `download`) at a level that matches the CLI commands, not internal function calls.
- Produce finished Mermaid source, ready to paste into `README.md` as-is.

**Non-Goals:**
- Re-documenting the database schema/ER relationships (already in `docs/DATABASE.md`) - link to it instead.
- Diagramming the Helm chart / Kubernetes CronJob topology (already covered in prose in `charts/stalkerr/README.md`).
- Restructuring any existing README section other than inserting the new one.

## Decisions

**Mermaid, not ASCII.** Confirmed with the user: this targets GitHub rendering (real boxes/arrows) over terminal-portability. Trade-off accepted below.

**Placement:** new `## Architecture` section immediately after `## Features` and before `## Quick Start` - a reader sees the shape of the system before being handed install steps.

**Four diagrams, one concern each**, rather than one combined diagram (a single graph mixing ~9 external systems with two internal pipelines becomes unreadable):

1. System context
2. Deployment topology (docker-compose)
3. `process` pipeline
4. `download` pipeline

**ER diagram excluded**, replaced with a link to `docs/DATABASE.md#entity-relationships`, to avoid a second hand-maintained copy drifting out of sync with the first.

**Alternatives considered:**
- Single "big picture" diagram - rejected, unreadable at this node count.
- Committed PNG/SVG images instead of Mermaid source - rejected, not diffable/greppable in a text review, and Mermaid renders natively on GitHub without an export step.

### Diagram 1 - System context

```mermaid
flowchart LR
    subgraph Stalkeer["Stalkeer (CLI + REST API + Scheduler)"]
        direction TB
        API["REST API :8080"]
        CLI["CLI commands\nprocess / download / m3u-download"]
    end

    IPTV["IPTV provider\n(M3U source)"]
    TMDB[("TMDB API")]
    TVDB[("TVDB API")]
    Radarr["Radarr\n(optional)"]
    Sonarr["Sonarr\n(optional)"]
    Jellyfin["Jellyfin\n(optional)"]
    DB[("PostgreSQL")]
    Prom["Prometheus\n(optional)"]
    UI["Browser /\nReact Dashboard"]

    IPTV -- "M3U playlist" --> CLI
    CLI -- "direct HTTP download" --> IPTV
    CLI -- "metadata enrich" --> TMDB
    CLI -- "season/episode IDs" --> TVDB
    CLI -- "missing movies, root folders" --> Radarr
    CLI -- "missing episodes, root folders" --> Sonarr
    CLI -- "library rescan notify" --> Jellyfin
    CLI --- DB
    API --- DB
    Prom -. "scrape /metrics :8081" .-> API
    UI -- "REST /api/v1" --> API
```

### Diagram 2 - Deployment topology (docker-compose)

```mermaid
flowchart TD
    subgraph compose["docker-compose (profiles: stalkerr / development / cron)"]
        FE["frontend :5173\n(React dashboard)"]
        SRV["server :8080 / :8081\n(stalkeer server)"]
        PROC["process\n(one-shot job)"]
        DL["download\n(one-shot job)"]
        M3UDL["m3u-download\n(one-shot job)"]
        PG[("postgres :5432")]
        OFELIA["ofelia\n(cron sidecar, profile: cron)"]
        RADARR_DEV["radarr :7878\n(dev-only stand-in)"]
        SONARR_DEV["sonarr :8989\n(dev-only stand-in)"]
    end

    RADARR_EXT["your Radarr\n(RADARR_URL)"]
    SONARR_EXT["your Sonarr\n(SONARR_URL)"]

    FE -- "proxy /api/v1" --> SRV
    SRV --> PG
    PROC --> PG
    DL --> PG
    M3UDL --> PG
    OFELIA -- "docker.sock: triggers on schedule" --> PROC
    OFELIA --> DL
    OFELIA --> M3UDL
    DL -. "production" .-> RADARR_EXT
    DL -. "production" .-> SONARR_EXT
    DL -. "local dev only" .-> RADARR_DEV
    DL -. "local dev only" .-> SONARR_DEV
```

### Diagram 3 - `process` pipeline

```mermaid
flowchart TD
    A["M3U file / URL"] --> B["parser\n(parse lines)"]
    B --> C["filter\n(include/exclude patterns)"]
    C --> D["classifier\n(movie / tvshow / channel / uncategorized)"]
    D --> E{"duplicate?\nsource_name + line_hash"}
    E -- yes --> Z["skip"]
    E -- no --> F["TMDB enrich\ntitle, year, genres, poster"]
    F --> G["batch save"]
    G --> H[("processed_lines")]
    G --> I[("movies / tvshows /\nchannels / uncategorized")]
    G --> J[("processing_logs")]
```

### Diagram 4 - `download` pipeline

```mermaid
flowchart TD
    R["Radarr: missing monitored movies"] --> RC["reconcile root-folder paths\nagainst DB"]
    S["Sonarr: missing monitored episodes"] --> RC
    RC --> BS["scheduler.BuildStreams"]

    BS --> T1["Tier 1: never-downloaded content,\nmatched to processed_lines by TMDB/TVDB id\n(+season/episode), ranked by\nresolution / language / French variant"]
    BS --> T2["Tier 2: already-downloaded but a\nbetter-quality candidate exists (upgrade)"]
    BS --> RESUME["+ merge resumable / incomplete\ndownload_info rows"]

    T1 --> POOL["worker pool (N parallel)\nScheduler.ClaimNext()"]
    T2 --> POOL
    RESUME --> POOL

    POOL --> HTTP["HTTP GET direct link\n(Range / resume support)"]
    HTTP --> MOVE["move to Radarr/Sonarr-style destination\n(quality-tagged filename)"]
    MOVE --> UPD["update download_info +\nprocessed_lines.state"]
    UPD --> NOTIFY["notifyJellyfin(changed paths)\noptional, rescans only touched folders"]
```

## Risks / Trade-offs

- [Mermaid doesn't render in plain-text viewers - terminal `cat`, some git GUIs, non-Mermaid Markdown renderers] -> Mitigation: node/edge labels are plain English, so the raw fenced code block still reads as a structured outline even unrendered; GitHub, GitLab, and VS Code's built-in preview all render Mermaid natively.
- [Diagrams drift from code as the pipeline evolves] -> Mitigation: each diagram is deliberately pinned at command/service granularity (not function-level), so ordinary feature work shouldn't invalidate them; only a new external integration, a new docker-compose service, or a new pipeline stage should require an update.

## Migration Plan

Single additive edit to `README.md` - insert the new section, no other content changes. Rollback is a plain revert of that one commit.
