## MODIFIED Requirements

### Requirement: Frontend Interactive Manual Override Modal Dialog
The React frontend dashboard MUST provide an interactive, accessible modal dialog to trigger manual overrides for items in the playlist.
1. The modal MUST be opened by clicking an edit/search button next to any item in the playlist table, or by clicking the "Associate" action in the Track Details sidepanel (whether that sidepanel was opened from the Playlist tab or from the Radarr/Sonarr tab).
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
   - The candidate list MUST be sourced from a dedicated backend search scoped to the opened item's cleaned title (the same cleaning rule as step 5), restricted to `content_type` `"tvshows"` and excluding the entry the modal was opened for. This search MUST be independent of, and unaffected by, whatever page, filter, or sort any other view (e.g. the Playlist tab or a processing-run's item dialog) currently has loaded.
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
- **WHEN** the user opens the override modal for a `"Breaking Bad S01E01"` item, a dedicated search for that item's cleaned title ("Breaking Bad") returns `"Breaking Bad S01E02"` and `"Breaking Bad S02E01"` entries not yet associated with any TV show, and the user selects the "Breaking Bad" TMDB result
- **THEN** the candidate list SHALL show both other entries pre-checked, alongside any other search results left unchecked

#### Scenario: Candidates are found regardless of what any other view currently has loaded
- **WHEN** the user opens the override modal for a series item from the processing-run item dialog (or any other entry point), and the Playlist tab elsewhere in the app currently has a different page, filter, or sort loaded that does not include the item's other episodes
- **THEN** the candidate list SHALL still include those other episodes, because it is sourced from a dedicated search rather than whatever the Playlist tab happens to have loaded

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

#### Scenario: Opening the override modal from the Track Details sidepanel
- **WHEN** the user clicks "Associate" in the Track Details sidepanel for a given occurrence
- **THEN** the override modal SHALL open pre-targeted at that same occurrence, behaving identically to opening it from the Playlist table's row-level "Associate" button
