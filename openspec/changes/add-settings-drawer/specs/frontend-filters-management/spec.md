## MODIFIED Requirements

### Requirement: Filters List View
The frontend SHALL provide a collapsible "Filtres" section within the Settings drawer, collapsed by default, that groups filter configuration by target attribute (Group Title, TVG Name). Expanding the section SHALL query `GET /api/v1/filters` for the active runtime override and the origin configuration endpoint for the `config.yml`-defined patterns, and SHALL render both: the origin patterns (read-only, labeled as origin/system) and the active runtime override, if any (labeled as an active override), so the user can see at a glance which configuration is currently in effect for that attribute.

#### Scenario: View current filter configurations
- **WHEN** the user opens the Settings drawer and expands the "Filtres" section
- **THEN** the frontend SHALL fetch both the origin configuration and the active runtime overrides, and render one section per attribute (Group Title, TVG Name) showing the origin include/exclude patterns and, if present, the active override's name and include/exclude patterns

#### Scenario: Attribute with no active override
- **WHEN** an attribute has no active runtime override
- **THEN** the frontend SHALL display only the origin `config.yml` patterns for that attribute, with no override section

#### Scenario: Attribute with an active override
- **WHEN** an attribute has an active runtime override
- **THEN** the frontend SHALL visually distinguish the origin patterns (labeled as origin/system) from the active override (labeled as an active override), making clear that the override is what is currently applied
