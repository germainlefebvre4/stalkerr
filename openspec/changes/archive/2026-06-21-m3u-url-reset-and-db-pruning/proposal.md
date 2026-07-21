## Why

Over time, as M3U playlists are downloaded and processed daily/weekly, stream URLs are rotated by providers (expiring old tokens and introducing new ones). Because Stalkeer hashes each stream entry based on `tvg-name` + `URL`, every URL change generates a new `ProcessedLine` record in the database. This leads to:
1. Significant database bloat (thousands of expired `processed_lines` pointing to the same media).
2. The inability to easily re-download/refresh specific media items that have expired or failed without using a brute-force `--force` on the entire dataset.

There is a need for a database maintenance mechanism to prune obsolete stream links and a surgical reset capability (CLI & REST API) to force URL refreshing and re-downloading of specific movies or TV shows.

## What Changes

- **New CLI Command `db-prune`**: Clean up `processed_lines` that are no longer present in the active M3U file, with options for hard/soft pruning (preserving/deleting downloaded histories), followed by cleaning up orphaned metadata in `movies` and `tvshows`.
- **New CLI Command `reset`**: Reset a specific movie or TV show by removing its associated processed lines, which marks it as pending for the next playlist processing run.
- **New API Endpoints**: Expose movie and TV show reset functionalities via HTTP POST endpoints (`/api/v1/movies/:id/reset` and `/api/v1/tvshows/:id/reset`) to enable integration with frontend interfaces.

## Capabilities

### New Capabilities

- `db-prune`: Provides the automated capability to compare active M3U hashes against the database and prune expired `processed_lines`, `movies`, and `tvshows` to maintain database hygiene.
- `media-reset`: Provides CLI and REST API endpoints to surgically reset individual movies or TV shows, deleting their associated stream rows to force a refresh on the next import.

### Modified Capabilities

- `cmd-entrypoints`: The CLI interface is extended to support the new `db-prune` and `reset` command structures.

## Impact

- **CLI/Command layer**: Extends `rootCmd` with new subcommands (`db-prune` and `reset` with nested subcommands `movie` and `tvshow`).
- **Database layer**: Introduces clean queries to delete records from `processed_lines`, `movies`, and `tvshows` safely.
- **API layer**: Adds two new routes to the Gin router with corresponding route handlers that fetch the entity, check for existence, delete associated `processed_lines`, and return a standard JSON success response.
