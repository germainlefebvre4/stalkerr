## Purpose

Defines how file-defined M3U sources and database-backed runtime sources relate to each other, keyed by source name: how origin sources are exposed for display, how runtime sources replace or add to them, and how the effective source list is resolved.

## ADDED Requirements

### Requirement: Origin M3U Source Exposure
The system SHALL expose the `config.yml`-defined M3U sources (origin configuration) via a read-only API endpoint, independently of whether a runtime source exists for any of them.

#### Scenario: Fetching origin sources
- **WHEN** a client requests the origin M3U source configuration
- **THEN** the system SHALL return every source defined in `config.yml`'s `m3u.sources` list, each with its `name`, `file_path`, and `download` settings, with `download.auth_password` reported per Sensitive Source Field Masking rather than returned raw

#### Scenario: No sources defined in config.yml
- **WHEN** `config.yml` defines no `m3u.sources` entries
- **THEN** the system SHALL return an empty list of origin sources rather than an error

### Requirement: Runtime Source, Keyed By Name
The system SHALL allow storing a runtime M3U source, identified by `name`. Storing a runtime source whose `name` matches an origin source SHALL replace that origin source entirely (its `file_path` and full `download` block) for the purposes of the effective source list; storing a runtime source whose `name` does not match any origin source SHALL add a new source that exists only at runtime. Storing a runtime source whose `name` already has a stored runtime source SHALL replace that stored runtime source's `file_path` and full `download` block in place, regardless of whether an origin source of the same name also exists; it SHALL NOT create a second runtime source for the same `name` or fail as a duplicate. Every store SHALL submit a complete `download` block; `download.auth_password` is the one field that MAY be omitted from that block, per Sensitive Source Field Masking, to mean "keep the current effective password unchanged" rather than clearing it.

#### Scenario: Overriding an origin source by name
- **WHEN** a runtime source is stored with the same `name` as an origin source, using a different `file_path` or `download` settings
- **THEN** the effective source list SHALL use the runtime source's `file_path` and `download` settings for that `name`, not the origin source's

#### Scenario: Adding a source that has no origin counterpart
- **WHEN** a runtime source is stored with a `name` that does not appear in `config.yml`'s `m3u.sources`
- **THEN** the effective source list SHALL include that source in addition to every origin source

#### Scenario: Re-storing a runtime source replaces it in place
- **WHEN** a runtime source already exists for name "c" (with no origin counterpart), and a new runtime source is stored for "c" with different `download` settings
- **THEN** the effective source for "c" SHALL reflect only the newly stored settings, replacing the previous runtime source, rather than adding a second runtime source for "c"

### Requirement: Effective Source List Resolution
The system SHALL resolve the effective M3U source list as: for each `name` known from either origin configuration or stored runtime sources, the runtime source if one exists for that `name`, otherwise the origin source. The system SHALL NOT merge fields between an origin source and a runtime source sharing the same `name`. This does not include the omission-based handling of `download.auth_password` (see Sensitive Source Field Masking), which resolves the value for that single field on a single store operation and is not a merge across origin and runtime.

#### Scenario: Mixed origin and runtime sources
- **WHEN** `config.yml` defines sources "a" and "b", and a runtime source is stored for "b" and a new runtime source "c"
- **THEN** the effective source list SHALL contain the origin source "a", the runtime source "b" in place of the origin one, and the runtime source "c"

#### Scenario: Effective list may be empty
- **WHEN** no origin sources are defined and no runtime sources are stored
- **THEN** the effective source list SHALL be empty, and the system SHALL treat this as a valid, if inactive, configuration rather than an error

### Requirement: Removing a Runtime Source
The system SHALL allow deleting a stored runtime source by name. If an origin source of the same name exists in `config.yml`, deleting the runtime source SHALL revert the effective source for that name to the origin source; if no origin source of that name exists, deleting the runtime source SHALL remove it from the effective source list entirely. Deleting a `name` for which no runtime source is stored SHALL be rejected as not found, whether or not an origin source exists for that `name`, and SHALL NOT affect the origin source.

#### Scenario: Deleting a runtime override reverts to origin
- **WHEN** a runtime source overriding origin source "b" is deleted
- **THEN** the effective source for "b" SHALL revert to the origin `config.yml` definition

#### Scenario: Deleting a runtime-only source removes it
- **WHEN** a runtime source "c", which has no origin counterpart, is deleted
- **THEN** "c" SHALL no longer appear in the effective source list

#### Scenario: Deleting a name with no stored runtime source
- **WHEN** a deletion is requested for a `name` that has no stored runtime source, including a `name` that only exists as an origin source
- **THEN** the system SHALL reject the request as not found, and the origin source for that `name`, if any, SHALL continue to be used unchanged

### Requirement: Sensitive Source Field Masking
The system SHALL report whether `download.auth_password` is currently set for a source (origin, effective, or a stored runtime source) as a boolean, never returning its raw value. `download.auth_username` is not masked. When storing a runtime source, the system SHALL treat an omitted `download.auth_password` as "keep the `name`'s current effective password unchanged" (whether that password currently comes from the origin source or a prior stored runtime source), a submitted non-empty value as a new password to store, and a submitted empty value as clearing the password for that runtime source.

#### Scenario: Reading a source with an auth password set
- **WHEN** a client requests a source (origin or effective) whose `download.auth_password` is non-empty
- **THEN** the system SHALL report that a password is set, as a boolean, without returning the raw value

#### Scenario: Reading a source with no auth password set
- **WHEN** a client requests a source (origin or effective) whose `download.auth_password` is empty
- **THEN** the system SHALL report that no password is set

#### Scenario: Storing a runtime source without submitting a password keeps the current one
- **WHEN** a client stores a runtime source for a `name` whose current effective `download.auth_password` is non-empty, omitting `download.auth_password` from the request
- **THEN** the system SHALL store that same effective password unchanged for the resulting runtime source, regardless of whether it previously came from the origin source or a prior runtime source

#### Scenario: Storing a runtime source with a new literal password
- **WHEN** a client stores a runtime source submitting a non-empty `download.auth_password`
- **THEN** the system SHALL store the submitted value as the new password for that runtime source

#### Scenario: Storing a runtime source with an empty password clears it
- **WHEN** a client stores a runtime source submitting `download.auth_password` as an explicit empty value
- **THEN** the system SHALL store the runtime source with no password set
