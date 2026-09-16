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
The frontend SHALL allow the user to dry-run test a single attribute's include/exclude patterns against a chosen M3U source's real playlist content, from two places: the create dialog (testing the patterns currently typed into the form, before saving) and each filter card in the "Filtres" section (testing the patterns already in effect — origin or active override — for that card, read-only). Both entry points SHALL open one single, shared test drawer that owns the M3U source selector and the "Tester" trigger, submit to the dry-run endpoint, and render its aggregate summary inside that drawer: total lines scanned, matched count, excluded count, and the top matched and top excluded values rendered as two side-by-side columns. The drawer SHALL display a contextual title identifying what is being tested (the attribute, and whether it is the create-dialog's in-progress patterns, a card's origin patterns, or a card's active override). The "Filtres" section SHALL use exactly one shared drawer instance for all of its cards rather than one per card. The create dialog's drawer SHALL open as a non-modal sibling of the dialog, so the in-progress form remains visible and editable while results are shown.

#### Scenario: Testing patterns while creating a filter
- **WHEN** the user has typed include/exclude patterns into the create dialog, selected a source in the test drawer, and clicks "Tester"
- **THEN** the frontend SHALL submit those exact in-progress values (not yet saved) to the dry-run endpoint and display the aggregate summary in the test drawer, alongside the still-open and still-editable create dialog, without submitting the create request

#### Scenario: Testing an already-saved filter from its card
- **WHEN** the user clicks "Tester" on a filter card (origin or active override) in the "Filtres" section
- **THEN** the frontend SHALL open the shared test drawer labeled for that card's attribute and origin/override status, let the user select a source inside the drawer, submit that card's existing include/exclude patterns and attribute to the dry-run endpoint, and display the aggregate summary in the drawer without opening the create dialog

#### Scenario: Testing a different card reuses the same drawer
- **WHEN** the user has the shared test drawer open for one filter card and clicks "Tester" on a different card
- **THEN** the frontend SHALL update the same open drawer to reflect the newly selected card's attribute and patterns, rather than opening a second drawer

#### Scenario: Source selection lives in the drawer
- **WHEN** the user opens the test drawer from any entry point
- **THEN** the frontend SHALL present the M3U source selector and the "Tester" trigger inside the drawer itself, not on the triggering card or in the create dialog's own fields

#### Scenario: Source has no archive yet
- **WHEN** the dry-run endpoint reports that the selected source has no downloaded archive available
- **THEN** the frontend SHALL display, inside the test drawer, a message explaining that no data is available for that source yet, and SHALL NOT trigger a download

#### Scenario: Invalid pattern prevents testing
- **WHEN** the dry-run endpoint rejects the request because an include or exclude pattern is not a valid regular expression
- **THEN** the frontend SHALL display, inside the test drawer, an inline error identifying the invalid pattern instead of a summary

### Requirement: Dry-Run Content Search
Once a single-attribute dry-run's aggregate summary is displayed in the test drawer, the frontend SHALL offer a search field scoped to the tested attribute that lets the user verify specific content: submitting a search term SHALL query the dry-run endpoint's content search mode and display each matching line, together with whether it would match or be excluded under the tested patterns, as a table with a fixed maximum height and its own internal vertical scrollbar, so that a search returning many lines does not grow the height of the enclosing drawer.

#### Scenario: Verifying a specific line was excluded
- **WHEN** the user enters a channel or group name in the dry-run search field after viewing a summary
- **THEN** the frontend SHALL display every archive line whose value for the tested attribute contains that text, each labeled as would-match or would-be-excluded, so the user can confirm the expected lines disappeared or remained

#### Scenario: Search available only after a summary exists
- **WHEN** the user has not yet run a dry-run test
- **THEN** the frontend SHALL NOT display the content search field

#### Scenario: Many search results stay within a bounded, scrollable table
- **WHEN** a content search returns enough matching lines that listing them all would exceed the results table's maximum height
- **THEN** the frontend SHALL keep the table at its maximum height and let the user scroll within the table to see the remaining lines, instead of growing the enclosing drawer

### Requirement: Combined Dry-Run Testing
The "Filtres" section SHALL provide a "Tester l'ensemble" entry point that dry-run tests the currently effective patterns — the active runtime override if one exists, otherwise the origin `config.yml` patterns — for both Group Title and TVG Name together, against a chosen M3U source, opening the same shared test drawer used for single-attribute testing. The combined summary SHALL report the total lines scanned and, distinctly: how many lines are kept (pass both attributes' patterns), how many are excluded by the Group Title patterns only, how many by the TVG Name patterns only, and how many by both. While a combined summary is displayed, the drawer's content-search field SHALL let the user choose which of the two attributes (Group Title or TVG Name) the search term is matched against, and SHALL label each returned line with its combined verdict (kept, or excluded — and by which attribute's patterns).

#### Scenario: Testing the combined active configuration
- **WHEN** the user clicks "Tester l'ensemble" in the "Filtres" section and selects a source in the test drawer
- **THEN** the frontend SHALL submit the active override's patterns (or the origin `config.yml` patterns, if no override exists) for both Group Title and TVG Name to the dry-run endpoint in a single combined request, and display the cause-ventilated summary in the shared test drawer

#### Scenario: Attributing why lines were excluded
- **WHEN** a combined summary is displayed
- **THEN** the frontend SHALL show, as distinct figures, the counts of lines kept, excluded by Group Title only, excluded by TVG Name only, and excluded by both

#### Scenario: Choosing which attribute to search in combined mode
- **WHEN** the user has a combined summary displayed and wants to verify a specific line
- **THEN** the frontend SHALL let the user pick whether the search term is matched against each archive line's Group Title value or its TVG Name value, and SHALL display each matching line's combined verdict accordingly

#### Scenario: No override for one or both attributes
- **WHEN** the user tests the combined configuration and one or both attributes have no active runtime override
- **THEN** the frontend SHALL fall back to that attribute's origin `config.yml` patterns (or treat it as matching everything, if the origin also has none) without requiring the user to configure anything first
