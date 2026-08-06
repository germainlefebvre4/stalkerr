## MODIFIED Requirements

### Requirement: Filters List View
The frontend SHALL display a dedicated tab named "Filtres de Tri" that groups filter configuration by target attribute (Group Title, TVG Name). For each attribute, the frontend SHALL query `GET /api/v1/filters` for the active runtime override and the origin configuration endpoint for the `config.yml`-defined patterns, and SHALL render both: the origin patterns (read-only, labeled as origin/system) and the active runtime override, if any (labeled as an active override), so the user can see at a glance which configuration is currently in effect for that attribute.

#### Scenario: View current filter configurations
- **WHEN** the user selects the "Filtres de Tri" tab
- **THEN** the frontend SHALL fetch both the origin configuration and the active runtime overrides, and render one section per attribute (Group Title, TVG Name) showing the origin include/exclude patterns and, if present, the active override's name and include/exclude patterns

#### Scenario: Attribute with no active override
- **WHEN** an attribute has no active runtime override
- **THEN** the frontend SHALL display only the origin `config.yml` patterns for that attribute, with no override section

#### Scenario: Attribute with an active override
- **WHEN** an attribute has an active runtime override
- **THEN** the frontend SHALL visually distinguish the origin patterns (labeled as origin/system) from the active override (labeled as an active override), making clear that the override is what is currently applied

### Requirement: Create Filter Configuration
The frontend SHALL allow creating a new filter configuration via a modern popup dialog using Radix UI `Dialog` primitives, submitting a `POST` request to `/api/v1/filters` on confirmation. The dialog SHALL provide a way to load the currently active configuration (the active override if one exists, otherwise the origin configuration) for the selected attribute into the Include/Exclude fields, and SHALL warn the user before submission if an active override for the selected attribute will be replaced.

#### Scenario: Successfully create a new inclusion filter
- **WHEN** the user opens the "Créer Filtre" dialog, inputs name, selects attribute, adds inclusion patterns, and clicks "Enregistrer"
- **THEN** the frontend SHALL submit a `POST` request to `/api/v1/filters`, display a success notification, close the dialog, and refresh the filters list

#### Scenario: Load the currently active configuration into the form
- **WHEN** the user selects an attribute and clicks "Reprendre la config actuelle"
- **THEN** the frontend SHALL populate the Include and Exclude fields with the active runtime override's patterns for that attribute if one exists, or with the origin `config.yml` patterns for that attribute otherwise, leaving the fields editable

#### Scenario: Warning before replacing an existing active override
- **WHEN** the user selects an attribute that already has an active runtime override and submits the "Créer Filtre" form
- **THEN** the frontend SHALL display a warning that the existing active override for that attribute will be replaced before the request is submitted
