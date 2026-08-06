## MODIFIED Requirements

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
8. If steps 3-6 fail for a reason other than "item not found" or "TMDB disabled" (e.g. the TMDB detail fetch fails, or the database upsert fails), the response SHALL include a distinct machine-readable error code of `"override_failed"` so that clients can render a specific, translatable error message, separate from the existing `"not_found"` and `"tmdb_disabled"` codes.

#### Scenario: Forcing manual association to a movie VOD
- **WHEN** a client calls `POST /api/v1/items/42/override` with JSON body `{"tmdb_id": 27205, "type": "movie"}`
- **THEN** the system SHALL create or fetch local Movie `27205`, update ProcessedLine `42` to use `movies` ContentType and link to Movie `27205`, create or update a persistent `ManualMapping` for `ProcessedLine` `42`'s original title/group, set `override_by` to `"manual"`, and return a 200 OK response with the updated item

#### Scenario: Override fails because TMDB detail lookup errors out
- **WHEN** a client calls `POST /api/v1/items/42/override` with a valid `tmdb_id` and `type`, but the TMDB metadata fetch for that ID fails
- **THEN** the system SHALL return an error response with `ErrorResponse.error` set to `"override_failed"`, distinct from `"not_found"` and `"tmdb_disabled"`
