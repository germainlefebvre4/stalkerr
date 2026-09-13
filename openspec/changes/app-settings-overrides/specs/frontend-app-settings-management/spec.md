## Purpose

Provides the frontend UI, within the Settings drawer, for viewing and editing the application's runtime-overridable settings (Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads, logging) and for managing M3U sources, showing each field's origin and, where applicable, that a restart is required for a change to take effect.

## ADDED Requirements

### Requirement: Settings Fields Show Origin
For every field backed by `app-settings`, the frontend SHALL display its current effective value alongside a badge indicating its origin: "Interface" when an override is active, or "Config" otherwise.

#### Scenario: Field with no override
- **WHEN** the user views a settings field that has no stored override
- **THEN** the frontend SHALL display its current value with a "Config" origin badge

#### Scenario: Field with an active override
- **WHEN** the user views a settings field that has a stored override
- **THEN** the frontend SHALL display the overridden value with an "Interface" origin badge

### Requirement: Editing and Clearing a Settings Field
The frontend SHALL allow the user to set a new value for any overridable field, submitting it as a stored override, and to clear an existing override, reverting the field to its "Config" value.

#### Scenario: Setting an override
- **WHEN** the user enters a new value for a settings field and confirms
- **THEN** the frontend SHALL store it as an override for that field and update the displayed origin badge to "Interface"

#### Scenario: Clearing an override
- **WHEN** the user clears the override on a field that currently has one
- **THEN** the frontend SHALL remove the stored override and update the displayed value and badge to the field's "Config" value

### Requirement: Sensitive Fields Are Masked in the Form
For fields the backend reports as sensitive (API keys, auth tokens, passwords), the frontend SHALL render an input that shows only whether a value is currently set, never the value itself, and SHALL only submit a new value when the user explicitly types one.

#### Scenario: Viewing a sensitive field with a value set
- **WHEN** the user views a sensitive field that currently has a value (override or config)
- **THEN** the frontend SHALL show that a value is set without displaying it

#### Scenario: Leaving a sensitive field unchanged
- **WHEN** the user edits other fields in the same form without typing into a sensitive field
- **THEN** the frontend SHALL NOT submit any change to that sensitive field's stored value

### Requirement: Restart-Required Indicator
For settings fields the backend reports as not taking effect until the process restarts, the frontend SHALL display a "restart required" indicator alongside the field after an override is set or cleared.

#### Scenario: Changing a restart-required field
- **WHEN** the user overrides a field flagged as requiring a restart to take effect
- **THEN** the frontend SHALL display a "restart required" indicator for that field after the change is saved

### Requirement: Bootstrap Configuration Display
The frontend SHALL display the bootstrap configuration (database connection, API port, metrics port/enabled) as read-only, with its "Config" origin, and SHALL NOT provide an edit control for these fields.

#### Scenario: Viewing bootstrap configuration
- **WHEN** the user views the bootstrap configuration section
- **THEN** the frontend SHALL display its current values with a "Config" badge and no edit control

### Requirement: M3U Sources Management
The frontend SHALL provide a "Sources M3U" section in the Settings drawer, collapsed by default, listing the effective M3U sources (origin and runtime), and allowing the user to create a new source, edit an existing source (creating or replacing its runtime override), and delete a runtime-only source or a runtime override (reverting an overridden source to its origin definition).

#### Scenario: Viewing the effective sources list
- **WHEN** the user expands the "Sources M3U" section
- **THEN** the frontend SHALL display every effective source with its name, file path, and download settings, distinguishing origin-only sources from runtime-overridden or runtime-only ones

#### Scenario: Creating a new source
- **WHEN** the user submits a new source with a name that does not exist in the origin configuration
- **THEN** the frontend SHALL create it as a runtime-only source and add it to the displayed list

#### Scenario: Editing an origin-defined source
- **WHEN** the user edits a source that currently comes from `config.yml`
- **THEN** the frontend SHALL create a runtime override for that source's name and mark it as coming from the interface

#### Scenario: Deleting an override reverts to origin
- **WHEN** the user deletes a runtime override for a source that also has an origin definition
- **THEN** the frontend SHALL display that source's origin definition again
