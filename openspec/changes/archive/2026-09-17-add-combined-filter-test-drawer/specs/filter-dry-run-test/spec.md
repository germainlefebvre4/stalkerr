## MODIFIED Requirements

### Requirement: On-Demand Dry-Run Endpoint
The system SHALL expose an endpoint that accepts a source name and, for one or both of the attributes `group_title` and `tvg_name`, include/exclude patterns supplied by the caller, and SHALL evaluate exactly those supplied patterns against the identified source's most recently downloaded playlist archive — never the source's currently active runtime override or `config.yml` patterns for either attribute, and never a newly triggered download. An attribute for which the caller supplies no patterns SHALL impose no filtering on that attribute (every value matches), so supplying patterns for only one attribute reproduces today's single-attribute behavior unchanged.

#### Scenario: Testing caller-supplied patterns
- **WHEN** the caller submits a source name and include/exclude patterns for exactly one attribute
- **THEN** the system SHALL parse that source's latest downloaded archive and evaluate the supplied patterns against it for that attribute only, ignoring any active runtime override or origin `config.yml` patterns already configured for either attribute

#### Scenario: Testing caller-supplied patterns for both attributes together
- **WHEN** the caller submits a source name and include/exclude patterns for both `group_title` and `tvg_name`
- **THEN** the system SHALL evaluate exactly those caller-supplied patterns for both attributes against the archive, without looking up either attribute's active runtime override or origin `config.yml` patterns itself

#### Scenario: Testing patterns for an attribute with no saved filter yet
- **WHEN** the caller submits patterns for an attribute that has no active runtime override and no origin `config.yml` patterns
- **THEN** the system SHALL still evaluate the supplied patterns against the archive rather than requiring a filter to already exist

### Requirement: Aggregate Summary Result
By default (no content search requested), when the caller supplies patterns for exactly one attribute, the endpoint SHALL return the existing single-attribute aggregate summary: the total number of lines scanned, how many would match, how many would be excluded, and the top 20 distinct attribute values on each side (matched and excluded), each with the count of lines carrying that value, ordered by descending count. When the caller supplies patterns for both `group_title` and `tvg_name`, the endpoint SHALL instead return a combined summary: the total number of lines scanned, how many lines are kept (pass both attributes' patterns), how many are excluded by the `group_title` patterns only, how many by the `tvg_name` patterns only, how many by both, and, for each attribute individually, its own top 20 distinct matched and excluded values.

#### Scenario: Summary reflects the full archive
- **WHEN** the caller requests a dry-run for exactly one attribute without a content search
- **THEN** the system SHALL report the total lines scanned, the matched and excluded counts across the entire archive, and up to 20 of the most frequent distinct values on each side for that attribute

#### Scenario: Combined summary ventilated by cause
- **WHEN** the caller requests a dry-run for both attributes without a content search
- **THEN** the system SHALL evaluate each line against both attributes' patterns and report the total lines scanned, the count kept by both, and the counts excluded by `group_title` only, by `tvg_name` only, and by both, plus each attribute's own top 20 matched and excluded values

#### Scenario: Fewer than 20 distinct values on a side
- **WHEN** one side (matched or excluded) of one attribute has fewer than 20 distinct values
- **THEN** the system SHALL return all of them for that side rather than padding the list

### Requirement: Content Search Mode
The endpoint SHALL accept an optional search substring together with the name of the attribute the substring is matched against. When present, the system SHALL, instead of the aggregate summary, return the individual archive lines whose value for the named search attribute contains that substring (case-insensitive), capped at 100 results. When patterns were supplied for exactly one attribute, the search attribute SHALL be that same attribute and each returned line SHALL be tagged as would-match or would-be-excluded under the supplied patterns. When patterns were supplied for both attributes, the caller SHALL choose which of the two attributes to search against, and each returned line SHALL be tagged with its combined verdict: kept, or excluded — and, when excluded, by which attribute's patterns (or both).

#### Scenario: Searching for a specific line
- **WHEN** the caller requests a dry-run with a content search substring and patterns for exactly one attribute
- **THEN** the system SHALL return every archive line whose value for that attribute contains the substring, up to 100, each labeled as would-match or would-be-excluded

#### Scenario: Searching for a specific line under a combined test
- **WHEN** the caller requests a dry-run with a content search substring, patterns for both attributes, and a chosen search attribute
- **THEN** the system SHALL return every archive line whose value for the chosen search attribute contains the substring, up to 100, each labeled with its combined verdict (kept, or excluded and by which attribute's patterns)

#### Scenario: Search matches more than the cap
- **WHEN** more than 100 archive lines match the search substring
- **THEN** the system SHALL return the first 100 matches and SHALL indicate that the result was truncated

#### Scenario: No lines match the search
- **WHEN** no archive line's value for the search attribute contains the search substring
- **THEN** the system SHALL return an empty result rather than an error

### Requirement: Input Validation
The endpoint SHALL reject, without evaluating any archive, a request with: an unknown source name, patterns supplied for neither `group_title` nor `tvg_name`, a content-search request naming a search attribute other than `group_title` or `tvg_name`, a content-search request naming a search attribute for which no patterns were supplied, or an include/exclude pattern that fails to compile as a regular expression.

#### Scenario: Unknown source
- **WHEN** the caller submits a source name that does not match any configured or runtime M3U source
- **THEN** the system SHALL reject the request without reading any archive

#### Scenario: No attribute patterns supplied
- **WHEN** the caller submits a request with no include/exclude patterns for either `group_title` or `tvg_name`
- **THEN** the system SHALL reject the request without evaluating any archive

#### Scenario: Unsupported attribute
- **WHEN** the caller submits a content-search request naming a search attribute other than `group_title` or `tvg_name`, or naming an attribute for which no patterns were supplied
- **THEN** the system SHALL reject the request without evaluating any archive

#### Scenario: Invalid regular expression
- **WHEN** an include or exclude pattern supplied by the caller fails to compile as a regular expression
- **THEN** the system SHALL reject the request with an error identifying the invalid pattern, without evaluating the archive
