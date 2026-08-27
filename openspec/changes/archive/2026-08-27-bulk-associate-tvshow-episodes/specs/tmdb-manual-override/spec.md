## MODIFIED Requirements

### Requirement: Frontend Interactive Manual Override Modal Dialog
The React frontend dashboard MUST provide an interactive, accessible modal dialog to trigger manual overrides for items in the playlist.
1. The modal MUST be opened by clicking an edit/search button next to any item in the playlist table.
2. The modal MUST display the item's original title and group category.
3. The modal MUST provide strict selection between **Film (movie)** and **Série TV (tvshow)** modes.
4. The modal MUST pre-populate a search text field with a cleaned version of the item's original title and automatically initiate a search on loading.
5. Search results MUST be rendered with TMDB poster thumbnails (using the TMDB image CDN `https://image.tmdb.org/t/p/w92`), title, release year, note average, and description.
6. Clicking a result MUST mark it as selected.
7. If `"tvshow"` is chosen, optional "Saison" and "Épisode" input fields MUST be rendered, pre-populated from season/episode patterns extracted from the original title:
   - The extraction MUST recognize season/episode markers separated by zero or more whitespace or dash characters (e.g. `S02E25`, `S02 E25`, `S02-E25`), case-insensitively.
   - If both season and episode are confidently extracted, both fields MUST be pre-populated with the extracted values.
   - If a season/episode pattern cannot be confidently extracted (including the case where an episode marker is found without an associated season marker), both fields MUST be left empty rather than pre-populated with a guessed or default season.
8. If `"tvshow"` is chosen and a TMDB result is selected, the modal MUST additionally display a candidate list for bulk association:
   - The candidate list MUST contain every other currently loaded playlist entry (the page of results currently displayed in the playlist table) whose `content_type` is `"tvshows"`, excluding the entry the modal was opened for.
   - Each candidate MUST be rendered with its own raw title (`tvg_name`) and a checkbox, and MUST be independently checkable and uncheckable by the user.
   - A candidate MUST be pre-checked when its title, cleaned with the same cleaning rule applied to the search query (step 4), matches the cleaned title of the item the modal was opened for; all other candidates MUST start unchecked.
   - The pre-check heuristic is advisory only: the user MUST be able to check any unchecked candidate (to include an episode the heuristic missed) and uncheck any pre-checked candidate (to exclude a false-positive match) before confirming.
9. Clicking "Forcer l'association" MUST send the `POST /api/v1/items/:id/override` request for the item the modal was opened for, using the selected TMDB result, mode, and (for `"tvshow"`) season/episode fields.
10. If any candidates from the bulk list (step 8) are checked, clicking "Forcer l'association" MUST additionally send one `POST /api/v1/items/:id/override` request per checked candidate, sequentially, each using the same selected TMDB `tmdb_id` and `type: "tvshow"` and omitting `season`/`episode` so each candidate's own season/episode is auto-extracted from its own title by the backend, exactly as it already is for a single override without explicit season/episode.
11. Once all requests from steps 9-10 have completed (successfully or not), the modal MUST report a single outcome summary showing how many of the total requested associations succeeded (e.g. "11/12 associés"), and MUST identify which specific item(s) failed when at least one failed, rather than surfacing only a single generic success or error message.
12. The modal MUST close and refresh the playlist table only after the batch in steps 9-11 has completed, whether fully or partially successful; if every request in the batch failed, the modal MUST remain open and display the error(s) instead of closing.

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
