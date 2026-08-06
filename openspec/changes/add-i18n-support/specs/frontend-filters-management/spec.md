## MODIFIED Requirements

### Requirement: Create Filter Configuration
The frontend SHALL allow creating a new filter configuration via a modern popup dialog using Radix UI `Dialog` primitives, submitting a `POST` request to `/api/v1/filters` on confirmation. If the `POST` request fails, the backend response SHALL include a distinct machine-readable error code of `"filter_create_failed"` (rather than a generic error code) so the frontend can render a specific, translated error message in the dialog instead of a generic one.

#### Scenario: Successfully create a new inclusion filter
- **WHEN** the user opens the "Créer Filtre" dialog, inputs name, selects attribute, adds inclusion patterns, and clicks "Enregistrer"
- **THEN** the frontend SHALL submit a `POST` request to `/api/v1/filters`, display a success notification, close the dialog, and refresh the filters list.

#### Scenario: Filter creation fails
- **WHEN** the user submits the "Créer Filtre" dialog and the `POST /api/v1/filters` request fails (e.g. a duplicate name)
- **THEN** the backend SHALL respond with `ErrorResponse.error` set to `"filter_create_failed"`, and the frontend SHALL display the translated message for that code inline in the dialog instead of the raw backend `message` text
