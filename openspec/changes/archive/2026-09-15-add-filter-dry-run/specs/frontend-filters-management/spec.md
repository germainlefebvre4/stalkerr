## ADDED Requirements

### Requirement: Dry-Run Filter Testing
The frontend SHALL allow the user to dry-run test a filter's include/exclude patterns against a chosen M3U source's real playlist content, from two places: the create dialog (testing the patterns currently typed into the form, before saving) and each filter card in the "Filtres" section (testing the patterns already in effect — origin or active override — for that card, read-only). Both SHALL submit to the dry-run endpoint and render its aggregate summary (total lines scanned, matched count, excluded count, top matched and excluded values) in the same place the test was triggered from.

#### Scenario: Testing patterns while creating a filter
- **WHEN** the user has typed include/exclude patterns into the create dialog, selected a source, and clicks "Tester"
- **THEN** the frontend SHALL submit those exact in-progress values (not yet saved) to the dry-run endpoint and display the aggregate summary within the dialog, without submitting the create request

#### Scenario: Testing an already-saved filter from its card
- **WHEN** the user clicks "Tester" on a filter card (origin or active override) in the "Filtres" section
- **THEN** the frontend SHALL submit that card's existing include/exclude patterns and attribute to the dry-run endpoint, with a source selected by the user, and display the aggregate summary inline on the card without opening the create dialog

#### Scenario: Source has no archive yet
- **WHEN** the dry-run endpoint reports that the selected source has no downloaded archive available
- **THEN** the frontend SHALL display a message explaining that no data is available for that source yet, and SHALL NOT trigger a download

#### Scenario: Invalid pattern prevents testing
- **WHEN** the dry-run endpoint rejects the request because an include or exclude pattern is not a valid regular expression
- **THEN** the frontend SHALL display an inline error identifying the invalid pattern instead of a summary

### Requirement: Dry-Run Content Search
Once a dry-run's aggregate summary is displayed, the frontend SHALL offer a search field scoped to the tested attribute that lets the user verify specific content: submitting a search term SHALL query the dry-run endpoint's content search mode and display each matching line together with whether it would match or be excluded under the tested patterns.

#### Scenario: Verifying a specific line was excluded
- **WHEN** the user enters a channel or group name in the dry-run search field after viewing a summary
- **THEN** the frontend SHALL display every archive line whose value for the tested attribute contains that text, each labeled as would-match or would-be-excluded, so the user can confirm the expected lines disappeared or remained

#### Scenario: Search available only after a summary exists
- **WHEN** the user has not yet run a dry-run test
- **THEN** the frontend SHALL NOT display the content search field
