## MODIFIED Requirements

### Requirement: Configuration Page Container
Clicking the settings entry point SHALL navigate to a dedicated Configuration page, distinct from the Home/Downloads/Playlist/Logs tabs, organized into the following tabs, in order: Général, Intégrations, Contenu, Notifications, Avancé, Système. The page SHALL default to the "Général" tab when opened with no prior state, SHALL restore the tab most recently active in the Configuration page across sessions when opened with no tab identifier in the URL, and SHALL activate the tab identified by the URL when one is present. Selecting a tab SHALL update the page's URL to identify the newly active tab. The page SHALL provide an explicit way to return to the tab that was active in the main app before the settings icon was clicked, and SHALL restore that tab when the user does so, independently of which Configuration-page tab was last active.

#### Scenario: Navigating to the page shows all sections in order
- **WHEN** the user clicks the settings entry point with no previously remembered Configuration-page tab and no tab identifier in the URL
- **THEN** the app SHALL navigate to the Configuration page, display its six tabs in order Général, Intégrations, Contenu, Notifications, Avancé, Système, and activate "Général"

#### Scenario: Returning from the page preserves the previously active tab
- **WHEN** the user navigated to the Configuration page while a given main tab (e.g. "Playlist") was active, selected a different Configuration-page tab, and then returns
- **THEN** "Playlist" SHALL become the active main tab again, unchanged

#### Scenario: Expanding a collapsed section reveals its content in place
- **WHEN** the user selects a tab other than the currently active one
- **THEN** the page SHALL reveal that tab's content in place, without navigating away from the Configuration page

#### Scenario: Reopening restores the last active Configuration-page tab
- **WHEN** a user who previously had "Avancé" active as the Configuration page's tab reopens the Configuration page in a new session with no tab identifier in the URL
- **THEN** "Avancé" SHALL become the active tab again

#### Scenario: A URL identifies a specific tab
- **WHEN** the user navigates directly to a URL carrying the "Intégrations" tab identifier
- **THEN** the Configuration page SHALL open with "Intégrations" already active

#### Scenario: Selecting a tab updates the URL
- **WHEN** the user selects the "Notifications" tab while viewing the Configuration page
- **THEN** the page's URL SHALL update to identify "Notifications" as the active tab

### Requirement: About Section
The Configuration page's "Système" tab SHALL display the application's name and a link to its source repository, alongside that tab's health, disk, and build-version information.

#### Scenario: About section shows app identity and repository link
- **WHEN** the user views the "Système" tab
- **THEN** the tab SHALL display the application name and a link that opens the source repository

## ADDED Requirements

### Requirement: Configuration Overview Summary
The Configuration page SHALL display a summary banner above its tabs showing the total number of app-settings fields, M3U sources, and filter configurations currently carrying an active override, and, distinctly, how many of those active overrides are on a field flagged as requiring a restart to take effect. When no override is currently active anywhere on the page, the banner SHALL state that explicitly rather than showing an unexplained zero.

#### Scenario: Banner reflects the current total across all tabs
- **WHEN** the Configuration page is opened and 3 app-settings fields, 1 M3U source, and 2 filter configurations currently carry an active override
- **THEN** the banner SHALL display a total override count of 6

#### Scenario: Banner reflects restart-required overrides distinctly
- **WHEN** at least one currently active override is on a field flagged as requiring a restart
- **THEN** the banner SHALL display a count of overrides requiring a restart, separate from the total override count

#### Scenario: No active overrides
- **WHEN** no app-settings field, M3U source, or filter configuration currently carries an active override
- **THEN** the banner SHALL indicate that no settings are currently overridden

### Requirement: Per-Tab Override Indicators
Each of the Intégrations, Contenu, Notifications, and Avancé tab labels SHALL display a count of active overrides among the settings/sources it contains, and SHALL display a distinct marker when at least one of those active overrides requires a restart.

#### Scenario: Tab shows its own override count
- **WHEN** the "Intégrations" tab currently has 2 fields with an active override
- **THEN** its tab label SHALL display a badge showing "2"

#### Scenario: Tab shows no badge when it has no active overrides
- **WHEN** a tab currently has no field, source, or filter with an active override
- **THEN** its tab label SHALL display no override-count badge

#### Scenario: Tab shows a restart marker
- **WHEN** a tab has at least one active override on a field flagged as requiring a restart
- **THEN** its tab label SHALL display a distinct restart-required marker in addition to its override count

### Requirement: Quick Settings Search
The Configuration page SHALL provide a search input that filters the settings fields and Filtres/Sources M3U list entries visible on the active tab by matching the query text (case-insensitive substring match) against their translated label, hiding non-matching entries. The search query SHALL persist when the user switches tabs.

#### Scenario: Search filters fields within the active tab
- **WHEN** the user is on the "Intégrations" tab and types "api" into the search input
- **THEN** only fields whose label matches "api" SHALL remain visible on that tab, and other fields SHALL be hidden

#### Scenario: Clearing the search shows every field again
- **WHEN** the user clears the search input
- **THEN** every field in the active tab SHALL be visible again

#### Scenario: Search query persists across tab switches
- **WHEN** the user has typed a query while on one tab and switches to another tab
- **THEN** the same query SHALL remain in the search input and SHALL filter the newly active tab's visible entries

### Requirement: General Tab Layout
The "Général" tab's Apparence, Langue, and Préférences groups SHALL each render as a card, laid out using the same responsive, capped-width grid described by the `frontend-app-settings-management` Settings Field Layout requirement.

#### Scenario: Compact preference cards share a row
- **WHEN** the "Général" tab is displayed at a width sufficient for more than one 280px-minimum card
- **THEN** the Apparence, Langue, and Préférences cards SHALL pack into the available row width instead of each spanning it alone
