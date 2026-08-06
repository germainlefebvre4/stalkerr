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
8. Clicking "Forcer l'association" MUST send the `POST /api/v1/items/:id/override` request, display a success banner, close the modal, and refresh the playlist table.

#### Scenario: User searches and forces an association for a TV show
- **WHEN** the user opens the override modal for a series item, chooses "Série TV", cleans the search to "Malcolm", selects the correct show from the poster results, specifies season `1` and episode `5`, and clicks "Forcer l'association"
- **THEN** the frontend SHALL issue the POST request to the backend override endpoint, close the modal on success, and refresh the list to show the newly matched TV Show metadata

#### Scenario: Season and episode separated by a space in the raw title
- **WHEN** the user opens the override modal for a series item whose raw title is `"Inspecteur Gadget S02 E25"`
- **THEN** the "Saison" field SHALL be pre-populated with `2` and the "Épisode" field SHALL be pre-populated with `25`

#### Scenario: Episode marker found without a recognizable season marker
- **WHEN** the user opens the override modal for a series item whose raw title contains an episode marker (e.g. `E25`) but no recognizable season marker
- **THEN** both the "Saison" and "Épisode" fields SHALL be left empty rather than pre-populated with a guessed season
