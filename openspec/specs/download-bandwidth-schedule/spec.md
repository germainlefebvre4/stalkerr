# download-bandwidth-schedule Specification

## Purpose

Lets the user define a weekly schedule of time windows, each tagged `none`, `throttle`, or `stop`, and resolves which of those actions is currently active so `adaptive-download-throttling` can enforce it.

## Requirements

### Requirement: Weekly Schedule Window Definition
The system SHALL allow defining zero or more bandwidth schedule windows, each specifying one or more days of the week, a start time and an end time (in the server's local time), and an action of `none`, `throttle`, or `stop`.

#### Scenario: Defining a throttle window
- **WHEN** a window is created for Monday-Friday, 08:00-18:00, action `throttle`
- **THEN** the schedule stores it and it becomes eligible to be the active window during that time range on those days

### Requirement: Overnight Windows
A window's end time MAY be earlier than its start time, in which case the window SHALL be treated as spanning from its start time on its configured day(s) through its end time on the following day.

#### Scenario: Window crossing midnight
- **WHEN** a window is defined for day Friday, start 22:00, end 07:00, action `stop`
- **THEN** the window SHALL be active from Friday 22:00 through Saturday 07:00

### Requirement: Resolving the Active Window
At any instant, the system SHALL determine the currently active schedule action by evaluating all defined windows against the current day and time (server local time) and, when more than one window is simultaneously active, selecting the most restrictive action among them, using the ordering `stop` > `throttle` > `none`.

#### Scenario: No window active
- **WHEN** the current day/time falls outside every defined window
- **THEN** the active action SHALL be `none`

#### Scenario: Two overlapping windows
- **WHEN** a `throttle` window and a `stop` window are both active at the current instant
- **THEN** the active action SHALL be `stop`

### Requirement: Default Schedule Is Unrestricted
When no schedule window has been defined, the system SHALL report the active action as `none` at all times, so a system that has never configured a schedule sees no change in existing download behavior.

#### Scenario: No windows configured
- **WHEN** no schedule window has ever been created
- **THEN** the active action SHALL always resolve to `none`

### Requirement: Managing Schedule Windows Independently
The system SHALL allow creating, updating, and deleting individual schedule windows independently, without requiring the full weekly schedule to be replaced for a single window's change.

#### Scenario: Deleting one window leaves others intact
- **WHEN** one schedule window is deleted
- **THEN** every other previously defined window SHALL remain unchanged and continue to be evaluated
