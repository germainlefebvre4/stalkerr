## MODIFIED Requirements

### Requirement: Download fallback loop over quality candidates
The unified `download` command SHALL attempt each candidate URL, for both movies and TV episodes, in quality-preference order. On download failure, the failed `ProcessedLine` SHALL be marked `state = "failed"` and the next candidate SHALL be attempted. The loop stops on the first successful download.

#### Scenario: First candidate fails, second succeeds
- **WHEN** the preferred 720p URL returns a network error
- **THEN** its `ProcessedLine.state` SHALL be set to `"failed"` and the next candidate (1080p) SHALL be attempted

#### Scenario: All candidates fail
- **WHEN** every candidate URL fails to download
- **THEN** the content item SHALL be counted as failed and the command SHALL continue to the next content item without crashing

#### Scenario: Successful download stops the loop
- **WHEN** the 720p candidate downloads successfully
- **THEN** no further candidates SHALL be attempted for that content item
