## MODIFIED Requirements

### Requirement: Header entry point on desktop and mobile
The app header SHALL NOT display a dedicated system status icon. System status SHALL be reachable via the Configuration page's "Système" tab (opened from the header's single settings icon), identically on desktop and mobile viewports. This entry point SHALL NOT carry any badge or indicator reflecting health state; it is a plain navigation affordance.

#### Scenario: Icon visible on desktop
- **WHEN** the app is viewed on a desktop-width viewport
- **THEN** the header SHALL show the single settings icon, and system status SHALL be reachable by opening the Configuration page and selecting its "Système" tab, with no separate status icon present in the header

#### Scenario: Icon visible on mobile
- **WHEN** the app is viewed on a mobile-width viewport
- **THEN** the same "Système" tab SHALL be reachable through the same Configuration page entry point, with no separate status icon shown

#### Scenario: Icon does not reflect health state
- **WHEN** any dependency is currently down, unknown, or has never been checked this session
- **THEN** neither the settings icon nor the "Système" tab label SHALL change appearance (no badge, color, or count tied to health state)

### Requirement: Status dialog fetches on demand only
Activating the Configuration page's "Système" tab SHALL fetch the aggregated system status. No system status request SHALL be made before that tab has been activated at least once, and no automatic background polling SHALL occur while it is inactive or active.

#### Scenario: Opening the dialog triggers a fetch
- **WHEN** the user activates the "Système" tab for the first time in a session
- **THEN** the tab SHALL issue exactly one request for the aggregated system status

#### Scenario: No fetch before first open
- **WHEN** the app loads and the user has not yet activated the "Système" tab
- **THEN** no system status request SHALL have been made

#### Scenario: Reopening the dialog fetches fresh data
- **WHEN** the user switches away from the "Système" tab to another Configuration-page tab and then back
- **THEN** the tab SHALL issue a new request rather than reusing the previously displayed data
