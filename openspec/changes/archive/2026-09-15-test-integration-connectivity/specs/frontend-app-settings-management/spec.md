## ADDED Requirements

### Requirement: Test Connection Action for Radarr, Sonarr, and Jellyfin
For the Radarr, Sonarr, and Jellyfin field groups, the frontend SHALL provide a "Test connection" action, appearing in the group's action area whenever that group has an unsaved change, that submits the group's current in-progress field values (not necessarily saved) to the connectivity-test endpoint and displays the result as a status badge on the card, distinguishing OK from KO with its specific reason. The frontend SHALL NOT submit a value for a sensitive field the user has not explicitly retyped since it was masked; that field is sent empty.

#### Scenario: Testing an unsaved change
- **WHEN** the user has an unsaved change in the Radarr, Sonarr, or Jellyfin card and clicks "Tester la connexion"
- **THEN** the frontend submits the card's current in-form values, including any unsaved edits, to the connectivity-test endpoint and displays the returned OK/KO-with-reason result as a badge on the card

#### Scenario: Test action requires an unsaved change
- **WHEN** the Radarr, Sonarr, or Jellyfin card has no unsaved change
- **THEN** the frontend SHALL NOT show the "Test connection" action for that card

#### Scenario: Testing without retyping a sensitive field
- **WHEN** the user modifies the URL field but does not retype the API key field
- **THEN** the frontend submits the test with an empty API key value rather than the previously saved key

#### Scenario: Result distinguishes the failure reason
- **WHEN** the connectivity test returns KO with reason "unauthorized"
- **THEN** the badge SHALL display that specific reason, distinguishing it from unreachable, timeout, or other failure reasons

#### Scenario: Test action disabled without a URL
- **WHEN** the card's URL field is empty
- **THEN** the frontend SHALL disable the "Test connection" action rather than submitting a request
