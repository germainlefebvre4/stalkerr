## ADDED Requirements

### Requirement: Filters List View
The frontend SHALL display a dedicated tab named "Filtres de Tri" that lists all active filter configurations in a grid of clean, visually elevated cards by querying `GET /api/v1/filters`.

#### Scenario: View current filter configurations
- **WHEN** the user selects the "Filtres de Tri" tab
- **THEN** the frontend SHALL fetch filter configurations from `/api/v1/filters` and render each config showing its name, target attribute, and inclusion/exclusion regex lists.

### Requirement: Create Filter Configuration
The frontend SHALL allow creating a new filter configuration via a modern popup dialog using Radix UI `Dialog` primitives, submitting a `POST` request to `/api/v1/filters` on confirmation.

#### Scenario: Successfully create a new inclusion filter
- **WHEN** the user opens the "Créer Filtre" dialog, inputs name, selects attribute, adds inclusion patterns, and clicks "Enregistrer"
- **THEN** the frontend SHALL submit a `POST` request to `/api/v1/filters`, display a success notification, close the dialog, and refresh the filters list.

### Requirement: Delete Filter Configuration
The frontend SHALL allow deleting an existing filter configuration via a delete button on each filter card, which triggers a `DELETE` request to `/api/v1/filters/:id` on confirmation.

#### Scenario: Delete a filter configuration
- **WHEN** the user clicks the "Supprimer 🗑️" button on a filter card and confirms the action
- **THEN** the frontend SHALL submit a `DELETE` request to `/api/v1/filters/:id`, display a success notification, and remove the card from the list.
