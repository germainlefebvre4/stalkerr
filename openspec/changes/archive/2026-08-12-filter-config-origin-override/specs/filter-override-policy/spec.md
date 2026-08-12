## Purpose

Defines how the file-based (origin) filter configuration and the database-backed runtime overrides relate to each other: how origin config is exposed for display, and how many runtime overrides may be active per attribute at once.

## ADDED Requirements

### Requirement: Origin Filter Configuration Exposure
The system SHALL expose the `config.yml`-defined filter patterns (origin configuration) for each supported attribute (`group_title`, `tvg_name`) via a read-only API endpoint, independently of whether a runtime override exists for that attribute.

#### Scenario: Fetching origin configuration
- **WHEN** a client requests the origin filter configuration
- **THEN** the system SHALL return, for each supported attribute, the include and exclude patterns currently defined in `config.yml`

#### Scenario: Attribute with no configured patterns
- **WHEN** `config.yml` defines no include or exclude patterns for a given attribute
- **THEN** the system SHALL return that attribute with empty include and exclude pattern lists rather than omitting it or returning an error

### Requirement: Single Active Runtime Override Per Attribute
The system SHALL allow at most one active runtime filter per attribute (`group_title` or `tvg_name`) at any time. Creating or updating a runtime filter for an attribute that already has an active runtime override SHALL automatically replace the previous override rather than allowing both to remain active.

#### Scenario: Create runtime filter on attribute with no active override
- **WHEN** a runtime filter is created for an attribute that currently has no active runtime override
- **THEN** the system SHALL activate the new filter as the sole runtime override for that attribute

#### Scenario: Create runtime filter on attribute with an existing active override
- **WHEN** a runtime filter is created for an attribute that already has an active runtime override
- **THEN** the system SHALL deactivate (remove) the previous override for that attribute and activate the new filter as the sole runtime override, without requiring a separate confirmation step at the API level

#### Scenario: Update runtime filter's attribute onto an attribute with an existing active override
- **WHEN** an existing runtime filter is updated so that its attribute changes to one that already has a different active override
- **THEN** the system SHALL deactivate (remove) the previous override on the target attribute and activate the updated filter as the sole runtime override for that attribute

### Requirement: Effective Filter Resolution Per Attribute
For each attribute, the system SHALL apply the active runtime override's patterns if one exists, or the origin `config.yml` patterns otherwise, when deciding whether an M3U entry is accepted. The system SHALL NOT combine origin and runtime patterns for the same attribute.

#### Scenario: No active runtime override
- **WHEN** an attribute has no active runtime override
- **THEN** the system SHALL apply the origin `config.yml` include/exclude patterns for that attribute when matching entries

#### Scenario: Active runtime override present
- **WHEN** an attribute has an active runtime override
- **THEN** the system SHALL apply only that override's include/exclude patterns for that attribute when matching entries, ignoring the origin `config.yml` patterns for that attribute
