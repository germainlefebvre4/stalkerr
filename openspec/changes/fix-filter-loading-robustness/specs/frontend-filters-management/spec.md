## MODIFIED Requirements

### Requirement: Create Filter Configuration
The frontend SHALL allow creating a new filter configuration via a modern popup dialog using Radix UI `Dialog` primitives, submitting a `POST` request to `/api/v1/filters` on confirmation. If the `POST` request fails, the backend response SHALL include a distinct machine-readable error code — `"filter_create_failed"` for a generic failure (e.g. a duplicate name), or `"invalid_pattern"` when a submitted include/exclude pattern does not compile as a valid regular expression — so the frontend can render a specific, translated error message in the dialog instead of a generic one.

#### Scenario: Successfully create a new inclusion filter
- **WHEN** the user opens the "Créer Filtre" dialog, inputs name, selects attribute, adds inclusion patterns, and clicks "Enregistrer"
- **THEN** the frontend SHALL submit a `POST` request to `/api/v1/filters`, display a success notification, close the dialog, and refresh the filters list

#### Scenario: Filter creation fails
- **WHEN** the user submits the "Créer Filtre" dialog and the `POST /api/v1/filters` request fails (e.g. a duplicate name)
- **THEN** the backend SHALL respond with `ErrorResponse.error` set to `"filter_create_failed"`, and the frontend SHALL display the translated message for that code inline in the dialog instead of the raw backend `message` text

#### Scenario: Filter creation fails due to an invalid pattern
- **WHEN** the user submits the "Créer Filtre" dialog with an include or exclude pattern that does not compile as a valid regular expression
- **THEN** the backend SHALL respond with `ErrorResponse.error` set to `"invalid_pattern"`, and the frontend SHALL display the translated message for that code inline in the dialog instead of closing or refreshing the list
