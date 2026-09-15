# frontend-app-settings-management Specification

## Purpose

Provides the frontend UI, within the Configuration page, for viewing and editing the application's runtime-overridable settings (Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads, logging) and for managing M3U sources, showing each field's origin and, where applicable, that a restart is required for a change to take effect.

## Requirements

### Requirement: Settings Fields Show Origin
For every field backed by `app-settings`, the frontend SHALL display its current effective value alongside a compact visual indicator of its origin: one visual state when an override is active ("Interface"), a different one otherwise ("Config"). The indicator SHALL reveal the full origin label ("Interface" or "Config") on hover or keyboard focus.

#### Scenario: Field with no override
- **WHEN** the user views a settings field that has no stored override
- **THEN** the frontend SHALL display its current value with the "Config" indicator state, and hovering or focusing the indicator SHALL reveal the label "Config"

#### Scenario: Field with an active override
- **WHEN** the user views a settings field that has a stored override
- **THEN** the frontend SHALL display the overridden value with the "Interface" indicator state, and hovering or focusing the indicator SHALL reveal the label "Interface"

### Requirement: Editing and Clearing a Settings Field
Typing a new value into an overridable field SHALL stage it as a pending change for that field, without submitting it. The field's group SHALL display a save action once it has at least one pending change; confirming that action SHALL submit every pending field in the group as stored overrides, and cancelling it SHALL discard every pending change in the group, reverting each field to its last-saved effective value. Clearing an existing override on a field (reverting it to its "Config" value) SHALL remain an immediate, single-field action, independent of any other pending changes in the same group.

#### Scenario: Typing a new value stages a pending change
- **WHEN** the user types a new value into an overridable field without yet confirming the group's save action
- **THEN** the frontend SHALL mark that field as having a pending change and SHALL NOT submit it to the backend

#### Scenario: Setting an override
- **WHEN** the user confirms the save action for a group that has one or more fields with a pending change
- **THEN** the frontend SHALL submit each pending field's new value as a stored override and update each field's displayed origin indicator to "Interface" once every submission succeeds

#### Scenario: Discarding a group's pending changes
- **WHEN** the user cancels a group that has one or more fields with a pending change
- **THEN** the frontend SHALL revert every field in that group to its last-saved effective value and SHALL remove the group's save action

#### Scenario: Clearing an override
- **WHEN** the user activates a field's reset-to-config control on a field that currently has an active override, while other fields in the same group have unrelated pending changes
- **THEN** the frontend SHALL immediately remove that field's stored override and revert it to its "Config" value, without affecting the other fields' pending changes

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

### Requirement: M3U Sources Management
The frontend SHALL provide a "Sources M3U" section within the Configuration page's "Contenu" tab, listing the effective M3U sources (origin and runtime), and allowing the user to create a new source, edit an existing source (creating or replacing its runtime override), and delete a runtime-only source or a runtime override (reverting an overridden source to its origin definition). If the name submitted for a new source already identifies an existing runtime-only source, the frontend SHALL warn the user that the existing runtime source will be replaced before submitting, using the same trigger, dialog layout, and inline replace-warning presentation as the Filtres creation dialog (see `frontend-filters-management`). In the create/edit dialog, the source's enabled/disabled control SHALL appear before its other configuration fields (name excepted).

#### Scenario: Viewing the effective sources list
- **WHEN** the user views the "Contenu" tab's "Sources M3U" section
- **THEN** the frontend SHALL display every effective source with its name, file path, and download settings, distinguishing origin-only sources from runtime-overridden or runtime-only ones

#### Scenario: Auth password status shown only when set
- **WHEN** the user views a source card for a source that currently has no `download.auth_password` set
- **THEN** the frontend SHALL display no auth-password-related text for that source, rather than indicating that it is unset

#### Scenario: Auth password status shown when set
- **WHEN** the user views a source card for a source that currently has `download.auth_password` set
- **THEN** the frontend SHALL indicate that a password is set, without displaying its raw value

#### Scenario: Editing a source without changing its auth password
- **WHEN** the user edits a source's other fields without typing into the auth password field
- **THEN** the frontend SHALL omit `download.auth_password` from the submitted request, so the backend keeps the current effective password rather than clearing it

#### Scenario: Editing a source to clear its auth password
- **WHEN** the user explicitly clears the auth password field before submitting
- **THEN** the frontend SHALL submit `download.auth_password` as an explicit empty value, distinct from omitting it, so the backend clears the stored password

#### Scenario: Creating a new source
- **WHEN** the user submits a new source with a name that does not exist in the origin configuration or among existing runtime sources
- **THEN** the frontend SHALL create it as a runtime-only source and add it to the displayed list

#### Scenario: Editing an origin-defined source
- **WHEN** the user edits a source that currently comes from `config.yml`
- **THEN** the frontend SHALL create a runtime override for that source's name and mark it as coming from the interface

#### Scenario: Warning before replacing an existing runtime-only source
- **WHEN** the user submits a new source whose name matches an existing runtime-only source
- **THEN** the frontend SHALL warn, using the same inline dialog warning presentation as Filtres, that the existing runtime source will be replaced, and SHALL replace it in place (not create a second one) if the user confirms

#### Scenario: Deleting an override reverts to origin
- **WHEN** the user deletes a runtime override for a source that also has an origin definition
- **THEN** the frontend SHALL display that source's origin definition again

#### Scenario: Enabled control appears first in the dialog
- **WHEN** the user opens the create or edit dialog for an M3U source
- **THEN** the enabled/disabled control SHALL appear immediately after the name field and before the file path, URL, authentication, and tuning fields

### Requirement: Settings Field Layout
Overridable field groups (e.g. Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads tuning, Logging, M3U update interval) SHALL each render as a card within a content column capped at 960px, centered within its tab panel. A group of 6 or fewer fields SHALL render as a compact card with a minimum width of 280px, allowing multiple compact cards to share a row at sufficient width. A group of more than 6 fields SHALL always span the tab panel's full width. Within a card, boolean, numeric, and short-selection fields SHALL lay out on an internal multi-column grid; text, URL, and secret fields SHALL span the card's full width.

#### Scenario: A small group shares a row with another small group
- **WHEN** the "Intégrations" tab is displayed at a width sufficient for more than one 280px-minimum card, and Radarr (5 fields) and Sonarr (5 fields) are both displayed
- **THEN** the Radarr and Sonarr cards SHALL pack into the available row width instead of each spanning it alone

#### Scenario: A large group spans the full row
- **WHEN** the "Avancé" tab displays the Downloads tuning group (13 fields)
- **THEN** its card SHALL span the tab panel's full width regardless of available row width

#### Scenario: A boolean field does not span the full card width
- **WHEN** a card contains a boolean field alongside a numeric field
- **THEN** neither field SHALL span the card's full width; both SHALL lay out on the card's internal multi-column grid

### Requirement: Boolean Field Control
Boolean overridable fields SHALL render using a toggle switch control, not a text dropdown.

#### Scenario: Boolean field renders as a toggle
- **WHEN** the user views a boolean overridable field (e.g. Radarr's "enabled")
- **THEN** the frontend SHALL render it as a toggle switch, and toggling it SHALL stage the new value as a pending change per the Editing and Clearing a Settings Field requirement

### Requirement: Enumerated Field Control
Overridable fields whose value is constrained by the backend to a fixed set (e.g. the application and database logging levels) SHALL render as a `<select>` offering exactly that fixed set of technical values, each shown with a localized display label distinct from its technical value.

#### Scenario: Logging level renders as a labeled select
- **WHEN** the user views the "Logging" group's application or database log level field
- **THEN** the frontend SHALL render a `<select>` listing the four backend-accepted levels (`debug`, `info`, `warn`, `error`) as localized labels (e.g. "Debug", "Info", "Warning", "Erreur"), and selecting one SHALL stage its technical value as a pending change per the Editing and Clearing a Settings Field requirement

### Requirement: Test Connection Action for Radarr, Sonarr, and Jellyfin
For the Radarr, Sonarr, and Jellyfin field groups, the frontend SHALL provide a "Test connection" action, appearing in the group's action area whenever that group has an unsaved change, that submits the group's current in-progress field values (not necessarily saved) to the connectivity-test endpoint and displays the result as a status badge on the card, distinguishing OK from KO with its specific reason. The frontend SHALL NOT submit a value for a sensitive field the user has not explicitly retyped since it was masked; that field is sent empty.

#### Scenario: Testing an unsaved change
- **WHEN** the user has an unsaved change in the Radarr, Sonarr, or Jellyfin card and clicks "Tester la connexion"
- **THEN** the frontend submits the card's current in-form values, including any unsaved edits, to the connectivity-test endpoint and displays the returned OK/KO-with-reason result as a badge on the card

#### Scenario: Test action requires an unsaved change
- **WHEN** the Radarr, Sonarr, or Jellyfin card has no unsaved change
- **THEN** the frontend SHALL NOT show the "Test connection" action for that card

#### Scenario: Testing without retyping a sensitive field
- **WHEN** the user modifies the URL field but does not retype the API key field
- **THEN** the frontend submits the test with an empty API key value rather than the previously saved key

#### Scenario: Result distinguishes the failure reason
- **WHEN** the connectivity test returns KO with reason "unauthorized"
- **THEN** the badge SHALL display that specific reason, distinguishing it from unreachable, timeout, or other failure reasons

#### Scenario: Test action disabled without a URL
- **WHEN** the card's URL field is empty
- **THEN** the frontend SHALL disable the "Test connection" action rather than submitting a request
