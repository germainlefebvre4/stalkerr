## MODIFIED Requirements

### Requirement: Configuration Page Container
Clicking the settings entry point SHALL navigate to a dedicated Configuration page, distinct from the Home/Downloads/Playlist/Logs tabs, organized into the following tabs, in order: Général, Intégrations, Contenu, Notifications, Avancé, Système. The page SHALL default to the "Général" tab when opened with no prior state, SHALL restore the tab most recently active in the Configuration page across sessions when opened with no tab identifier in the URL, and SHALL activate the tab identified by the URL when one is present. Selecting a tab SHALL update the page's URL to identify the newly active tab. The page SHALL provide an explicit way to return to the tab that was active in the main app before the settings icon was clicked, and SHALL restore that tab when the user does so, independently of which Configuration-page tab was last active. Below the mobile breakpoint, the page SHALL present these six tabs as a native `<select>` dropdown instead of the desktop tab bar, and SHALL NOT require horizontal scrolling to reach any tab; selecting an option in that dropdown SHALL switch the active tab exactly as selecting a desktop tab does, including the URL update.

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

#### Scenario: Mobile viewport shows a dropdown instead of a scrolling tab bar
- **WHEN** the Configuration page is viewed on a mobile-width viewport
- **THEN** the frontend SHALL render a `<select>` listing the six tabs instead of the desktop tab bar, and reaching any of the six tabs SHALL NOT require horizontal scrolling

#### Scenario: Selecting a tab from the mobile dropdown switches the active tab
- **WHEN** the user is on a mobile-width viewport and chooses "Système" from the tab dropdown
- **THEN** the frontend SHALL activate the "Système" tab and update the page's URL to identify it, exactly as selecting the "Système" desktop tab would

### Requirement: Per-Tab Override Indicators
Each of the Intégrations, Contenu, Notifications, and Avancé tab labels SHALL display a count of active overrides among the settings/sources it contains, and SHALL display a distinct marker when at least one of those active overrides requires a restart. On the mobile dropdown, each of these tabs' option text SHALL include the same override count and, when applicable, the same restart-required signal, rendered as text since the option cannot carry a separate visual badge.

#### Scenario: Tab shows its own override count
- **WHEN** the "Intégrations" tab currently has 2 fields with an active override
- **THEN** its tab label SHALL display a badge showing "2"

#### Scenario: Tab shows no badge when it has no active overrides
- **WHEN** a tab currently has no field, source, or filter with an active override
- **THEN** its tab label SHALL display no override-count badge

#### Scenario: Tab shows a restart marker
- **WHEN** a tab has at least one active override on a field flagged as requiring a restart
- **THEN** its tab label SHALL display a distinct restart-required marker in addition to its override count

#### Scenario: Mobile dropdown option carries the same override signal as text
- **WHEN** the "Intégrations" tab currently has 3 fields with an active override, at least one of which requires a restart, and the Configuration page is viewed on a mobile-width viewport
- **THEN** the dropdown option for "Intégrations" SHALL include both the override count and a restart-required marker in its text

## ADDED Requirements

### Requirement: System Tab Values Never Force Horizontal Scroll
The "Système" tab's disk usage entries SHALL wrap or break as needed to stay within their card's width, regardless of how long an individual mount path is, and SHALL NOT force the Configuration page to scroll horizontally.

#### Scenario: A long disk mount path wraps instead of overflowing
- **WHEN** a disk usage entry's path has no natural break point and is wider than its card
- **THEN** the frontend SHALL wrap or break that path within the card, and the Configuration page SHALL NOT gain a horizontal scrollbar
