# Spec Delta

## MODIFIED Requirements

### Requirement: Cancelling is reversible only through the existing full media reset
The system SHALL NOT provide a targeted "reactivate" action for a cancelled or retry-exhausted occurrence. The only ways to make that exact occurrence eligible again SHALL be the existing full reset of its movie or TV show (capability `media-reset`), which clears the entire download history for that movie/show, or an explicit forced-download request targeting that same occurrence (capability `force-download-media-occurrence`), which reactivates only that occurrence.

#### Scenario: Reversing a cancellation requires a full reset
- **WHEN** a user wants to retry a cancelled occurrence without forcing it
- **THEN** they SHALL use the existing movie/TV show reset endpoint, which clears all processed lines for that movie/show, rather than a targeted reactivate action

#### Scenario: A forced download reactivates only the targeted cancelled occurrence
- **WHEN** a user submits a forced-download request for one cancelled occurrence of a movie that has other cancelled occurrences
- **THEN** only the targeted occurrence SHALL be reactivated, and the other cancelled occurrences SHALL remain cancelled
