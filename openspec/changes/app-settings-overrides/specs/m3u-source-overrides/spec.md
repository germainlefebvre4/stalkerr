## Purpose

Defines how file-defined M3U sources and database-backed runtime sources relate to each other, keyed by source name: how origin sources are exposed for display, how runtime sources replace or add to them, and how the effective source list is resolved.

## ADDED Requirements

### Requirement: Origin M3U Source Exposure
The system SHALL expose the `config.yml`-defined M3U sources (origin configuration) via a read-only API endpoint, independently of whether a runtime source exists for any of them.

#### Scenario: Fetching origin sources
- **WHEN** a client requests the origin M3U source configuration
- **THEN** the system SHALL return every source defined in `config.yml`'s `m3u.sources` list, each with its `name`, `file_path`, and `download` settings

#### Scenario: No sources defined in config.yml
- **WHEN** `config.yml` defines no `m3u.sources` entries
- **THEN** the system SHALL return an empty list of origin sources rather than an error

### Requirement: Runtime Source, Keyed By Name
The system SHALL allow storing a runtime M3U source, identified by `name`. Storing a runtime source whose `name` matches an origin source SHALL replace that origin source entirely (its `file_path` and full `download` block) for the purposes of the effective source list; storing a runtime source whose `name` does not match any origin source SHALL add a new source that exists only at runtime.

#### Scenario: Overriding an origin source by name
- **WHEN** a runtime source is stored with the same `name` as an origin source, using a different `file_path` or `download` settings
- **THEN** the effective source list SHALL use the runtime source's `file_path` and `download` settings for that `name`, not the origin source's

#### Scenario: Adding a source that has no origin counterpart
- **WHEN** a runtime source is stored with a `name` that does not appear in `config.yml`'s `m3u.sources`
- **THEN** the effective source list SHALL include that source in addition to every origin source

### Requirement: Effective Source List Resolution
The system SHALL resolve the effective M3U source list as: for each `name` known from either origin configuration or stored runtime sources, the runtime source if one exists for that `name`, otherwise the origin source. The system SHALL NOT merge fields between an origin source and a runtime source sharing the same `name`.

#### Scenario: Mixed origin and runtime sources
- **WHEN** `config.yml` defines sources "a" and "b", and a runtime source is stored for "b" and a new runtime source "c"
- **THEN** the effective source list SHALL contain the origin source "a", the runtime source "b" in place of the origin one, and the runtime source "c"

#### Scenario: Effective list may be empty
- **WHEN** no origin sources are defined and no runtime sources are stored
- **THEN** the effective source list SHALL be empty, and the system SHALL treat this as a valid, if inactive, configuration rather than an error

### Requirement: Removing a Runtime Source
The system SHALL allow deleting a stored runtime source by name. If an origin source of the same name exists in `config.yml`, deleting the runtime source SHALL revert the effective source for that name to the origin source; if no origin source of that name exists, deleting the runtime source SHALL remove it from the effective source list entirely.

#### Scenario: Deleting a runtime override reverts to origin
- **WHEN** a runtime source overriding origin source "b" is deleted
- **THEN** the effective source for "b" SHALL revert to the origin `config.yml` definition

#### Scenario: Deleting a runtime-only source removes it
- **WHEN** a runtime source "c", which has no origin counterpart, is deleted
- **THEN** "c" SHALL no longer appear in the effective source list
