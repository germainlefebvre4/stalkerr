## ADDED Requirements

### Requirement: Completion Date on Download Cards

**GIVEN** a download whose `completed_at` timestamp is set
**WHEN** rendering the download card
**THEN** the system SHALL display the formatted completion date inline in the technical specs row, alongside Format, Resolution, Size, and Duration, using the same locale-aware date formatting as the Playlist tab's `downloaded_at` column.

**GIVEN** a download whose `completed_at` timestamp is not set (pending, downloading, failed, or retrying)
**WHEN** rendering the download card
**THEN** the system SHALL NOT display a completion date or a placeholder for one in the technical specs row.

**GIVEN** the Downloads tab rendered on a viewport narrower than the mobile breakpoint
**WHEN** a download card shows its completion date
**THEN** the system SHALL display the same completion date, in the same technical specs row, as on desktop, without truncation or omission.

#### Scenario: Completed download shows its completion date
- **WHEN** a download's `completed_at` is `2026-08-27T14:33:00Z` and the technical specs row is rendered
- **THEN** the row SHALL include the formatted completion date alongside Format, Resolution, Size, and Duration

#### Scenario: In-progress download shows no completion date
- **WHEN** a download's `completed_at` is absent and the technical specs row is rendered
- **THEN** the row SHALL omit any completion date element or placeholder, leaving the existing Format/Resolution/Size/Duration content unchanged

#### Scenario: Completion date renders identically on mobile
- **WHEN** the Downloads tab is rendered on a viewport narrower than the mobile breakpoint and a download card has a completion date
- **THEN** the card SHALL display the same completion date in the technical specs row as it would on desktop
