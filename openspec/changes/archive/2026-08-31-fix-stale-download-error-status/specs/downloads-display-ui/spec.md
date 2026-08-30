## ADDED Requirements

### Requirement: Error Banner Gated on Failed Status
The Downloads tab SHALL render a download card's error banner only when that download's current `status` is `failed`. The presence of a non-empty `error_message` alone SHALL NOT be sufficient to render the banner, since `error_message` may be a value carried over from a prior attempt on the same record.

#### Scenario: Completed download with a leftover error_message does not show the banner
- **WHEN** a download card's `status` is `completed` and its `error_message` is non-empty (leftover from an earlier failed attempt)
- **THEN** the system SHALL NOT render the error banner on that card

#### Scenario: Failed download still shows the banner
- **WHEN** a download card's `status` is `failed` and its `error_message` is non-empty
- **THEN** the system SHALL render the error banner with the error message, as before

#### Scenario: In-progress download with a leftover error_message does not show the banner
- **WHEN** a download card's `status` is `downloading` or `retrying` and its `error_message` is non-empty (leftover from an earlier failed attempt)
- **THEN** the system SHALL NOT render the error banner on that card
