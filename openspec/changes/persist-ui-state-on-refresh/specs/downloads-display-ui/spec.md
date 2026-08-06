## ADDED Requirements

### Requirement: Filter Persistence Across Refresh
The frontend SHALL reflect the Downloads tab's status, type, and problem filters in the browser URL's query string under the parameters `dlStatus`, `dlType`, and `dlProblem` respectively, and SHALL restore this exact filter selection from the URL on page load or refresh. These parameter names SHALL NOT be reused by any other tab's filters (e.g. the Playlist tab's `type` and `state` parameters), so that navigating between tabs or sharing a URL cannot cross-apply one tab's filter values to another.

#### Scenario: Restore Downloads filters from the URL after a refresh
- **WHEN** the user has selected the "Échoués" status filter and the "Séries" type filter on the Downloads tab, and refreshes the browser
- **THEN** the frontend SHALL read `dlStatus` and `dlType` from the URL query string and re-render the Downloads list with the same filters applied, instead of resetting to "Tous".

#### Scenario: Downloads and Playlist filters do not collide
- **WHEN** the user has an active Playlist content-type filter (`type=movies`) and switches to the Downloads tab and selects the "Séries" type filter
- **THEN** the frontend SHALL write the Downloads type filter to `dlType=tvshows` without altering or removing the Playlist's `type=movies` query parameter.
