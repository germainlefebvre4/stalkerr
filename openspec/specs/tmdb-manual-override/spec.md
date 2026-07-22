# tmdb-manual-override Specification

## Purpose
TBD - created by archiving change tmdb-manual-override-ui. Update Purpose after archive.
## Requirements
### Requirement: Database schema updates for overrides and manual mapping
The system MUST support database persistence for manual overrides and permanent user mappings.
1. The `processed_lines` table GORM model (`ProcessedLine`) MUST include:
   - `OverrideBy` (nullable varchar(50)) to track who/what made the correction.
   - `OverrideAt` (nullable timestamp) to record when the correction was applied.
2. A new `manual_mappings` table GORM model (`ManualMapping`) MUST be introduced, containing:
   - `ID` (Primary Key, Auto-increment).
   - `TvgName` (varchar(255), not null, composite unique index `idx_manual_mappings_unique` with `GroupTitle`).
   - `GroupTitle` (varchar(255), not null, composite unique index `idx_manual_mappings_unique` with `TvgName`).
   - `ContentType` (varchar(20), not null, representing `"movies"` or `"tvshows"`).
   - `TMDBID` (integer, not null, the selected TMDB ID).
   - `Season` (nullable integer, for series).
   - `Episode` (nullable integer, for series).
   - `CreatedAt` and `UpdatedAt` timestamps.
3. The database auto-migration function MUST automatically create and update these tables.

#### Scenario: Running database auto-migrations
- **WHEN** the server starts up and runs migrations
- **THEN** the `processed_lines` table SHALL contain the new columns `override_by` and `override_at`, and the `manual_mappings` table SHALL be created with a composite unique index on `(tvg_name, group_title)`

---

### Requirement: Secure TMDB Search Proxy API Endpoint
The backend API MUST expose a secure proxy endpoint `GET /api/v1/tmdb/search` to query movies or TV shows on TMDB without exposing the server's TMDB API key to the frontend client.
1. It MUST accept the following query parameters:
   - `query` (string, required): The search text.
   - `type` (string, required): MUST be `"movie"` or `"tvshow"`.
   - `year` (integer, optional): Only valid when `type` is `"movie"`.
2. If TMDB integration is disabled or not configured in `config.yml`, the endpoint MUST return `503 Service Unavailable` with error code `"tmdb_disabled"`.
3. If successful, the endpoint MUST return a JSON list of matches containing `id`, `title`/`name`, `original_title`/`original_name`, `release_date`/`first_air_date`, `overview`, and `poster_path`.

#### Scenario: Successfully searching a movie on TMDB
- **WHEN** a client calls `GET /api/v1/tmdb/search?query=Inception&type=movie`
- **THEN** the system SHALL return a 200 OK response with a list of TMDB movie search results containing titles and poster paths

#### Scenario: Searching TMDB when TMDB is disabled
- **WHEN** TMDB is disabled in the configuration and a client calls `GET /api/v1/tmdb/search?query=Inception&type=movie`
- **THEN** the system SHALL return a 503 Service Unavailable response with error code `"tmdb_disabled"`

---

### Requirement: Manual Override API Endpoint
The backend API MUST expose an endpoint `POST /api/v1/items/:id/override` to manually associate a VOD item with a specific TMDB movie or TV show.
1. It MUST accept a JSON body containing:
   - `tmdb_id` (integer, required).
   - `type` (string, required, either `"movie"` or `"tvshow"`).
   - `season` (integer, optional).
   - `episode` (integer, optional).
2. It MUST fetch the `ProcessedLine` by ID. If not found, return `404 Not Found`.
3. If `type` is `"movie"`:
   - It MUST fetch detailed movie metadata and external IDs from TMDB, then find or create the local `Movie` GORM record.
   - It MUST associate the `ProcessedLine` with this `Movie`'s ID, clear any other relations (`TVShowID = nil`, `ChannelID = nil`, `UncategorizedID = nil`), and set `ContentType` to `"movies"`.
4. If `type` is `"tvshow"`:
   - It MUST fetch detailed TV show metadata and external IDs from TMDB.
   - It MUST extract the season and episode (from request body if provided, or from the original title using the classifier as fallback), then find or create the local `TVShow` GORM record.
   - It MUST associate the `ProcessedLine` with this `TVShow`'s ID, clear other relations (`MovieID = nil`, `ChannelID = nil`, `UncategorizedID = nil`), and set `ContentType` to `"tvshows"`.
5. It MUST upsert a `ManualMapping` record in the database for `(TvgName, GroupTitle)` pointing to this TMDB ID and configuration, so that future imports of this identical title automatically apply the correction.
6. It MUST set `OverrideBy = "manual"` and `OverrideAt = time.Now()`.
7. It MUST return the updated item representation.

#### Scenario: Forcing manual association to a movie VOD
- **WHEN** a client calls `POST /api/v1/items/42/override` with JSON body `{"tmdb_id": 27205, "type": "movie"}`
- **THEN** the system SHALL create or fetch local Movie `27205`, update ProcessedLine `42` to use `movies` ContentType and link to Movie `27205`, create or update a persistent `ManualMapping` for `ProcessedLine` `42`'s original title/group, set `override_by` to `"manual"`, and return a 200 OK response with the updated item

---

### Requirement: Frontend Interactive Manual Override Modal Dialog
The React frontend dashboard MUST provide an interactive, accessible modal dialog to trigger manual overrides for items in the playlist.
1. The modal MUST be opened by clicking an edit/search button next to any item in the playlist table.
2. The modal MUST display the item's original title and group category.
3. The modal MUST provide strict selection between **Film (movie)** and **Série TV (tvshow)** modes.
4. The modal MUST pre-populate a search text field with a cleaned version of the item's original title and automatically initiate a search on loading.
5. Search results MUST be rendered with TMDB poster thumbnails (using the TMDB image CDN `https://image.tmdb.org/t/p/w92`), title, release year, note average, and description.
6. Clicking a result MUST mark it as selected.
7. If `"tvshow"` is chosen, optional "Saison" and "Épisode" input fields MUST be rendered, pre-populated using season/episode patterns extracted from the original title.
8. Clicking "Forcer l'association" MUST send the `POST /api/v1/items/:id/override` request, display a success banner, close the modal, and refresh the playlist table.

#### Scenario: User searches and forces an association for a TV show
- **WHEN** the user opens the override modal for a series item, chooses "Série TV", cleans the search to "Malcolm", selects the correct show from the poster results, specifies season `1` and episode `5`, and clicks "Forcer l'association"
- **THEN** the frontend SHALL issue the POST request to the backend override endpoint, close the modal on success, and refresh the list to show the newly matched TV Show metadata

