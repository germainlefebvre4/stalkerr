# frontend-bandwidth-schedule-management Specification

## Purpose

Provides the Configuration-page UI for defining the weekly bandwidth schedule and configuring the Jellyfin-based throttle, so a user can adapt download behavior without editing configuration files or calling the API directly.

## Requirements

### Requirement: Bandwidth Schedule Section Placement
The frontend SHALL provide a bandwidth-schedule section within the Configuration page's "Avancé" tab, grouping the weekly schedule editor, the Jellyfin-throttle settings, and the shared throttle rate in one place.

#### Scenario: Locating the section
- **WHEN** the user opens the Configuration page's "Avancé" tab
- **THEN** the frontend SHALL display a bandwidth-schedule section alongside the existing Downloads tuning settings

### Requirement: Weekly Schedule Window Editor
The frontend SHALL let the user view, create, edit, and delete individual weekly schedule windows, each specifying one or more days of the week, a start time, an end time, and an action (`none`, `throttle`, or `stop`), without requiring the full schedule to be resubmitted for a single window's change.

#### Scenario: Adding a window
- **WHEN** the user creates a new window for Monday-Friday, 08:00-18:00, action `throttle`
- **THEN** the frontend SHALL submit it and display it alongside any previously defined windows

#### Scenario: Deleting a window
- **WHEN** the user deletes one window
- **THEN** the frontend SHALL remove only that window, leaving every other window unchanged

### Requirement: Overnight Window Input
The editor SHALL accept an end time earlier than the start time for a window without treating it as a validation error, consistent with the backend's overnight-window support.

#### Scenario: Entering an overnight window
- **WHEN** the user enters start time 22:00 and end time 07:00 for a window
- **THEN** the frontend SHALL accept and submit the window without an error, and SHALL display it in a way that makes clear it spans past midnight

### Requirement: Jellyfin Throttle Settings
The frontend SHALL provide, within the same section, a toggle to enable or disable Jellyfin active-playback detection and, when enabled, a choice between `throttle` and `stop` as the action applied while playback is active.

#### Scenario: Enabling Jellyfin-based throttling
- **WHEN** the user enables Jellyfin active-playback detection and selects action `stop`
- **THEN** the frontend SHALL save both the enabled state and the selected action

### Requirement: Shared Throttle Rate Field
The frontend SHALL provide a single throttle-rate field, used by the effective policy whenever it is `throttle`, regardless of whether the weekly schedule, Jellyfin, or both contributed that `throttle` action.

#### Scenario: Editing the shared rate
- **WHEN** the user changes the throttle rate value
- **THEN** the new rate SHALL apply the next time the effective policy is `throttle`, whichever signal triggers it

### Requirement: Effective Policy Preview
The frontend SHALL display the currently effective download policy (`none`, `throttle`, or `stop`) and, when it is not `none`, which signal(s) are currently contributing to it, refreshing this display without requiring a page reload.

#### Scenario: Viewing the current policy
- **WHEN** the user opens the bandwidth-schedule section while a schedule window is active
- **THEN** the frontend SHALL show the effective policy as `throttle` (or `stop`) and indicate that the schedule is the contributing signal
