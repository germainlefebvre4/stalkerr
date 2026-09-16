# frontend-filters-management Specification

## Purpose
Provides the frontend UI for managing runtime filter overrides — viewing active overrides alongside the origin `config.yml` patterns, creating new overrides, and deleting them — for each filterable attribute (Group Title, TVG Name).

## Requirements

### Requirement: Filters List View
The frontend SHALL provide a "Filtres" section within the Configuration page's "Contenu" tab that groups filter configuration by target attribute (Group Title, TVG Name). Viewing the section SHALL query `GET /api/v1/filters` for the active runtime override and the origin configuration endpoint for the `config.yml`-defined patterns, and SHALL render both: the origin patterns (read-only, labeled as origin/system) and the active runtime override, if any (labeled as an active override), so the user can see at a glance which configuration is currently in effect for that attribute.

#### Scenario: View current filter configurations
- **WHEN** the user views the "Contenu" tab's "Filtres" section
- **THEN** the frontend SHALL fetch both the origin configuration and the active runtime overrides, and render one section per attribute (Group Title, TVG Name) showing the origin include/exclude patterns and, if present, the active override's name and include/exclude patterns

#### Scenario: Attribute with no active override
- **WHEN** an attribute has no active runtime override
- **THEN** the frontend SHALL display only the origin `config.yml` patterns for that attribute, with no override section

#### Scenario: Attribute with an active override
- **WHEN** an attribute has an active runtime override
- **THEN** the frontend SHALL visually distinguish the origin patterns (labeled as origin/system) from the active override (labeled as an active override), making clear that the override is what is currently applied

### Requirement: Create Filter Configuration
The frontend SHALL allow creating a new filter configuration via a popup dialog using Radix UI `Dialog` primitives, opened from a trigger in the "Filtres" section's own header, submitting a `POST` request to `/api/v1/filters` on confirmation. If the `POST` request fails, the backend response SHALL include a distinct machine-readable error code of `"filter_create_failed"` (rather than a generic error code) so the frontend can render a specific, translated error message in the dialog instead of a generic one. The dialog SHALL provide a way to load the currently active configuration (the active override if one exists, otherwise the origin configuration) for the selected attribute into the Include/Exclude fields, and SHALL warn the user, via an inline warning banner within the dialog, before submission if an active override for the selected attribute will be replaced. This trigger, dialog layout, and inline replace-warning presentation SHALL be the same shared pattern used by the Sources M3U creation dialog (see `frontend-app-settings-management`).

#### Scenario: Successfully create a new inclusion filter
- **WHEN** the user opens the "Filtres" section's create dialog, inputs name, selects attribute, adds inclusion patterns, and clicks "Enregistrer"
- **THEN** the frontend SHALL submit a `POST` request to `/api/v1/filters`, display a success notification, close the dialog, and refresh the filters list

#### Scenario: Filter creation fails
- **WHEN** the user submits the create dialog and the `POST /api/v1/filters` request fails (e.g. a duplicate name)
- **THEN** the backend SHALL respond with `ErrorResponse.error` set to `"filter_create_failed"`, and the frontend SHALL display the translated message for that code inline in the dialog instead of the raw backend `message` text

#### Scenario: Load the currently active configuration into the form
- **WHEN** the user selects an attribute and clicks "Reprendre la config actuelle"
- **THEN** the frontend SHALL populate the Include and Exclude fields with the active runtime override's patterns for that attribute if one exists, or with the origin `config.yml` patterns for that attribute otherwise, leaving the fields editable

#### Scenario: Warning before replacing an existing active override
- **WHEN** the user selects an attribute that already has an active runtime override and submits the create dialog
- **THEN** the frontend SHALL display an inline warning banner within the dialog that the existing active override for that attribute will be replaced before the request is submitted

### Requirement: Delete Filter Configuration
The frontend SHALL allow deleting an existing filter configuration via a delete button on each filter card, which triggers a `DELETE` request to `/api/v1/filters/:id` on confirmation.

#### Scenario: Delete a filter configuration
- **WHEN** the user clicks the "Supprimer 🗑️" button on a filter card and confirms the action
- **THEN** the frontend SHALL submit a `DELETE` request to `/api/v1/filters/:id`, display a success notification, and remove the card from the list.

### Requirement: Dry-Run Filter Testing
The frontend SHALL allow the user to dry-run test a filter's include/exclude patterns against a chosen M3U source's real playlist content, from two places: the create dialog (testing the patterns currently typed into the form, before saving) and each filter card in the "Filtres" section (testing the patterns already in effect — origin or active override — for that card, read-only). Both SHALL submit to the dry-run endpoint and render its aggregate summary (total lines scanned, matched count, excluded count, top matched and excluded values) in the same place the test was triggered from.

#### Scenario: Testing patterns while creating a filter
- **WHEN** the user has typed include/exclude patterns into the create dialog, selected a source, and clicks "Tester"
- **THEN** the frontend SHALL submit those exact in-progress values (not yet saved) to the dry-run endpoint and display the aggregate summary within the dialog, without submitting the create request

#### Scenario: Testing an already-saved filter from its card
- **WHEN** the user clicks "Tester" on a filter card (origin or active override) in the "Filtres" section
- **THEN** the frontend SHALL submit that card's existing include/exclude patterns and attribute to the dry-run endpoint, with a source selected by the user, and display the aggregate summary inline on the card without opening the create dialog

#### Scenario: Source has no archive yet
- **WHEN** the dry-run endpoint reports that the selected source has no downloaded archive available
- **THEN** the frontend SHALL display a message explaining that no data is available for that source yet, and SHALL NOT trigger a download

#### Scenario: Invalid pattern prevents testing
- **WHEN** the dry-run endpoint rejects the request because an include or exclude pattern is not a valid regular expression
- **THEN** the frontend SHALL display an inline error identifying the invalid pattern instead of a summary

### Requirement: Dry-Run Content Search
Once a dry-run's aggregate summary is displayed, the frontend SHALL offer a search field scoped to the tested attribute that lets the user verify specific content: submitting a search term SHALL query the dry-run endpoint's content search mode and display each matching line together with whether it would match or be excluded under the tested patterns.

#### Scenario: Verifying a specific line was excluded
- **WHEN** the user enters a channel or group name in the dry-run search field after viewing a summary
- **THEN** the frontend SHALL display every archive line whose value for the tested attribute contains that text, each labeled as would-match or would-be-excluded, so the user can confirm the expected lines disappeared or remained

#### Scenario: Search available only after a summary exists
- **WHEN** the user has not yet run a dry-run test
- **THEN** the frontend SHALL NOT display the content search field
