## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Enumerated Field Control
Overridable fields whose value is constrained by the backend to a fixed set (e.g. the application and database logging levels) SHALL render as a `<select>` offering exactly that fixed set of technical values, each shown with a localized display label distinct from its technical value.

#### Scenario: Logging level renders as a labeled select
- **WHEN** the user views the "Logging" group's application or database log level field
- **THEN** the frontend SHALL render a `<select>` listing the four backend-accepted levels (`debug`, `info`, `warn`, `error`) as localized labels (e.g. "Debug", "Info", "Warning", "Erreur"), and selecting one SHALL stage its technical value as a pending change per the Editing and Clearing a Settings Field requirement

## REMOVED Requirements

### Requirement: Bootstrap Configuration Display
**Reason**: The bootstrap connection variables (database host/port/user/password/dbname/sslmode, API port, metrics enabled/port/path) are internal deployment details with little value to an end user browsing the Configuration page. The "Système" tab keeps its service health, disk usage, and build-version information; only this read-only bootstrap card is removed.
**Migration**: None. These values remain visible through the deployment's own configuration (Helm values, environment variables, `config.yml`) for operators who need them; no data or capability is lost, only its display in the Configuration page.
