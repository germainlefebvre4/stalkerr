<div align="center">
  <img src="./media/logo/stalkerr-black.svg" alt="Stalkerr Logo" style="width:250px;"/>

# Stalkeer

**Turn m3u playlists into source for Radarr and Sonarr**

Parse M3U playlists and download missing media items from Radarr and Sonarr via direct links.

</div>

## Features

- 📺 Parse M3U playlist files for movies and TV shows
- 📥 **M3U Download** - Download and archive M3U playlists from remote URLs with automatic rotation
- 🎬 **TMDB Integration** - Enrich media metadata with title, year, genres, posters, and more
- 🗄️ Store media information in PostgreSQL database
- 🔍 Filter playlist items based on configurable patterns
- 🌍 Multi-language support for TMDB metadata (English, French, Spanish, etc.)
- 🎯 Identify missing items from Radarr and Sonarr
- ⬇️ Download missing items via direct links
- 🚀 REST API for querying and managing media items
- 📊 Processing logs and statistics
- 🖥️ **Web Dashboard (IHM)** - Modern light-themed dashboard built with React 19 and Radix UI to explore playlists, monitor download progress, view processing logs, and execute folder-level media moves. Fully responsive on mobile.
- 🔗 **Radarr/Sonarr path reconciliation** - Automatic reconciliation of download destinations against Radarr/Sonarr root folders and monitored items
- 🏷️ Automatic quality/language tagging of downloaded filenames
- ⚠️ Dedicated downloads error tab with cancel, resync-path, and rename actions
- 🗂️ Grouped playlist view with per-group filters and expandable latest-run details

## Architecture

Stalkeer bridges an IPTV/M3U playlist to Radarr/Sonarr: it figures out what those tools are still missing and downloads it straight from the M3U source over HTTP (no torrents involved, despite the repo's home). The diagrams below give the shape of the system before you dive into setup.

![Stalkerr Overview](./docs/images/stalkerr-overview.png)

For more information about architecture implementation, you can read the page [ARCHITECTURE.md](./docs/ARCHITECTURE.md).

## Quick Start

### Option 1: Docker (Recommended)

The fastest way to get started:

```bash
# Clone and start
git clone https://github.com/glefebvre/stalkeer.git
cd stalkeer
cp .env.example .env
docker-compose up -d

# Verify
curl http://localhost:8080/health
```

For detailed Docker deployment instructions, see:
- [Docker Quick Start](DOCKER-QUICKSTART.md) - Get running in 5 minutes
- [Docker Deployment Guide](docs/DOCKER-DEPLOYMENT.md) - Complete deployment guide

**Scheduled Execution**: enable the `cron` profile (`docker-compose --profile stalkerr --profile cron up -d`) for an opt-in Ofelia sidecar that runs `m3u-download`/`process`/`download` on the same default cadence as the Helm chart's `CronJob` resources below. See [Docker Quick Start](DOCKER-QUICKSTART.md#next-steps) for the full recipe, including a lower-privilege host-crontab alternative.

### Option 2: Kubernetes (Helm)

A Helm chart is available under [charts/stalkerr](charts/stalkerr) for deploying Stalkeer to a Kubernetes cluster, including a `CronJob` for scheduled processing. See [charts/stalkerr/README.md](charts/stalkerr/README.md) for values and installation instructions.

### Option 3: Build from Source

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 12 or higher
- M3U playlist file

### Installation

```bash
# Clone the repository
git clone https://github.com/glefebvre/stalkeer.git
cd stalkeer

# Install dependencies
go mod download

# Build the application
make build
```

### Configuration

1. Copy the example configuration:
```bash
cp config.yml.example config.yml
```

2. Edit `config.yml` with your settings:
```yaml
database:
  host: localhost
  port: 5432
  user: stalkeer
  password: your_password
  dbname: stalkeer

m3u:
  update_interval: 3600
  sources:
    - name: default
      file_path: /path/to/playlist.m3u

tmdb:
  enabled: true
  api_key: your_tmdb_api_key  # Get from https://www.themoviedb.org/settings/api
  language: en-US  # Language for metadata (en-US, fr-FR, es-ES, etc.)

api:
  port: 8080
```

Or use environment variables (`m3u.sources` is config-file only - see [docs/M3U-DOWNLOAD.md](docs/M3U-DOWNLOAD.md)):
```bash
export STALKEER_DATABASE_USER=stalkeer
export STALKEER_DATABASE_PASSWORD=your_password
export STALKEER_DATABASE_DBNAME=stalkeer
```

### Running

```bash
# Process an M3U playlist file and store to database
./bin/stalkeer process /path/to/playlist.m3u

# Process with TMDB enrichment in French
./bin/stalkeer process /path/to/playlist.m3u --tmdb-language fr-FR

# Process without TMDB enrichment (faster)
./bin/stalkeer process /path/to/playlist.m3u --skip-tmdb

# Process with limit
./bin/stalkeer process /path/to/playlist.m3u --limit 100

# Force re-processing of existing entries
./bin/stalkeer process /path/to/playlist.m3u --force

# Dry-run analysis without database changes
./bin/stalkeer dryrun /path/to/playlist.m3u --limit 100

# Start the REST API server
./bin/stalkeer server --port 8080

# Run database migrations
./bin/stalkeer migrate

# Resume incomplete or failed downloads
./bin/stalkeer resume-downloads

# Resume downloads with options
./bin/stalkeer resume-downloads --limit 10 --parallel 5 --verbose

# Preview what would be resumed (dry-run)
./bin/stalkeer resume-downloads --dry-run

# Download missing movies/episodes from Radarr and Sonarr
./bin/stalkeer download --limit 20

# Preview matches without downloading
./bin/stalkeer download --dry-run

# Check version
./bin/stalkeer version

# Get help
./bin/stalkeer --help
./bin/stalkeer process --help
```

### CLI Commands

#### m3u-download

Download M3U playlist from a remote URL and save it to the configured file path:

```bash
stalkeer m3u-download [flags]

Flags:
      --url string     M3U playlist URL (overrides config)
      --no-archive     skip creating archive copy
```

The m3u-download command:
- Downloads M3U playlist from the configured URL or --url flag
- Validates M3U format before saving
- Saves to each configured source's effective file path atomically
- Creates a timestamped archive copy (unless --no-archive)
- Automatically rotates old archives based on retention settings

Example usage:
```bash
# Download using configured URL
stalkeer m3u-download

# Download from specific URL
stalkeer m3u-download --url https://example.com/playlist.m3u

# Download without creating archive
stalkeer m3u-download --no-archive
```

#### m3u-list-archives

List all archived M3U playlist files with timestamps and sizes:

```bash
stalkeer m3u-list-archives
```

Example output:
```
Archived M3U files (./m3u_playlist):

Filename                                 Size         Modified
--------------------------------------------------------------------------------
playlist_20260203_230154.054749.m3u      40.96 MB     2026-02-03 23:01:54
playlist_20260203_225902.373353.m3u      40.96 MB     2026-02-03 22:59:02
playlist_20260202_120000.000000.m3u      40.95 MB     2026-02-02 12:00:00

Total: 3 archived files
```

#### m3u-cleanup-archives

Manually trigger rotation of M3U archive files:

```bash
stalkeer m3u-cleanup-archives [flags]

Flags:
      --retention int   number of archives to keep (default: use config value)
```

Example usage:
```bash
# Clean up using configured retention count
stalkeer m3u-cleanup-archives

# Keep only the 3 most recent archives
stalkeer m3u-cleanup-archives --retention 3
```

#### process

Process an M3U playlist file, classify content, enrich with TMDB metadata, and store entries in the database:

```bash
stalkeer process [m3u file] [flags]

Flags:
      --force              re-process existing entries
      --limit int          maximum number of items to process (0 = no limit)
      --batch-size int     batch size for database inserts (default 100)
      --progress int       show progress every N entries (default 1000)
      --skip-tmdb          skip TMDB metadata enrichment
      --tmdb-language      TMDB API language (e.g., 'en-US', 'fr-FR')
```

Example output:
```
=== Processing Complete ===
Total lines in file:  1000
Successfully processed: 985
Duplicates skipped:   5
Filtered out:         8
Errors:               2

Content breakdown:
  Movies:        742
  TV Shows:      231
  Channels:      0
  Uncategorized: 12

TMDB Enrichment:
  Matched:       891
  Not found:     76
  Errors:        6
  Match rate:    92.1%

Processing time: 1.2s
```

#### resume-downloads

Resume incomplete or failed downloads that were interrupted:

```bash
stalkeer resume-downloads [flags]

Flags:
      --dry-run             preview downloads without executing
      --limit int           maximum number of downloads to process (0 = no limit)
      --parallel int        number of concurrent downloads (default 3)
      --max-retries int     maximum retry attempts (default 5)
      --clean-stale-locks   clean up stale download locks before resuming (default true)
  -v, --verbose             verbose output
      --service string      filter by service type: all, radarr, sonarr (default "all")
```

The resume-downloads command identifies and resumes downloads that:
- Were interrupted by application shutdown or crashes
- Failed due to temporary network issues
- Are in pending, downloading, paused, or failed states
- Haven't exceeded the maximum retry limit

Example usage:
```bash
# Resume all incomplete downloads
stalkeer resume-downloads

# Preview what would be resumed
stalkeer resume-downloads --dry-run --verbose

# Resume up to 10 downloads with 5 concurrent workers
stalkeer resume-downloads --limit 10 --parallel 5

# Resume only failed Radarr downloads
stalkeer resume-downloads --service radarr
```

#### download

Unified command that downloads missing movies (Radarr) and TV episodes (Sonarr) by matching against the M3U playlist. This command replaces the removed `radarr` and `sonarr` commands. Resuming incomplete downloads and tier-2 upgrades are now handled automatically by the scheduler (see `downloads.force_tier_probability` in the config) instead of dedicated `--resume`/`--force` flags:

```bash
stalkeer download [flags]

Flags:
      --dry-run      preview planned streams without downloading
      --limit int    maximum number of work units (movies/series) to consider (0 = no limit)
      --parallel int number of concurrent worker streams
  -v, --verbose      verbose output
```

#### cleanup

Clean up orphaned temp download files:

```bash
stalkeer cleanup [flags]

Flags:
      --dry-run             preview cleanup without deleting files
      --retention-hours int delete temp files older than this many hours (default 24)
```

#### config

Validate and display the current configuration:

```bash
stalkeer config [flags]

Flags:
      --show-secrets   reveal password fields
```

#### db-prune

Prune expired M3U stream URLs and orphaned metadata:

```bash
stalkeer db-prune [flags]

Flags:
      --dry-run   simulate pruning and display metrics without deleting
      --hard      force delete downloaded and downloading stream records
```

#### enrich-tvdb

Backfill missing TVDB IDs on Movie and TVShow records:

```bash
stalkeer enrich-tvdb [flags]

Flags:
      --dry-run     preview records that would be updated without writing to database
      --limit int   maximum number of records to process (0 = no limit)
  -v, --verbose     verbose output
```

#### reset

Surgically reset a specific movie or TV show stream state by ID:

```bash
stalkeer reset movie --id 42
stalkeer reset tvshow --id 17
```

#### dryrun

Analyze M3U playlist file without making database changes:

```bash
stalkeer dryrun [m3u file] [flags]

Flags:
      --limit int   maximum number of items to analyze (default 100)
```

#### server

Start the REST API server:

```bash
stalkeer server [flags]

Flags:
  -p, --port int       port to run the server on (default 8080)
  -a, --address string address to bind the server to (default "0.0.0.0")
```

#### migrate

Run database migrations:

```bash
stalkeer migrate
```

### Using Docker Compose

```bash
# Start PostgreSQL
docker-compose up -d postgres

# Run the application
./bin/stalkeer process
```

## Development

### Project Structure

```
stalkeer/
├── cmd/                    # CLI commands (process, download, resume-downloads,
│                           # cleanup, config, db-prune, enrich-tvdb, reset, server, ...)
├── frontend/               # React 19 + Radix UI Frontend Web Dashboard
├── internal/               # Private application code
│   ├── api/               # REST API handlers
│   ├── classifier/        # Content classification (movie/TV/channel)
│   ├── config/            # Configuration management
│   ├── database/          # Database connection and migrations
│   ├── downloader/        # Download engine (resume, tiers, path reconciliation)
│   ├── external/          # Radarr/Sonarr/TMDB/TVDB API clients
│   ├── filter/            # Include/exclude pattern filtering
│   ├── logger/            # Modular application/database logging
│   ├── matcher/           # Playlist-to-Radarr/Sonarr matching
│   ├── models/            # Data models
│   ├── parser/            # M3U parser
│   ├── processor/         # M3U processing pipeline
│   ├── scheduler/         # Download scheduler (tier selection, cron)
│   └── testutil/          # Test helpers and fixtures
├── charts/stalkerr/        # Helm chart for Kubernetes deployment
├── docs/                  # Documentation
├── .github/               # GitHub workflows and templates
├── config.yml.example     # Example configuration
├── docker-compose.yml     # Docker services
├── Makefile              # Build automation
└── README.md             # This file
```

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Generate coverage report
make coverage

# Format code
make fmt

# Run linters
make lint

# Clean build artifacts
make clean
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

See [docs/TESTING.md](docs/TESTING.md) for detailed testing documentation.

### Development Setup

To quickly run and develop both the Go backend API and the React 19 Frontend Web Dashboard locally, use the following commands:

#### 1. Prerequisites

Ensure you have the following installed on your machine:
- **Go** 1.25 or higher
- **Node.js** 22 or higher (with `npm`)
- **Docker** and **Docker Compose** (for the database and profile dependencies)

#### 2. Local Database & Dependencies
Start the PostgreSQL container and any other necessary development profile dependencies:
```bash
make docker-up
```

#### 3. Frontend Installation
Install the React dashboard dependencies in the `/frontend` directory:
```bash
make front-install
```

#### 4. Run Both Servers Concurrently
Start the local development server unifing both Go API on port `8080` and Vite Frontend on port `5173`:
```bash
make dev
```
*   **Web Dashboard (IHM)**: Available at [http://localhost:5173](http://localhost:5173)
*   **API v1 Server**: Available at [http://localhost:8080](http://localhost:8080)

*Note:* Vite is preconfigured to automatically proxy all REST API requests on `/api/v1/*` directly to port `8080` to avoid CORS issues.

For advanced settings, see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed development setup instructions.

## API Documentation

The REST API provides endpoints for managing processed M3U items, movies, TV shows, filters, and downloads.

### Health Check

```bash
GET /health
```

### Items (processed M3U entries)

```bash
GET  /api/v1/items                    # List all processed items
GET  /api/v1/items/grouped            # List items grouped by title
GET  /api/v1/items/:id                # Get item by ID
PUT  /api/v1/items/:id                # Update item
POST /api/v1/items/search             # Search items
POST /api/v1/items/:id/override       # Manually override an item's match
POST /api/v1/items/:id/force-download # Force (re-)download an item
```

### Movies

```bash
GET  /api/v1/movies             # List all movies
GET  /api/v1/movies/:id         # Get movie by ID
POST /api/v1/movies/:id/move    # Move a movie's folder
POST /api/v1/movies/:id/reset   # Reset a movie's stream state
```

### TV Shows

```bash
GET  /api/v1/tvshows            # List all TV shows
GET  /api/v1/tvshows/:id        # Get TV show by ID
POST /api/v1/tvshows/:id/move   # Move a TV show's folder
POST /api/v1/tvshows/:id/reset  # Reset a TV show's stream state
```

### Radarr / Sonarr

```bash
GET /api/v1/radarr/movies                  # List Radarr-monitored movies
GET /api/v1/radarr/movies/:id/matches       # Get playlist matches for a Radarr movie
GET /api/v1/sonarr/series                   # List Sonarr-monitored series
GET /api/v1/sonarr/series/:id/episodes      # Get episodes for a Sonarr series
GET /api/v1/radarr-sonarr/stats             # Radarr/Sonarr reconciliation stats
```

### Filters

```bash
GET    /api/v1/filters            # List filters
POST   /api/v1/filters            # Create a filter
PATCH  /api/v1/filters/:id        # Update a filter
DELETE /api/v1/filters/:id        # Delete a filter
DELETE /api/v1/filters/runtime    # Clear runtime-only filters
```

### Downloads

```bash
GET  /api/v1/downloads                     # List downloads (enriched)
GET  /api/v1/downloads/simple              # List downloads (simple)
POST /api/v1/downloads/:id/rename          # Rename a download
POST /api/v1/downloads/:id/resync-path     # Resync a download's destination path
POST /api/v1/downloads/:id/cancel          # Cancel a download
```

### Dry Run, Stats & Logs

```bash
POST /api/v1/dryrun            # Execute a dry-run analysis
GET  /api/v1/stats             # Get processing statistics
GET  /api/v1/processing-logs   # List processing logs
GET  /api/v1/config/paths      # Get configured download paths
```

### TMDB Proxy

```bash
GET /api/v1/tmdb/search   # Proxy a TMDB search query
```

## Configuration

### Database Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `database.host` | string | `localhost` | PostgreSQL host |
| `database.port` | int | `5432` | PostgreSQL port |
| `database.user` | string | - | Database user (required) |
| `database.password` | string | - | Database password |
| `database.dbname` | string | - | Database name (required) |
| `database.sslmode` | string | `disable` | SSL mode |

### M3U Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `m3u.sources` | list | - | List of M3U sources (required, non-empty); see [docs/M3U-DOWNLOAD.md](docs/M3U-DOWNLOAD.md) |
| `m3u.update_interval` | int | `3600` | Update interval in seconds |

### Logging Configuration

Stalkeer supports modular logging with independent control for application and database logging:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `logging.format` | string | `json` | Log output format (`json` or `text`) |
| `logging.app.level` | string | `info` | Application log level (`debug`, `info`, `warn`, `error`) |
| `logging.database.level` | string | `info` | Database log level (`debug`, `info`, `warn`, `error`) |
| `logging.level` | string | `info` | Legacy fallback log level (deprecated) |

**Formats:**
- `json`: Structured JSON logs (recommended for production)
- `text`: Human-readable text logs (useful for development)

**Example:**
```yaml
logging:
  format: json  # or text
  app:
    level: info
  database:
    level: warn  # Less verbose for DB queries
```

For more details, see [docs/LOGGING.md](docs/LOGGING.md).

### API Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `api.port` | int | `8080` | API server port |

### Filter Configuration

File-based filters applied to `group_title` and `tvg_name` fields (can be overridden by runtime filters via the `/api/v1/filters` endpoints):

| Field | Type | Description |
|-------|------|-------------|
| `filter.group_title.include_patterns` | []string | Regex patterns a group title must match |
| `filter.group_title.exclude_patterns` | []string | Regex patterns that exclude a group title |
| `filter.tvg_name.include_patterns` | []string | Regex patterns a channel/stream name must match |
| `filter.tvg_name.exclude_patterns` | []string | Regex patterns that exclude a channel/stream name |

### TMDB Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tmdb.enabled` | bool | `true` | Enable TMDB metadata enrichment |
| `tmdb.api_key` | string | - | TMDB API key |
| `tmdb.language` | string | `en-US` | Language for TMDB metadata |
| `tmdb.requests_per_second` | float | `4.0` | Max TMDB API requests per second (0 disables rate limiting) |

### Radarr / Sonarr Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `radarr.enabled` | bool | `false` | Enable Radarr integration |
| `radarr.url` | string | - | Radarr base URL |
| `radarr.api_key` | string | - | Radarr API key |
| `radarr.sync_interval` | int | `3600` | Sync interval in seconds |
| `radarr.quality_profile_id` | int | `1` | Radarr quality profile ID |
| `sonarr.enabled` | bool | `false` | Enable Sonarr integration |
| `sonarr.url` | string | - | Sonarr base URL |
| `sonarr.api_key` | string | - | Sonarr API key |
| `sonarr.sync_interval` | int | `3600` | Sync interval in seconds |
| `sonarr.quality_profile_id` | int | `1` | Sonarr quality profile ID |

### Downloads Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `downloads.movies_path` | string | `./data/downloads/movies` | Fallback path when Radarr's `movie.Path` is empty |
| `downloads.tvshows_path` | string | `./data/downloads/tvshows` | Fallback path when Sonarr's `series.Path` is empty |
| `downloads.temp_dir` | string | OS temp dir | Directory for in-progress downloads |
| `downloads.max_parallel` | int | `3` | Number of concurrent downloads |
| `downloads.timeout` | int | `600` | Download timeout in seconds |
| `downloads.retry_attempts` | int | `3` | HTTP sub-retry attempts within a single download attempt |
| `downloads.resume_enabled` | bool | `true` | Enable resumable downloads |
| `downloads.progress_interval_mb` | int | `10` | Persist progress every N megabytes |
| `downloads.progress_interval_seconds` | int | `30` | Persist progress every N seconds |
| `downloads.lock_timeout_minutes` | int | `5` | Consider a download lock stale after this many minutes |
| `downloads.max_retry_attempts` | int | `5` | Max full download attempts before an occurrence is permanently skipped |
| `downloads.force_tier_probability` | float | `0.1` | Chance (0-1) that the `download` command scheduler draws an already-downloaded/upgrade candidate instead of new content |

## Environment Variables

Most configuration options can be overridden with environment variables using the `STALKEER_` prefix:

- `STALKEER_DATABASE_HOST`
- `STALKEER_DATABASE_PORT`
- `STALKEER_DATABASE_USER`
- `STALKEER_DATABASE_PASSWORD`
- `STALKEER_DATABASE_DBNAME`
- `STALKEER_LOGGING_LEVEL`
- `STALKEER_API_PORT`

`m3u.sources` is the exception: it's a structured list and is config-file only, with no environment variable equivalent (see [docs/M3U-DOWNLOAD.md](docs/M3U-DOWNLOAD.md)).

Or use a PostgreSQL connection string:
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/stalkeer"
```

## Contributing

Contributions are welcome! Please read our contributing guidelines (coming soon) before submitting pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Commit / PR title format

This repository only allows squash merges, so a pull request's title becomes its commit message on `main`. That title **must** follow the [Conventional Commits](https://www.conventionalcommits.org/) format (see [`.github/skills/git-commit/SKILL.md`](.github/skills/git-commit/SKILL.md) for the full type list and examples) - CI ([`commitlint.yml`](.github/workflows/commitlint.yml)) enforces this on every pull request, and the format drives automatic semantic-version bumps and changelog generation via [release-please](https://github.com/googleapis/release-please).

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [GORM](https://gorm.io/) - ORM library for Go
- [Gin](https://gin-gonic.com/) - HTTP web framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/glefebvre/stalkeer/issues)
- 💬 [Discussions](https://github.com/glefebvre/stalkeer/discussions)
