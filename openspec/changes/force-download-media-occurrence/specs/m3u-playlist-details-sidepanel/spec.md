## ADDED Requirements

### Requirement: Force-download action in the occurrence sidepanel
The playlist sidepanel (drawer) SHALL offer a "Forcer le téléchargement" action for the single occurrence currently displayed, letting the user request a forced download of exactly that occurrence without affecting any of its sibling occurrences.

The action SHALL be unavailable (hidden or disabled, with an explanatory reason) when the displayed occurrence is not matched to a movie or TV show, or when the occurrence's own state already reflects a download that is in progress or already completed.

After the action is triggered, the sidepanel SHALL reflect the outcome without requiring the user to wait for the underlying file transfer: a distinguishable error message when the media could not be confirmed to exist in Radarr/Sonarr (or when that check failed), and a distinct "queued"/in-progress acknowledgement when the request was accepted, after which the sidepanel remains usable and closable while the transfer continues in the background.

#### Scenario: Action available for an eligible occurrence
- **WHEN** the user opens the sidepanel for an occurrence matched to a movie or TV show, not yet downloaded and not currently downloading
- **THEN** the sidepanel SHALL display an enabled "Forcer le téléchargement" action

#### Scenario: Action unavailable for an unmatched occurrence
- **WHEN** the user opens the sidepanel for an occurrence with no associated movie or TV show
- **THEN** the sidepanel SHALL NOT offer an active "Forcer le téléchargement" action for that occurrence

#### Scenario: Action unavailable for an already-downloaded occurrence
- **WHEN** the user opens the sidepanel for an occurrence whose own state is already "downloaded", or for which a download is already in progress
- **THEN** the sidepanel SHALL NOT offer an active "Forcer le téléchargement" action for that occurrence

#### Scenario: Successful trigger shows a queued acknowledgement
- **WHEN** the user clicks "Forcer le téléchargement" and the request is accepted (media confirmed to exist, eligibility checks passed)
- **THEN** the sidepanel SHALL display a distinct queued/in-progress acknowledgement, and SHALL remain usable and closable while the download continues in the background

#### Scenario: Failed trigger shows an explicit error
- **WHEN** the user clicks "Forcer le téléchargement" and the request is refused (media not found in Radarr/Sonarr, the existence check failed, or the occurrence turned out ineligible)
- **THEN** the sidepanel SHALL display an explicit, distinguishable error message and SHALL NOT show a queued/in-progress acknowledgement
