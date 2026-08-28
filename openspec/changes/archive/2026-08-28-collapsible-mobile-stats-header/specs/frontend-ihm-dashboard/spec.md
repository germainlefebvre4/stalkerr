## MODIFIED Requirements

### Requirement: Statistics KPI Cards
The frontend SHALL display a horizontal grid of 4 visual statistics cards beneath the main header, fetching data from `/api/v1/stats`. These cards must display:
- Total items in the M3U playlist.
- Number of movies identified.
- Number of TV shows identified.
- Download success percentage (downloaded count vs failed count).

Below the mobile breakpoint defined by `frontend-responsive-layout`, this grid SHALL render collapsed by default behind a tap-to-expand toggle instead of always showing all 4 cards, per the "Responsive Card Density" requirement in `frontend-responsive-layout`.

#### Scenario: Display global stats on dashboard load
- **WHEN** the dashboard loads or is refreshed
- **THEN** the frontend SHALL fetch statistics from `/api/v1/stats` and render them in the KPI grid with modern, translucent pastel backgrounds.

#### Scenario: Stats fetch happens regardless of collapsed state on mobile
- **WHEN** the dashboard loads on a viewport narrower than the mobile breakpoint
- **THEN** the frontend SHALL still fetch statistics from `/api/v1/stats` even though the KPI grid renders collapsed, so the values are already available the moment the user expands it.
