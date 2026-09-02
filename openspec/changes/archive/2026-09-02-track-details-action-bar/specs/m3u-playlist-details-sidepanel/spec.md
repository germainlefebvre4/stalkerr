## MODIFIED Requirements

### Requirement: Force-download action in the occurrence sidepanel
The playlist sidepanel (drawer) SHALL offer a "Forcer le téléchargement" action for the single occurrence currently displayed, letting the user request a forced download of exactly that occurrence without affecting any of its sibling occurrences.

This action SHALL be presented together with the "Associate" action (see Requirement: Associate action in the occurrence sidepanel) in a single action bar positioned near the top of the sidepanel, directly below the panel header, rather than in a separate section further down the panel. Any eligibility hint, queued acknowledgement, or error message related to this action SHALL be displayed directly beneath this action bar.

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

#### Scenario: Force Download and Associate share one action bar
- **WHEN** the user opens the sidepanel for any occurrence
- **THEN** the "Associate" and "Forcer le téléchargement" actions SHALL both render side by side in a single action bar near the top of the sidepanel, above the TMDB metadata, pipeline state, and provenance sections

## ADDED Requirements

### Requirement: Associate action in the occurrence sidepanel
The playlist sidepanel (drawer) SHALL offer an "Associate" action for the occurrence currently displayed, opening the same manual TMDB association dialog used by the Playlist table's "Associate"/"Correct" action, pre-targeted at that occurrence. This action SHALL be available regardless of whether the occurrence is already matched to a movie/TV show, already downloaded, or mid-download, since correcting or confirming a TMDB match is independent of the occurrence's download state.

#### Scenario: Associate action opens the manual override dialog for the displayed occurrence
- **WHEN** the user clicks "Associate" in the sidepanel for a given occurrence
- **THEN** the manual override dialog SHALL open targeting that occurrence, identical to the dialog opened via the Playlist table's row-level "Associate" button

#### Scenario: Associate action available for an already-matched occurrence
- **WHEN** the sidepanel displays an occurrence already matched to a movie or TV show
- **THEN** the "Associate" action SHALL still be enabled, allowing the user to correct or reassociate it

#### Scenario: Associate action available for an occurrence mid-download or already downloaded
- **WHEN** the sidepanel displays an occurrence that is downloading or already downloaded
- **THEN** the "Associate" action SHALL remain enabled, unlike the "Forcer le téléchargement" action which becomes disabled in this state

#### Scenario: Associate action available from both entry points
- **WHEN** the sidepanel is opened either from the Playlist tab or from the Radarr/Sonarr tab
- **THEN** the "Associate" action SHALL be available and functional in both contexts
