## ADDED Requirements

### Requirement: Cancel Action in the Sidepanel
For a download whose `status` is `pending`, `failed`, or `retrying`, the sidepanel SHALL offer an "Annuler" action alongside whichever other actions are available for that status. Activating it SHALL call the cancel endpoint (capability `cancel-download-occurrence`) scoped to that single download's id, and on success SHALL update only that download's displayed status to the new terminal excluded status. The sidepanel SHALL NOT offer an active "Annuler" action for a download whose `status` is `completed` or `downloading`, or that is already in the terminal excluded status.

#### Scenario: Cancel offered for a failed download
- **WHEN** the user opens the sidepanel for a download whose `status` is `failed`
- **THEN** the sidepanel SHALL show an active "Annuler" action

#### Scenario: Cancelling updates only the selected download
- **WHEN** the user clicks "Annuler" for a `failed` download and the cancel request succeeds
- **THEN** the system SHALL update only that download's displayed status, leaving every other download in the list unchanged, without requiring a full page reload

#### Scenario: Cancel unavailable for a completed download
- **WHEN** the user opens the sidepanel for a download whose `status` is `completed`
- **THEN** the sidepanel SHALL NOT offer an active "Annuler" action

#### Scenario: Cancel unavailable for an actively downloading item
- **WHEN** the user opens the sidepanel for a download whose `status` is `downloading`
- **THEN** the sidepanel SHALL NOT offer an active "Annuler" action

#### Scenario: Cancel unavailable once already in the terminal excluded status
- **WHEN** the user opens the sidepanel for a download that is already cancelled or retry-exhausted
- **THEN** the sidepanel SHALL NOT offer an active "Annuler" action
