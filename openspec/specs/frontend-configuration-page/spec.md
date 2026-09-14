# frontend-configuration-page Specification

## Purpose

Gives the dashboard a single, discoverable home — one header entry point and a dedicated Configuration page — for cross-cutting preferences (appearance, language, startup defaults), filters management, system diagnostics, and the application's backend-settable configuration (Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads, logging, M3U sources).

## Requirements

### Requirement: Settings Header Entry Point
The app header SHALL display exactly one settings entry point (icon), positioned identically on desktop and mobile viewports: sharing the same header row as the app title on both. This entry point SHALL be visually distinct from the icon used by the "Logs" tab.

#### Scenario: Single entry point on desktop
- **WHEN** the app is viewed on a desktop-width viewport
- **THEN** the header SHALL show exactly one settings icon

#### Scenario: Single entry point shares the title row on mobile
- **WHEN** the app is viewed on a mobile-width viewport
- **THEN** the settings icon SHALL be visible on the same header row as the app title, without wrapping to its own row

#### Scenario: Icon is distinct from the Logs tab icon
- **WHEN** the header and the tab bar (desktop segmented tabs or mobile bottom tab bar) are both visible
- **THEN** the settings icon SHALL be visually distinguishable from the "Logs" tab's icon

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

### Requirement: Theme Toggle
The Configuration page's "Apparence" section SHALL provide a light/dark theme toggle. Selecting a theme SHALL immediately re-render the app using that theme's tokens and SHALL persist the choice across sessions. Absent a stored preference, the app SHALL use the light theme.

#### Scenario: Switching to dark theme re-renders immediately
- **WHEN** the user selects the dark option in the Apparence section
- **THEN** the frontend SHALL immediately apply the dark theme's tokens across the app and persist the choice

#### Scenario: Stored theme preference is restored on next visit
- **WHEN** a user who previously selected the dark theme reopens the dashboard in a new browser session
- **THEN** the frontend SHALL render in the dark theme without requiring the user to switch again

#### Scenario: No stored preference defaults to light
- **WHEN** a user opens the dashboard for the first time, with no theme preference stored
- **THEN** the frontend SHALL render in the light theme

### Requirement: Reduce Motion Toggle
The Configuration page's "Apparence" section SHALL provide a "reduce animations" toggle. When enabled, the frontend SHALL disable the toast notification's entrance animation and the card hover lift/transition effect, and SHALL persist the choice across sessions.

#### Scenario: Enabling reduce animations disables the toast entrance animation
- **WHEN** the user enables "reduce animations" and an action subsequently triggers a toast notification
- **THEN** the toast SHALL appear without its entrance animation

#### Scenario: Enabling reduce animations disables the card hover lift
- **WHEN** the user enables "reduce animations"
- **THEN** cards SHALL NOT apply their hover lift/transform/shadow transition

#### Scenario: Preference persists across sessions
- **WHEN** a user who previously enabled "reduce animations" reopens the dashboard in a new browser session
- **THEN** the frontend SHALL keep animations reduced without requiring the user to re-enable the toggle

### Requirement: Default Startup Tab Preference
The Configuration page's "Préférences" section SHALL display and let the user set the tab the app opens to on its next load (absent a `tab` URL parameter), using the same persisted value that normal tab navigation already updates. Setting it from "Préférences" SHALL NOT change the currently active tab in the session in which it was set.

#### Scenario: Setting the default tab does not navigate the current session
- **WHEN** the user selects "Downloads" as the default startup tab in Préférences while "Home" is the currently active tab
- **THEN** the frontend SHALL persist "downloads" as the startup tab, and "Home" SHALL remain the active tab in the current session

#### Scenario: Default tab is used on next load
- **WHEN** the app is reloaded with no `tab` URL parameter, after the user set "Downloads" as the default startup tab
- **THEN** the frontend SHALL open directly to the Downloads tab

### Requirement: Default Playlist Page Size and View Preferences
The Configuration page's "Préférences" section SHALL let the user set the default Playlist page size (10, 50, or 100 entries) and the default Playlist view (list or grouped), using the same persisted preferences the Playlist tab's own controls already read and write. Changing either value from "Préférences" SHALL immediately apply if the Playlist tab is currently mounted, and SHALL otherwise be used the next time the Playlist tab is opened.

#### Scenario: Changing page size applies live when Playlist is open
- **WHEN** the Playlist tab is currently active and the user changes the default page size to `50` in Préférences
- **THEN** the Playlist tab's currently displayed page size SHALL update to `50`

#### Scenario: Changing default view applies without visiting Playlist
- **WHEN** the Playlist tab is not currently mounted and the user sets the default view to "Grouped" in Préférences
- **THEN** the frontend SHALL persist "grouped" as the default view, and the Playlist tab SHALL open in the grouped view the next time it is activated

### Requirement: About Section
The Configuration page's "Système" tab SHALL display the application's name and a link to its source repository, alongside that tab's health, disk, and build-version information.

#### Scenario: About section shows app identity and repository link
- **WHEN** the user views the "Système" tab
- **THEN** the tab SHALL display the application name and a link that opens the source repository

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
