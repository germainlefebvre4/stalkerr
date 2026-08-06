# frontend-filters-management Specification

## Purpose
TBD - Created by syncing spec changes for frontend-ui-refactoring-bright.

## Requirements

### Requirement: Filters List View
The frontend SHALL display a dedicated tab named "Filtres de Tri" that lists all active filter configurations in a grid of clean, visually elevated cards by querying `GET /api/v1/filters`.

#### Scenario: View current filter configurations
- **WHEN** the user selects the "Filtres de Tri" tab
- **THEN** the frontend SHALL fetch filter configurations from `/api/v1/filters` and render each config showing its name, target attribute, and inclusion/exclusion regex lists.

### Requirement: Create Filter Configuration
The frontend SHALL allow creating a new filter configuration via a modern popup dialog using Radix UI `Dialog` primitives, submitting a `POST` request to `/api/v1/filters` on confirmation. If the `POST` request fails, the backend response SHALL include a distinct machine-readable error code of `"filter_create_failed"` (rather than a generic error code) so the frontend can render a specific, translated error message in the dialog instead of a generic one.

#### Scenario: Successfully create a new inclusion filter
- **WHEN** the user opens the "Créer Filtre" dialog, inputs name, selects attribute, adds inclusion patterns, and clicks "Enregistrer"
- **THEN** the frontend SHALL submit a `POST` request to `/api/v1/filters`, display a success notification, close the dialog, and refresh the filters list.

#### Scenario: Filter creation fails
- **WHEN** the user submits the "Créer Filtre" dialog and the `POST /api/v1/filters` request fails (e.g. a duplicate name)
- **THEN** the backend SHALL respond with `ErrorResponse.error` set to `"filter_create_failed"`, and the frontend SHALL display the translated message for that code inline in the dialog instead of the raw backend `message` text

### Requirement: Delete Filter Configuration
The frontend SHALL allow deleting an existing filter configuration via a delete button on each filter card, which triggers a `DELETE` request to `/api/v1/filters/:id` on confirmation.

#### Scenario: Delete a filter configuration
- **WHEN** the user clicks the "Supprimer 🗑️" button on a filter card and confirms the action
- **THEN** the frontend SHALL submit a `DELETE` request to `/api/v1/filters/:id`, display a success notification, and remove the card from the list.
