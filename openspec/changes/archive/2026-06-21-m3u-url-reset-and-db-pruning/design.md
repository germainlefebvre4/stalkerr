## Context

As M3U stream URLs expire or are rotated by providers, the Stalkeer database accumulates stale `processed_lines` (since hashes are computed using `tvgName + URL`). To maintain a clean database and prevent unlimited growth, we need a pruning mechanism (`stalkeer db-prune`).
Furthermore, to allow users to force a fresh URL retrieve/re-download of specific media directly from the planned React frontend UI, we need a surgical reset command (`stalkeer reset`) and matching HTTP REST API endpoints (`POST /api/v1/movies/:id/reset` and `POST /api/v1/tvshows/:id/reset`).

## Goals / Non-Goals

**Goals:**
- Provide a `db-prune` CLI command to clean up expired stream URLs and orphaned movies/tvshows.
- Provide a `reset` CLI command and API endpoints to surgically reset specific movies or TV shows by deleting their child `processed_lines`.
- Ensure the reset mechanism allows the next M3U processing run to automatically fetch fresh URLs and associate them to existing movie/tvshow entries without losing TMDB metadata.
- Maintain full compatibility with the existing database schema, model definitions, and GORM configurations.

**Non-Goals:**
- Do not modify the existing `line_hash` calculation algorithm.
- Do not implement real-time URL checking in this change (only pruning against the downloaded playlist file).

## Decisions

### Decision 1: Command and Code Organization
We will follow the `cmd-entrypoints` guidelines and create two new, self-contained Go files under `cmd/`:
1. `cmd/db_prune.go`: Houses the `db-prune` command.
2. `cmd/reset.go`: Houses the `reset` command (with subcommands `movie` and `tvshow`).

This ensures `cmd/main.go` remains untouched and minimal.

### Decision 2: Pruning Logic
The `db-prune` command will load the active M3U file, parse it to extract active stream hashes, and run database deletion operations:
- **Soft Prune (Default)**:
  ```sql
  DELETE FROM processed_lines 
  WHERE line_hash NOT IN (<active_hashes>) 
    AND state NOT IN ('downloaded', 'downloading');
  ```
- **Hard Prune (`--hard`)**:
  ```sql
  DELETE FROM processed_lines 
  WHERE line_hash NOT IN (<active_hashes>);
  ```
- **Post-Prune Orphan Cleanup**:
  ```sql
  DELETE FROM movies WHERE id NOT IN (SELECT DISTINCT movie_id FROM processed_lines WHERE movie_id IS NOT NULL);
  DELETE FROM tvshows WHERE id NOT IN (SELECT DISTINCT tv_show_id FROM processed_lines WHERE tv_show_id IS NOT NULL);
  ```

### Decision 3: API Endpoint Integration
In `internal/api/handlers.go`, we will implement two new HTTP handlers:
1. `resetMovieHandler(c *gin.Context)`: Bound to `POST /api/v1/movies/:id/reset`
2. `resetTVShowHandler(c *gin.Context)`: Bound to `POST /api/v1/tvshows/:id/reset`

These handlers will:
- Check if the entity (`Movie` or `TVShow`) exists in the database.
- If not, return a `404 Not Found`.
- Delete all `processed_lines` referencing the target entity ID (e.g. `db.Where("movie_id = ?", id).Delete(&models.ProcessedLine{})`).
- Return a `200 OK` with `{"status": "success", "message": "Media reset successfully"}`.

## Risks / Trade-offs

- **[Risk] Orphaning Active Downloads** → If we hard-prune a line while a download is running, the download state will diverge.
  - *Mitigation*: We will use a soft prune by default which protects both `downloaded` and `downloading` states. We will also log a warning if any pruned line has active downloads.
- **[Risk] TMDB API Rate Limits on Re-import** → If a user resets many items or performs a hard prune that orphans and deletes many movies, subsequent imports will trigger many TMDB API calls.
  - *Mitigation*: The surgical reset deletes ONLY the `processed_lines` while keeping the `movies` and `tvshows` tables intact, preserving TMDB metadata completely and avoiding TMDB queries. The `db-prune` command only deletes metadata records that are completely unreferenced (i.e. no longer in the M3U at all).
