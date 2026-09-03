## ADDED Requirements

### Requirement: Resync with Radarr/Sonarr Action in the Sidepanel
For a completed download, the sidepanel SHALL offer a "Resynchroniser" action alongside the existing "Déplacer"/"Renommer" actions. Activating it SHALL call `POST /api/v1/downloads/:id/resync-path` (capability `api-media-management`) scoped to that single download's id, and SHALL display the outcome to the user: that the path was corrected (and update the displayed folder/path accordingly), that no correction was needed, or that the item is not managed by Radarr/Sonarr. The sidepanel SHALL NOT expose an active "Resynchroniser" action for a download that is not completed. This action is read-and-correct only: it SHALL NOT be offered from the Erreurs tab's dedicated sidepanel (capability `downloads-errors-view`), which remains free of corrective actions.

#### Scenario: Resync corrects a stale path
- **WHEN** the user clicks "Resynchroniser" in the sidepanel of a completed download whose Radarr/Sonarr root has since changed
- **THEN** the system SHALL call the resync endpoint, update only that download's displayed folder/path on success, and show a confirmation that the path was corrected

#### Scenario: Resync reports the path was already correct
- **WHEN** the user clicks "Resynchroniser" and the endpoint reports no correction was needed
- **THEN** the system SHALL show a message indicating the path was already up to date, without altering the displayed download

#### Scenario: Resync reports the item is not managed by Radarr/Sonarr
- **WHEN** the user clicks "Resynchroniser" and the endpoint responds with `"not_managed_by_radarr_sonarr"`
- **THEN** the system SHALL show a message indicating the item is not managed by Radarr/Sonarr, without altering the displayed download

#### Scenario: Resync unavailable for a non-completed download
- **WHEN** the user opens the sidepanel for a download whose `status` is not `completed`
- **THEN** the sidepanel SHALL NOT offer an active "Resynchroniser" action

#### Scenario: Resync not offered in the Erreurs tab sidepanel
- **WHEN** the user opens the Erreurs tab's dedicated sidepanel for any download
- **THEN** the "Resynchroniser" action SHALL NOT be rendered there, consistent with that sidepanel offering no corrective actions
