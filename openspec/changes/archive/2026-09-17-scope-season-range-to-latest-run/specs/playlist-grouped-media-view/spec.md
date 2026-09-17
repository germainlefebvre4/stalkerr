## MODIFIED Requirements

### Requirement: Grouped Row Content
Each movie or TV-show group entry SHALL include the title and year. Each TV-show group entry SHALL additionally include the range of seasons covered by the episodes contributing to that group (e.g. `S01`-`S05`), computed from the minimum and maximum non-null season values among the episodes belonging to the group's most-recent-processing-run (the run reported by its run-attribution field, capability `processing-run-items`); when the group has no run-attribution value, the range SHALL instead be computed from all of the group's underlying items. Movie group entries SHALL NOT include a season range.

#### Scenario: Movie group has no season range
- **WHEN** a movie group entry is returned
- **THEN** it SHALL include its title and year and SHALL NOT include season range information.

#### Scenario: TV show group spans multiple seasons
- **WHEN** a TV show's episodes from its most recent processing run span seasons 1 through 5
- **THEN** the group entry's season range SHALL be reported as spanning `S01` to `S05`.

#### Scenario: TV show group is a single season
- **WHEN** all of a TV show's episodes from its most recent processing run belong to season 2
- **THEN** the group entry's season range SHALL be reported as a single season, `S02`, rather than a range.

#### Scenario: Season range narrows to the latest run, not the show's full history
- **WHEN** a TV show has episodes from seasons 1 through 12 ingested by earlier processing runs, and its most recent processing run only added episodes for seasons 13 and 14
- **THEN** the group entry's season range SHALL be reported as `S13`-`S14`, not `S01`-`S14`.

#### Scenario: Season range falls back to full history when no run is attributed
- **WHEN** a TV show group's most recent contributing item predates run attribution being recorded, so the group has no run-attribution value
- **THEN** the group entry's season range SHALL be computed from all of that show's underlying items, matching the behavior before run-scoping was introduced.
