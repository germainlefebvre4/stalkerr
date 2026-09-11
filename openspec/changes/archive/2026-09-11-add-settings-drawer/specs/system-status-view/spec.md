## MODIFIED Requirements

### Requirement: Header entry point on desktop and mobile
The app header SHALL NOT display a dedicated system status icon. System status SHALL be reachable via the "Système" section of the Settings drawer (opened from the header's single settings icon), identically on desktop and mobile viewports. This entry point SHALL NOT carry any badge or indicator reflecting health state; it is a plain navigation affordance.

#### Scenario: Icon visible on desktop
- **WHEN** the app is viewed on a desktop-width viewport
- **THEN** the header SHALL show the single settings icon, and system status SHALL be reachable by opening the Settings drawer and expanding its "Système" section, with no separate status icon present in the header

#### Scenario: Icon visible on mobile
- **WHEN** the app is viewed on a mobile-width viewport
- **THEN** the same "Système" section SHALL be reachable through the same Settings drawer entry point, with no separate status icon shown

#### Scenario: Icon does not reflect health state
- **WHEN** any dependency is currently down, unknown, or has never been checked this session
- **THEN** neither the settings icon nor the "Système" section header SHALL change appearance (no badge, color, or count tied to health state)

### Requirement: Status dialog fetches on demand only
Expanding the Settings drawer's "Système" section SHALL fetch the aggregated system status. No system status request SHALL be made before the section has been expanded at least once, and no automatic background polling SHALL occur while the section is collapsed or expanded.

#### Scenario: Opening the dialog triggers a fetch
- **WHEN** the user expands the "Système" section for the first time in a session
- **THEN** the section SHALL issue exactly one request for the aggregated system status

#### Scenario: No fetch before first open
- **WHEN** the app loads and the user has not yet expanded the "Système" section
- **THEN** no system status request SHALL have been made

#### Scenario: Reopening the dialog fetches fresh data
- **WHEN** the user collapses the "Système" section and expands it again
- **THEN** the section SHALL issue a new request rather than reusing the previously displayed data
