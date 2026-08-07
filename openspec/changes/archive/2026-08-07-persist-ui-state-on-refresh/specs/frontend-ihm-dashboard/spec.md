## MODIFIED Requirements

### Requirement: Real-time Monitoring Dashboard
The frontend SHALL present dedicated tabs to display active downloads (incorporating file-size progress bars) and execution logs, using polling to automatically keep content fresh. The interface MUST employ visual shimmers on loading and progress indicators, and active pulsing indicators for the API health state. The frontend SHALL persist the active tab state in the browser's `localStorage`, reflect it in the URL query string under the `tab` parameter, and restore it upon page refresh (preferring the `tab` URL parameter when present, falling back to `localStorage` otherwise).

#### Scenario: Display active download progress bar and logs status
- **WHEN** the user navigates to the "Downloads" view
- **THEN** the frontend SHALL render active progress bars indicating the percentage of completion with animated gradient shimmer effects, and automatically re-fetch data every 5 seconds.

#### Scenario: Persist and restore active tab on refresh
- **WHEN** the user selects a tab and refreshes the browser page
- **THEN** the frontend SHALL write the selected tab name to `localStorage` and to the URL's `tab` query parameter upon switch, and read/restore it upon mount to activate the correct view.

#### Scenario: Restore active tab from a shared URL
- **WHEN** a user opens a URL containing `?tab=downloads` (regardless of what is stored in their `localStorage`)
- **THEN** the frontend SHALL activate the "Downloads" tab, taking the `tab` query parameter as authoritative over any previously stored `localStorage` value.
