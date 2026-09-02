## MODIFIED Requirements

### Requirement: Selecting an occurrence opens the full media detail drawer
Selecting a playlist occurrence row (a movie's occurrence, or one of a series episode's occurrences after expanding it) SHALL open the same media detail drawer used by the Playlist tab for that occurrence - TMDB metadata, pipeline state, M3U provenance, raw line and stream URL, the "Associate" action, and the force-download action - rather than only the resolution/state summary shown in the occurrences list.

#### Scenario: Opening a movie occurrence's detail
- **WHEN** the user selects an occurrence row in a matched movie's sidepanel
- **THEN** the full media detail drawer SHALL open for that occurrence, including its "Associate" action and its force-download action if eligible

#### Scenario: Detail drawer stacks beside the occurrences sidepanel on desktop
- **WHEN** the media detail drawer is opened from the occurrences sidepanel on a desktop-width viewport
- **THEN** the detail drawer SHALL appear as an additional panel positioned immediately to the left of the occurrences sidepanel, with both panels visible and the occurrences sidepanel unchanged

#### Scenario: Detail view replaces the sidepanel on mobile
- **WHEN** the media detail drawer is opened from the occurrences sidepanel on a mobile-width viewport
- **THEN** the detail view SHALL replace the occurrences sidepanel's content in place, and the user SHALL be able to return to the occurrences list from it

#### Scenario: Associate action available from the Radarr/Sonarr detail drawer
- **WHEN** the user opens the media detail drawer from the Radarr/Sonarr tab's occurrences sidepanel
- **THEN** the "Associate" action SHALL be present and functional, letting the user correct or confirm the occurrence's TMDB match without leaving the Radarr/Sonarr tab
