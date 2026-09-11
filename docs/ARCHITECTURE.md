
# Architecture

Stalkeer bridges an IPTV/M3U playlist to Radarr/Sonarr: it figures out what those tools are still missing and downloads it straight from the M3U source over HTTP (no torrents involved, despite the repo's home). The diagrams below give the shape of the system before you dive into setup.

## System context

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

## Deployment topology

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

## The "process" pipeline

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

## The "download" pipeline

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

For the underlying database schema and entity relationships, see [docs/DATABASE.md](docs/DATABASE.md#entity-relationships).
