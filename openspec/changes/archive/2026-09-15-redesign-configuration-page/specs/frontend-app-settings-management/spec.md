## MODIFIED Requirements

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

### Requirement: Bootstrap Configuration Display
The frontend SHALL display the bootstrap configuration (database connection, API port, metrics port/enabled) as read-only, with its "Config" origin, within the Configuration page's "Système" tab, and SHALL NOT provide an edit control for these fields.

#### Scenario: Viewing bootstrap configuration
- **WHEN** the user views the "Système" tab
- **THEN** the frontend SHALL display the bootstrap configuration's current values with the "Config" origin indicator and no edit control

### Requirement: M3U Sources Management
The frontend SHALL provide a "Sources M3U" section within the Configuration page's "Contenu" tab, listing the effective M3U sources (origin and runtime), and allowing the user to create a new source, edit an existing source (creating or replacing its runtime override), and delete a runtime-only source or a runtime override (reverting an overridden source to its origin definition). If the name submitted for a new source already identifies an existing runtime-only source, the frontend SHALL warn the user that the existing runtime source will be replaced before submitting, using the same trigger, dialog layout, and inline replace-warning presentation as the Filtres creation dialog (see `frontend-filters-management`).

#### Scenario: Viewing the effective sources list
- **WHEN** the user views the "Contenu" tab's "Sources M3U" section
- **THEN** the frontend SHALL display every effective source with its name, file path, and download settings, distinguishing origin-only sources from runtime-overridden or runtime-only ones, and SHALL display only whether `download.auth_password` is set, never its raw value

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

## ADDED Requirements

### Requirement: Settings Field Layout
Overridable field groups (e.g. Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads tuning, Logging, M3U update interval, the bootstrap group) SHALL each render as a card within a content column capped at 960px, centered within its tab panel. A group of 6 or fewer fields SHALL render as a compact card with a minimum width of 280px, allowing multiple compact cards to share a row at sufficient width. A group of more than 6 fields SHALL always span the tab panel's full width. Within a card, boolean, numeric, and short-selection fields SHALL lay out on an internal multi-column grid; text, URL, and secret fields SHALL span the card's full width.

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
