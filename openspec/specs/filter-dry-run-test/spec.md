# filter-dry-run-test Specification

## Purpose

Provides an on-demand backend check that evaluates a caller-supplied filter attribute and include/exclude pattern combination against a chosen M3U source's most recently downloaded playlist archive, reporting how much content would match or be excluded, without persisting anything or depending on previously saved filters.

## Requirements

### Requirement: On-Demand Dry-Run Endpoint
The system SHALL expose an endpoint that accepts a source name, a target attribute (`group_title` or `tvg_name`), and include/exclude patterns supplied by the caller, and SHALL evaluate those exact supplied patterns against the identified source's most recently downloaded playlist archive — never the source's currently active runtime override or `config.yml` patterns for that attribute, and never a newly triggered download.

#### Scenario: Testing caller-supplied patterns
- **WHEN** the caller submits a source name, an attribute, and include/exclude patterns
- **THEN** the system SHALL parse that source's latest downloaded archive and evaluate the supplied patterns against it, ignoring any active runtime override or origin `config.yml` patterns already configured for that attribute

#### Scenario: Testing patterns for an attribute with no saved filter yet
- **WHEN** the caller submits patterns for an attribute that has no active runtime override and no origin `config.yml` patterns
- **THEN** the system SHALL still evaluate the supplied patterns against the archive rather than requiring a filter to already exist

### Requirement: Aggregate Summary Result
By default (no content search requested), the endpoint SHALL return an aggregate summary: the total number of lines scanned, how many would match the supplied patterns, how many would be excluded, and the top 20 distinct attribute values on each side (matched and excluded), each with the count of lines carrying that value, ordered by descending count.

#### Scenario: Summary reflects the full archive
- **WHEN** the caller requests a dry-run without a content search
- **THEN** the system SHALL report the total lines scanned, the matched and excluded counts across the entire archive, and up to 20 of the most frequent distinct values on each side

#### Scenario: Fewer than 20 distinct values on a side
- **WHEN** one side (matched or excluded) has fewer than 20 distinct values
- **THEN** the system SHALL return all of them for that side rather than padding the list

### Requirement: Content Search Mode
The endpoint SHALL accept an optional search substring. When present, the system SHALL, instead of the aggregate summary, return the individual archive lines whose value for the tested attribute contains that substring (case-insensitive), each tagged as would-match or would-be-excluded under the supplied patterns, capped at 100 results. The search SHALL only consider the attribute being tested, not the other filterable attribute.

#### Scenario: Searching for a specific line
- **WHEN** the caller requests a dry-run with a content search substring
- **THEN** the system SHALL return every archive line whose tested-attribute value contains that substring, up to 100, each labeled as would-match or would-be-excluded

#### Scenario: Search matches more than the cap
- **WHEN** more than 100 archive lines match the search substring
- **THEN** the system SHALL return the first 100 matches and SHALL indicate that the result was truncated

#### Scenario: No lines match the search
- **WHEN** no archive line's tested-attribute value contains the search substring
- **THEN** the system SHALL return an empty result rather than an error

### Requirement: Missing Archive Handling
When the identified source has no previously downloaded archive available, the endpoint SHALL report this explicitly as a distinct, machine-readable condition and SHALL NOT attempt to download the source live.

#### Scenario: Source never downloaded
- **WHEN** the caller requests a dry-run for a source that has no archived download yet
- **THEN** the system SHALL respond with a distinct "no archive available" condition and SHALL NOT initiate a network download of that source

### Requirement: Input Validation
The endpoint SHALL reject, without evaluating any archive, a request with: an unknown source name, an attribute other than `group_title` or `tvg_name`, or an include/exclude pattern that fails to compile as a regular expression.

#### Scenario: Unknown source
- **WHEN** the caller submits a source name that does not match any configured or runtime M3U source
- **THEN** the system SHALL reject the request without reading any archive

#### Scenario: Unsupported attribute
- **WHEN** the caller submits an attribute other than `group_title` or `tvg_name`
- **THEN** the system SHALL reject the request without evaluating any pattern

#### Scenario: Invalid regular expression
- **WHEN** an include or exclude pattern supplied by the caller fails to compile as a regular expression
- **THEN** the system SHALL reject the request with an error identifying the invalid pattern, without evaluating the archive

### Requirement: No Persistence
The endpoint SHALL NOT create, modify, or delete any stored filter (runtime override or otherwise), and SHALL NOT modify `config.yml`, as a side effect of performing a dry-run.

#### Scenario: Dry-run does not affect saved filters
- **WHEN** a dry-run is performed with patterns that differ from the attribute's currently active filter
- **THEN** the currently active runtime override and origin configuration for that attribute SHALL remain unchanged after the dry-run completes
