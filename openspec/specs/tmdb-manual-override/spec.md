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
8. If steps 3-6 fail for a reason other than "item not found" or "TMDB disabled" (e.g. the TMDB detail fetch fails, or the database upsert fails), the response SHALL include a distinct machine-readable error code of `"override_failed"` so that clients can render a specific, translatable error message, separate from the existing `"not_found"` and `"tmdb_disabled"` codes.

#### Scenario: Forcing manual association to a movie VOD
- **WHEN** a client calls `POST /api/v1/items/42/override` with JSON body `{"tmdb_id": 27205, "type": "movie"}`
- **THEN** the system SHALL create or fetch local Movie `27205`, update ProcessedLine `42` to use `movies` ContentType and link to Movie `27205`, create or update a persistent `ManualMapping` for `ProcessedLine` `42`'s original title/group, set `override_by` to `"manual"`, and return a 200 OK response with the updated item

#### Scenario: Override fails because TMDB detail lookup errors out
- **WHEN** a client calls `POST /api/v1/items/42/override` with a valid `tmdb_id` and `type`, but the TMDB metadata fetch for that ID fails
- **THEN** the system SHALL return an error response with `ErrorResponse.error` set to `"override_failed"`, distinct from `"not_found"` and `"tmdb_disabled"`

---

### Requirement: Frontend Interactive Manual Override Modal Dialog
The React frontend dashboard MUST provide an interactive, accessible modal dialog to trigger manual overrides for items in the playlist.
1. The modal MUST be opened by clicking an edit/search button next to any item in the playlist table.
2. The modal MUST display the item's original title and group category.
3. If the item already has an associated `movie` or `tvshow` when the modal opens, the modal MUST display a persistent "current match" panel showing the current TMDB title and year, and - for a `tvshow` association - the current season and episode, plus who and when it was last overridden (`override_by`/`override_at`) when those are set. This panel MUST remain visible regardless of search input changes or TMDB result selection, and MUST NOT require a new TMDB result to be selected. It applies identically whether the current association came from automatic pipeline matching, a prior single-item override, or a prior bulk-associate batch.
4. The modal MUST provide strict selection between **Film (movie)** and **Série TV (tvshow)** modes.
5. The modal MUST pre-populate a search text field with a cleaned version of the item's original title and automatically initiate a search on loading.
6. Search results MUST be rendered with TMDB poster thumbnails (using the TMDB image CDN `https://image.tmdb.org/t/p/w92`), title, release year, note average, and description.
7. Clicking a result MUST mark it as selected.
8. If `"tvshow"` is chosen, "Saison" and "Épisode" input fields MUST be rendered and pre-populated from season/episode patterns extracted from the original title as soon as `"tvshow"` mode is active for the item - independent of whether a TMDB result has been selected yet:
   - The extraction MUST recognize season/episode markers separated by zero or more whitespace or dash characters (e.g. `S02E25`, `S02 E25`, `S02-E25`), case-insensitively.
   - If both season and episode are confidently extracted, both fields MUST be pre-populated with the extracted values.
   - If a season/episode pattern cannot be confidently extracted (including the case where an episode marker is found without an associated season marker), both fields MUST be left empty rather than pre-populated with a guessed or default season.
9. If `"tvshow"` is chosen and a TMDB result is selected, the modal MUST additionally display a candidate list for bulk association:
   - The candidate list MUST contain every other currently loaded playlist entry (the page of results currently displayed in the playlist table) whose `content_type` is `"tvshows"`, excluding the entry the modal was opened for.
   - Each candidate MUST be rendered with its own raw title (`tvg_name`), a checkbox, and its own match-state preview: if the candidate already has an associated `tvshow`, its current season/episode; otherwise, the season/episode detected from its own raw title using the same extraction rule as step 8. The candidate MUST be independently checkable and uncheckable by the user.
   - A candidate MUST be pre-checked when its title, cleaned with the same cleaning rule applied to the search query (step 5), matches the cleaned title of the item the modal was opened for; all other candidates MUST start unchecked.
   - The pre-check heuristic is advisory only: the user MUST be able to check any unchecked candidate (to include an episode the heuristic missed) and uncheck any pre-checked candidate (to exclude a false-positive match) before confirming.
10. Clicking "Forcer l'association" MUST send the `POST /api/v1/items/:id/override` request for the item the modal was opened for, using the selected TMDB result, mode, and (for `"tvshow"`) season/episode fields.
11. If any candidates from the bulk list (step 9) are checked, clicking "Forcer l'association" MUST additionally send one `POST /api/v1/items/:id/override` request per checked candidate, sequentially, each using the same selected TMDB `tmdb_id` and `type: "tvshow"` and omitting `season`/`episode` so each candidate's own season/episode is auto-extracted from its own title by the backend, exactly as it already is for a single override without explicit season/episode.
12. Once all requests from steps 10-11 have completed (successfully or not), the modal MUST report a single outcome summary showing how many of the total requested associations succeeded (e.g. "11/12 associés"), and MUST identify which specific item(s) failed when at least one failed, rather than surfacing only a single generic success or error message.
13. The modal MUST close and refresh the playlist table only after the batch in steps 10-12 has completed, whether fully or partially successful; if every request in the batch failed, the modal MUST remain open and display the error(s) instead of closing.

#### Scenario: User searches and forces an association for a TV show
- **WHEN** the user opens the override modal for a series item, chooses "Série TV", cleans the search to "Malcolm", selects the correct show from the poster results, specifies season `1` and episode `5`, and clicks "Forcer l'association"
- **THEN** the frontend SHALL issue the POST request to the backend override endpoint, close the modal on success, and refresh the list to show the newly matched TV Show metadata

#### Scenario: Season and episode separated by a space in the raw title
- **WHEN** the user opens the override modal for a series item whose raw title is `"Inspecteur Gadget S02 E25"`
- **THEN** the "Saison" field SHALL be pre-populated with `2` and the "Épisode" field SHALL be pre-populated with `25`

#### Scenario: Episode marker found without a recognizable season marker
- **WHEN** the user opens the override modal for a series item whose raw title contains an episode marker (e.g. `E25`) but no recognizable season marker
- **THEN** both the "Saison" and "Épisode" fields SHALL be left empty rather than pre-populated with a guessed season

#### Scenario: Other episodes of the same series are pre-selected as candidates
- **WHEN** the user opens the override modal for a `"Breaking Bad S01E01"` item, the currently loaded playlist page also contains `"Breaking Bad S01E02"` and `"Breaking Bad S02E01"` entries not yet associated with any TV show, and the user selects the "Breaking Bad" TMDB result
- **THEN** the candidate list SHALL show both other entries pre-checked, alongside any other `tvshows` entries on the page left unchecked

#### Scenario: User deselects a false-positive candidate before confirming
- **WHEN** the candidate list pre-checks an entry whose title only coincidentally resembles the selected series (e.g. a same-named special or an unrelated show), and the user unchecks it before clicking "Forcer l'association"
- **THEN** that entry SHALL NOT be included in the batch of override requests, and SHALL remain unassociated afterwards

#### Scenario: User manually adds a candidate the heuristic did not suggest
- **WHEN** the candidate list leaves an actual episode of the target series unchecked because its raw title did not match the cleaning heuristic, and the user checks it manually before clicking "Forcer l'association"
- **THEN** that entry SHALL be included in the batch of override requests using the selected TMDB result

#### Scenario: Bulk association partially fails
- **WHEN** the user confirms an association for the opened item plus 3 checked candidates, and one of the 4 underlying override requests fails (e.g. a transient backend error)
- **THEN** the modal SHALL report that 3 of 4 associations succeeded, identify the failing item, and SHALL still refresh the playlist table to reflect the 3 successful associations

#### Scenario: Reopening the modal for an item already matched by the automatic pipeline
- **WHEN** the user opens the override modal for an item that was matched automatically during import and never manually touched (`override_by` is unset)
- **THEN** the current-match panel SHALL show that item's current TMDB title, year, season, and episode, without waiting for a new search result to be selected

#### Scenario: Reopening the modal for an item previously associated via bulk-associate
- **WHEN** the user opens the override modal for an item that was corrected as a checked candidate in a prior bulk-associate batch
- **THEN** the current-match panel SHALL show the TMDB title, season, and episode that batch applied, and SHALL show `override_by`/`override_at` reflecting that manual correction

#### Scenario: Detected season/episode preview shown before any result is selected
- **WHEN** the user opens the override modal for a series item whose raw title is `"Breaking Bad S02E05"`, before selecting any TMDB search result
- **THEN** the "Saison" and "Épisode" fields SHALL already show `2` and `5`

#### Scenario: Bulk candidate row previews its own season/episode before confirming
- **WHEN** the candidate list is shown and includes an unassociated entry whose raw title is `"Breaking Bad S02E06"`
- **THEN** that candidate's row SHALL display season `2` and episode `6` as the season/episode it will receive if checked and the batch is confirmed

#### Scenario: Bulk candidate row shows its existing match when already associated
- **WHEN** the candidate list includes an entry that is already associated with a `tvshow` (e.g. from an earlier correction)
- **THEN** that candidate's row SHALL display its current season/episode from that existing association rather than a freshly detected value

