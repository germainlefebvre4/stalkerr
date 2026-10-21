# Spec Delta

## MODIFIED Requirements

### Requirement: Downloads Filter Option Sets
The Downloads tab SHALL offer three filter controls — status, type, and problem — each with a fixed set of options mapping to specific backend query values: the status filter maps "Actifs" / "Tous" / "Complétés" / "En cours" / "En attente" / "Réessai" / "Échoués" / "Annulés" to `active` / no filter / `completed` / `downloading` / `pending` / `retrying` / `failed` / `cancelled`; the type filter maps "Tous" / "Films" / "Séries" to no filter / `movies` / `tvshows`; the problem filter maps "Aucun" / "Année manquante" / "Année incorrecte" / "Format inconnu" / "Basse qualité" to no filter / `missing_year` / `year_mismatch` / `unknown_format` / `low_quality`.

#### Scenario: Selecting a status filter option applies the mapped backend value
- **WHEN** the user selects "Échoués" in the status filter
- **THEN** the frontend SHALL query downloads with `status=failed`

#### Scenario: Selecting the Actifs option excludes cancelled downloads
- **WHEN** the user selects "Actifs" in the status filter
- **THEN** the frontend SHALL query downloads with `status=active`

#### Scenario: Selecting a problem filter option applies the mapped backend value
- **WHEN** the user selects "Année manquante" in the problem filter
- **THEN** the frontend SHALL query downloads with `problem=missing_year`

#### Scenario: "Tous" resets a filter to no constraint
- **WHEN** the user selects "Tous" in the status or type filter
- **THEN** the frontend SHALL query downloads without that filter parameter

## ADDED Requirements

### Requirement: Active Status Is the Default Filter
When the URL carries no `dlStatus` parameter, the Downloads tab SHALL select the "Actifs" status option and query downloads with `status=active`, so that cancelled downloads are not listed until the user chooses "Tous" or "Annulés". Choosing "Tous" SHALL be persisted in the URL as `dlStatus=all` so that a refresh keeps showing every status instead of falling back to "Actifs".

#### Scenario: First load hides cancelled downloads
- **WHEN** the user opens the Downloads tab with no `dlStatus` in the URL
- **THEN** the status filter SHALL show "Actifs" and the list SHALL contain no `cancelled` download

#### Scenario: Choosing Tous survives a refresh
- **WHEN** the user selects "Tous" and refreshes the browser
- **THEN** the URL SHALL contain `dlStatus=all`, the status filter SHALL show "Tous", and cancelled downloads SHALL be listed

#### Scenario: A legacy URL with an explicit status is honored
- **WHEN** the user opens a URL containing `dlStatus=failed`
- **THEN** the status filter SHALL show "Échoués" and only failed downloads SHALL be listed

### Requirement: Downloads Title Search Input
The Downloads tab SHALL provide a text search input that filters downloads by title through the `search` query parameter of `GET /api/v1/downloads`. The value SHALL be persisted in the URL as `dlSearch`, restored on load or refresh, and applied after a debounce of the user's typing. Changing the search text SHALL reset the list to page 1. The tab SHALL also read the Playlist-style `searchName` URL parameter as an alias of `dlSearch` when `dlSearch` is absent.

#### Scenario: Typing a title narrows the list
- **WHEN** the user types "peacemaker" in the search input
- **THEN** the frontend SHALL re-fetch downloads with `search=peacemaker` and page 1, and set `dlSearch=peacemaker` in the URL

#### Scenario: Search restored after refresh
- **WHEN** the user refreshes the browser on a URL containing `dlSearch=peacemaker`
- **THEN** the search input SHALL show "peacemaker" and the list SHALL be filtered by it

#### Scenario: searchName alias
- **WHEN** the user opens `?tab=downloads&searchName=peacemaker` without `dlSearch`
- **THEN** the search input SHALL show "peacemaker" and the list SHALL be filtered by it

#### Scenario: Clearing the search removes the constraint
- **WHEN** the user empties the search input
- **THEN** the frontend SHALL re-fetch downloads without `search` and remove `dlSearch` from the URL

### Requirement: Hidden Cancelled Downloads Indicator
When the Downloads tab is filtered on "Actifs" and the response reports a `cancelled_hidden` count greater than zero, the tab SHALL display an indicator reading "N annulés masqués" (localized) below the list. Activating it SHALL switch the status filter to "Annulés" while keeping the type, problem and search filters. The indicator SHALL NOT be shown when the count is zero or when the status filter is not "Actifs".

#### Scenario: Indicator shows the hidden count
- **WHEN** the status filter is "Actifs" and the response has `cancelled_hidden` equal to `3`
- **THEN** the tab SHALL display "3 annulés masqués"

#### Scenario: Activating the indicator reveals cancelled downloads
- **WHEN** the user activates the "3 annulés masqués" indicator
- **THEN** the status filter SHALL switch to "Annulés", the search, type and problem filters SHALL be unchanged, and the list SHALL show the cancelled downloads

#### Scenario: No indicator without hidden downloads
- **WHEN** the response has `cancelled_hidden` equal to `0`, or the status filter is not "Actifs"
- **THEN** the tab SHALL NOT display the indicator
