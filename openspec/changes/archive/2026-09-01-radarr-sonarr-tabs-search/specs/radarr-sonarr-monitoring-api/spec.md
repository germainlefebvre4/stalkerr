## ADDED Requirements

### Requirement: Optional search parameter filters both listing endpoints by title
The Radarr movies listing endpoint and the Sonarr series listing endpoint SHALL each accept an optional `search` query parameter that filters the monitored catalog to entries whose title case-insensitively contains the given term, applied before pagination.

#### Scenario: Search term filters the full catalog before pagination
- **WHEN** a listing request includes a `search` term
- **THEN** the endpoint SHALL filter the full monitored catalog by title (case-insensitive substring match) before slicing the requested page, and `total` SHALL reflect the filtered count

#### Scenario: No search term supplied
- **WHEN** a listing request omits `search` (or supplies an empty value)
- **THEN** the endpoint SHALL behave exactly as it does today, unfiltered

#### Scenario: Search term matches nothing
- **WHEN** a `search` term matches no entry in the monitored catalog
- **THEN** the endpoint SHALL return HTTP 200 with an empty `data` array and `total` of 0, not an error

### Requirement: Radarr full-catalog matched/unmatched movie stats
The system SHALL expose an endpoint reporting the total number of Radarr-monitored movies and how many of them have a matching local playlist entry, computed across the entire monitored catalog rather than a single page.

#### Scenario: Stats reflect the full monitored catalog
- **WHEN** the stats endpoint is called
- **THEN** it SHALL return the total monitored movie count and the matched count, both computed over every currently monitored movie, not just one page

#### Scenario: Matching stays local-only
- **WHEN** the stats endpoint computes the matched count
- **THEN** it SHALL determine match status using only the local playlist database (as the existing per-page match computation does), without additional live Radarr calls beyond fetching the monitored movie list

### Requirement: Sonarr monitored series total count without per-series fan-out
The system SHALL expose the total number of Sonarr-monitored series without fetching each series' episode list from Sonarr.

#### Scenario: Total count avoids per-series episode fetches
- **WHEN** the Sonarr monitored-series total is requested
- **THEN** the system SHALL derive it from the monitored series listing alone, without issuing an additional Sonarr episode-list request per series
