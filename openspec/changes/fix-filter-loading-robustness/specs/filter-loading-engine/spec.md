## Purpose

Defines the contract for how filter patterns are validated when written, how they are parsed and loaded from `config.yml` and the database, and how loading failures are isolated and reported so one bad pattern cannot silently disable unrelated filters.

## ADDED Requirements

### Requirement: Runtime Filter Pattern Persistence Format
The system SHALL interpret a stored runtime filter's `include_patterns`/`exclude_patterns` value as a comma-separated list of individual regex patterns (with surrounding whitespace trimmed and empty tokens dropped), matching the format the API accepts and persists, so every runtime filter that is successfully created can also be successfully loaded and applied.

#### Scenario: A created runtime filter is loaded and applied
- **WHEN** a runtime filter is created via the API with `include_patterns` set to `"FRENCH, VFF"`
- **THEN** the filter loading engine SHALL later parse this into the two distinct patterns `FRENCH` and `VFF` and apply both when matching entries for that filter's attribute

#### Scenario: Trailing or duplicate separators are tolerated
- **WHEN** a stored pattern value contains extra whitespace or a trailing comma (e.g. `"FRENCH, VFF, "`)
- **THEN** the filter loading engine SHALL ignore the resulting empty token and load only the non-empty patterns

### Requirement: Per-Attribute Load Isolation
The system SHALL isolate a pattern-compilation failure in one attribute's origin (`config.yml`) configuration to that attribute only: it SHALL NOT prevent the other attribute's origin configuration, or any runtime override on either attribute, from loading.

#### Scenario: One attribute's origin pattern is invalid
- **WHEN** `config.yml` defines an invalid regex pattern for `tvg_name` but a valid configuration for `group_title`
- **THEN** the system SHALL still load `group_title`'s origin configuration, still attempt to load runtime overrides for both attributes, and SHALL treat `tvg_name` as having no origin filter (allowing all values for that attribute) rather than failing to load anything

### Requirement: Per-Row Load Isolation for Runtime Filters
The system SHALL isolate a load failure (unparsable or non-compiling pattern) in one runtime filter row to that row only: it SHALL NOT prevent other runtime filter rows, or any origin configuration, from loading.

#### Scenario: One runtime filter row fails to load
- **WHEN** a runtime filter row exists whose patterns cannot be compiled as valid regex
- **THEN** the system SHALL skip that row, log the failure, and continue loading all other runtime filter rows and all origin configuration

### Requirement: Pattern Validation at Write Time
The system SHALL validate that every pattern submitted in `include_patterns`/`exclude_patterns` compiles as a valid regular expression before persisting a runtime filter, on both creation and update.

#### Scenario: Creating a filter with an invalid pattern
- **WHEN** a client submits a request to create a runtime filter containing a pattern that does not compile as a valid regular expression
- **THEN** the system SHALL reject the request with a `400` response and an error code of `invalid_pattern`, and SHALL NOT persist the filter

#### Scenario: Updating a filter with an invalid pattern
- **WHEN** a client submits a request to update a runtime filter's patterns and one of the new patterns does not compile as a valid regular expression
- **THEN** the system SHALL reject the request with a `400` response and an error code of `invalid_pattern`, and SHALL leave the existing filter unchanged

### Requirement: Load Failure Observability
When the system skips an attribute's origin configuration or a runtime filter row due to a pattern load failure, it SHALL log the affected attribute, the offending pattern, and the underlying error.

#### Scenario: A load failure is logged with enough detail to act on
- **WHEN** the system skips an attribute or a runtime filter row due to a pattern load failure
- **THEN** the system SHALL emit a log entry identifying which attribute (and, for a runtime filter, which filter) and which pattern caused the failure
