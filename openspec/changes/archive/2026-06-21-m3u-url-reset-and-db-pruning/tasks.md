## 1. Database Operations

- [x] 1.1 Implement pruning query helper inside processor or database package to fetch active hashes from M3U and delete expired `processed_lines` (soft and hard mode).
- [x] 1.2 Implement orphaned metadata cleanup query helper to delete unreferenced rows from `movies` and `tvshows` tables.
- [x] 1.3 Implement surgical reset database helper that deletes all `processed_lines` referencing a specific movie ID or TV show ID.

## 2. CLI Command Implementation

- [x] 2.1 Create `cmd/db_prune.go` and implement the `db-prune` command with `--dry-run` and `--hard` flags.
- [x] 2.2 Create `cmd/reset.go` and implement the `reset` command with `movie` and `tvshow` nested subcommands and an `--id` flag.
- [x] 2.3 Verify CLI registration and ensure `cmd/main.go` remains clean and unmodified.

## 3. HTTP REST API Endpoint Implementation

- [x] 3.1 Implement Gin router HTTP handler for resetting a movie (`resetMovieHandler`) in `internal/api/handlers.go`.
- [x] 3.2 Implement Gin router HTTP handler for resetting a TV show (`resetTVShowHandler`) in `internal/api/handlers.go`.
- [x] 3.3 Register the new POST endpoints (`/api/v1/movies/:id/reset` and `/api/v1/tvshows/:id/reset`) in `internal/api/api.go`.

## 4. Testing and Verification

- [x] 4.1 Write unit tests for the database pruning logic to verify soft/hard pruning and orphan cleanup.
- [x] 4.2 Write unit tests for the surgical reset helper.
- [x] 4.3 Write integration/handler tests for the reset API endpoints.
- [x] 4.4 Manually verify CLI commands `stalkeer db-prune --dry-run` and `stalkeer reset` on a test database environment.
