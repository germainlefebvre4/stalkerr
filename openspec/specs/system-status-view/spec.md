# system-status-view Specification

## Purpose

Give users a quick, on-demand diagnostic view of the database, Radarr, Sonarr, TMDB, and disk health, reachable from the app header, without having to read logs.

## Requirements

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

### Requirement: Dialog displays three-state rows with a short reason on failure
The dialog SHALL display one row each for the database, Radarr, Sonarr, and TMDB, each visually distinguishing the **OK**, **KO**, and **not configured** states. A row in the **KO** state SHALL display the short reason reported by the API.

#### Scenario: States are visually distinguishable
- **WHEN** the dialog displays its rows
- **THEN** OK, KO, and not-configured states SHALL be visually distinguishable from one another (e.g. by color and/or icon)

#### Scenario: KO row shows its reason without needing logs
- **WHEN** a row is in the KO state
- **THEN** the row SHALL display the short human-readable reason from the API response

### Requirement: Dialog displays deduplicated disk usage
The dialog SHALL display one entry per distinct mounted volume reported by the API, showing at minimum used/available space for that volume.

#### Scenario: Merged volume shown once
- **WHEN** the API reports a single disk usage entry backing multiple configured paths
- **THEN** the dialog SHALL display that entry once, not once per path

### Requirement: Manual refresh action within the dialog
The dialog SHALL provide a refresh action that re-fetches the aggregated system status and updates the displayed rows without closing the dialog.

#### Scenario: Refresh updates displayed state
- **WHEN** the user triggers the refresh action while the dialog is open
- **THEN** the dialog SHALL issue a new request and update its rows to reflect the response

### Requirement: Dialog displays build version metadata
The dialog SHALL display the running instance's version, commit, and build date as reported by the API.

#### Scenario: Version metadata visible in dialog
- **WHEN** the dialog displays a successful status response
- **THEN** the version, commit, and build date SHALL be visible within the dialog
