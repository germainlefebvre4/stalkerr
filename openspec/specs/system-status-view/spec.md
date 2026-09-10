# system-status-view Specification

## Purpose

Give users a quick, on-demand diagnostic view of the database, Radarr, Sonarr, TMDB, and disk health, reachable from the app header, without having to read logs.

## Requirements

### Requirement: Header entry point on desktop and mobile
The app header SHALL display a system status icon next to the language selector, identically on desktop and mobile viewports. The icon SHALL NOT carry any badge or indicator reflecting health state; it is a plain navigation affordance.

#### Scenario: Icon visible on desktop
- **WHEN** the app is viewed on a desktop-width viewport
- **THEN** the system status icon SHALL be visible in the header next to the language selector

#### Scenario: Icon visible on mobile
- **WHEN** the app is viewed on a mobile-width viewport
- **THEN** the same system status icon SHALL be visible in the header next to the language selector

#### Scenario: Icon does not reflect health state
- **WHEN** any dependency is currently down, unknown, or has never been checked this session
- **THEN** the icon's appearance SHALL remain unchanged (no badge, color, or count tied to health state)

### Requirement: Status dialog fetches on demand only
Clicking the header icon SHALL open a dialog that fetches the aggregated system status. No system status request SHALL be made before the dialog has been opened at least once, and no automatic background polling SHALL occur while the dialog is closed or open.

#### Scenario: Opening the dialog triggers a fetch
- **WHEN** the user clicks the system status icon
- **THEN** the dialog SHALL open and issue exactly one request for the aggregated system status

#### Scenario: No fetch before first open
- **WHEN** the app loads and the user has not yet clicked the system status icon
- **THEN** no system status request SHALL have been made

#### Scenario: Reopening the dialog fetches fresh data
- **WHEN** the user closes the dialog and reopens it
- **THEN** the dialog SHALL issue a new request rather than reusing the previously displayed data

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
