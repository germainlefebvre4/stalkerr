# Spec Delta

## ADDED Requirements

### Requirement: Suivi and Fichier filter controls on the Films section
The Radarr sub-tab SHALL provide two independent multi-select filter controls, Suivi (Surveillé, Non surveillé) and Fichier (Téléchargé, Absent), that filter the Films table across the entire catalog before pagination, separately from the match-status filter. Values within one control combine as OR, the two controls combine as AND. Suivi SHALL default to Surveillé and Fichier to no selection (no restriction). Both selections SHALL persist across a page refresh via the URL.

#### Scenario: Default selection preserves today's view
- **WHEN** the user opens the Films section without having chosen any filter before
- **THEN** only Surveillé SHALL be selected in Suivi, Fichier SHALL have no selection, and the table SHALL show the same movies it showed before this change

#### Scenario: Filtering on downloaded movies
- **WHEN** the user selects only "Téléchargé" in Fichier
- **THEN** the table SHALL show only movies for which Radarr already has a file, re-paginated from the first page

#### Scenario: Both Suivi values selected
- **WHEN** the user selects both Surveillé and Non surveillé
- **THEN** the table SHALL show monitored and unmonitored movies alike

#### Scenario: Unmonitored movies that are downloaded
- **WHEN** the user selects Non surveillé in Suivi and Téléchargé in Fichier
- **THEN** the table SHALL show only unmonitored movies that Radarr already has a file for

#### Scenario: Filters combine with the match-status filter
- **WHEN** a Suivi/Fichier selection and a match-status selection are both active
- **THEN** the table SHALL show only movies satisfying all of them

#### Scenario: Selection persists across a page refresh
- **WHEN** the user changes the Suivi or Fichier selection and then refreshes the page
- **THEN** the same selections SHALL remain applied after the page reloads

#### Scenario: Obsolete URL value falls back to the default
- **WHEN** the page is opened with a URL carrying the retired `missing` value in the Suivi selection
- **THEN** the control SHALL fall back to its default (Surveillé) instead of failing

### Requirement: Suivi and Fichier filter controls on the Séries section
The Sonarr sub-tab SHALL provide two independent multi-select filter controls, Suivi (Surveillé, Non surveillé) and Fichier (Téléchargé, Partiel, Absent), that filter the Séries table across the entire catalog before pagination, separately from the match-status filter. Values within one control combine as OR, the two controls combine as AND. Suivi SHALL default to Surveillé and Fichier to no selection. Both selections SHALL persist across a page refresh via the URL.

#### Scenario: Default selection preserves today's view
- **WHEN** the user opens the Séries section without having chosen any filter before
- **THEN** only Surveillé SHALL be selected in Suivi, Fichier SHALL have no selection, and the table SHALL show the same series it showed before this change

#### Scenario: Filtering on partially downloaded series
- **WHEN** the user selects only "Partiel" in Fichier
- **THEN** the table SHALL show only series with some but not all aired episodes downloaded, re-paginated from the first page

#### Scenario: Several Fichier values
- **WHEN** the user selects "Partiel" and "Absent" in Fichier
- **THEN** the table SHALL show series whose file state is either Partiel or Absent

#### Scenario: Filters combine with the match-status filter
- **WHEN** a Suivi/Fichier selection and a match-status selection are both active
- **THEN** the table SHALL show only series satisfying all of them

#### Scenario: Selection persists across a page refresh
- **WHEN** the user changes the Suivi or Fichier selection and then refreshes the page
- **THEN** the same selections SHALL remain applied after the page reloads

#### Scenario: Obsolete URL value falls back to the default
- **WHEN** the page is opened with a URL carrying the retired `missing` value in the Suivi selection
- **THEN** the control SHALL fall back to its default (Surveillé) instead of failing

### Requirement: Manquants shortcut on both filter groups
Both the Films and Séries sections SHALL offer a one-click "Manquants" shortcut next to their Suivi and Fichier controls that sets Suivi to Surveillé only and Fichier to Absent and Partiel (Absent only for Films), replacing the current selections.

#### Scenario: Activating the shortcut on Séries
- **WHEN** the user clicks "Manquants" in the Séries section
- **THEN** Suivi SHALL be set to Surveillé only, Fichier to Partiel and Absent, and the table SHALL show monitored series that are not fully downloaded, re-paginated from the first page

#### Scenario: Activating the shortcut on Films
- **WHEN** the user clicks "Manquants" in the Films section
- **THEN** Suivi SHALL be set to Surveillé only, Fichier to Absent, and the table SHALL show monitored movies Radarr has no file for

#### Scenario: Shortcut reflects the current selection
- **WHEN** the current selections equal the shortcut's selections
- **THEN** the shortcut SHALL appear active

### Requirement: Fichier badge uses an actionability-aware color
Wherever a movie, series, or episode shows its Fichier state, the badge SHALL read Téléchargé, Partiel, or Absent. Téléchargé SHALL be green and Partiel orange. Absent SHALL be red when the item is monitored and neutral grey when it is not. A series Partiel badge SHALL also show the downloaded/aired ratio (e.g. 18/28).

#### Scenario: Absent and monitored
- **WHEN** a monitored movie has no file
- **THEN** its Fichier badge SHALL read Absent and be red

#### Scenario: Absent and unmonitored
- **WHEN** an unmonitored movie has no file
- **THEN** its Fichier badge SHALL read Absent and be neutral grey

#### Scenario: Series partially downloaded
- **WHEN** a series has 18 of 28 aired episodes downloaded
- **THEN** its Fichier badge SHALL read Partiel, be orange, and show 18/28

#### Scenario: Downloaded item
- **WHEN** an item's file state is downloaded
- **THEN** its Fichier badge SHALL read Téléchargé and be green, whether or not it is monitored

### Requirement: Sidepanel header shows the selected item's Suivi and Fichier
The sidepanel opened for a selected movie or series SHALL display that item's Suivi (Surveillé or Non surveillé) and Fichier badge next to its title/year header, using the same badge conventions as the table.

#### Scenario: Sidepanel header for a movie
- **WHEN** the user selects a movie row and its sidepanel opens
- **THEN** the header SHALL display that movie's Suivi and Fichier badges alongside its title and year

#### Scenario: Sidepanel header for a series
- **WHEN** the user selects a series row and its sidepanel opens
- **THEN** the header SHALL display that series' Suivi and Fichier badges, with the ratio when Partiel, alongside its title and year

### Requirement: Episode rows display their own Suivi and Fichier
Each monitored-episode row in the Séries sidepanel's per-episode breakdown SHALL display its own Suivi and Fichier (Téléchargé or Absent) from Sonarr's per-episode state, in addition to its existing matched/unmatched playlist badge.

#### Scenario: Missing episode is distinguishable
- **WHEN** an episode row is shown for an episode monitored in Sonarr without a file there
- **THEN** the row SHALL display a red Absent Fichier badge distinct from its matched/unmatched playlist badge

#### Scenario: Downloaded episode
- **WHEN** an episode row is shown for an episode that has a file in Sonarr
- **THEN** the row SHALL display a green Téléchargé Fichier badge

### Requirement: Suivi and Fichier badges render as compact status dots on mobile viewports
On mobile-width viewports, every Suivi and Fichier badge (table row, mobile card, sidepanel header, episode row) SHALL render as a small colored status dot with a short label, following the compact-indicator convention already used by this tab's mobile summary cards, instead of the full-size text badge.

#### Scenario: Mobile card shows two compact indicators
- **WHEN** the Films or Séries section is displayed on a mobile-width viewport
- **THEN** each mobile list card SHALL show Suivi and Fichier as two compact colored dots with short labels, alongside its matched/unmatched badge, without overflowing or wrapping awkwardly

#### Scenario: Mobile episode row shows compact indicators
- **WHEN** the Séries sidepanel's per-episode breakdown is displayed on a mobile-width viewport
- **THEN** each episode row SHALL show its Suivi and Fichier as compact colored dots with short labels, alongside its matched/unmatched badge

### Requirement: Films section lists Radarr movies with match status, Suivi and Fichier
The tab SHALL include a Films section listing Radarr movies matching the active Suivi and Fichier filters (Surveillé and no Fichier restriction, by default - see "Suivi and Fichier filter controls on the Films section"), each showing at minimum its title, year, whether it has a matching entry in the local playlist, its Suivi (Surveillé or Non surveillé) and its Fichier (Téléchargé or Absent), with pagination controls.

#### Scenario: Matched vs unmatched movies are visually distinguishable
- **WHEN** the Films section is displayed
- **THEN** movies with a playlist match SHALL be visually distinguished from movies without one (e.g. a status badge)

#### Scenario: Suivi and Fichier are shown separately from the match badge
- **WHEN** the Films section is displayed
- **THEN** a movie's Suivi and Fichier SHALL each be shown as their own indicator (a Suivi column and a Fichier badge), distinct from the matched/unmatched badge and from each other

#### Scenario: Paginating the Films section
- **WHEN** the user navigates to another page in the Films section
- **THEN** the section SHALL request and display that page's movies with their match status

### Requirement: Séries section lists Sonarr series with aggregate match status, Suivi and Fichier
The tab SHALL include a Séries section listing Sonarr series matching the active Suivi and Fichier filters (Surveillé and no Fichier restriction, by default - see "Suivi and Fichier filter controls on the Séries section"), each showing at minimum its title, an aggregate ratio of matched vs. monitored episodes (e.g. "8/12"), its Suivi, and its Fichier (Téléchargé, Partiel or Absent), with pagination controls. Per-episode detail is not shown in this list view.

#### Scenario: Series aggregate ratio is displayed
- **WHEN** the Séries section is displayed
- **THEN** each series SHALL show the count of monitored episodes matched in the playlist over the total count of monitored episodes

#### Scenario: Fully matched series is distinguishable from partially or unmatched series
- **WHEN** a series has all monitored episodes matched, some matched, or none matched
- **THEN** the three cases SHALL be visually distinguishable from one another

#### Scenario: Suivi and Fichier are shown separately from the matched ratio
- **WHEN** the Séries section is displayed
- **THEN** a series' Suivi and Fichier SHALL each be shown as their own indicator, distinct from the matched-ratio badge and from each other

## REMOVED Requirements

### Requirement: Séries section lists Sonarr monitored series with aggregate match status
**Reason**: Lists now follow the Suivi and Fichier filters and show both badges instead of the single État.
**Migration**: See "Séries section lists Sonarr series with aggregate match status, Suivi and Fichier".

### Requirement: Films section lists Radarr monitored movies with match status
**Reason**: Lists now follow the Suivi and Fichier filters and show both badges instead of the single État.
**Migration**: See "Films section lists Radarr movies with match status, Suivi and Fichier".

### Requirement: État status filter control on the Films section
**Reason**: The single État axis (Monitored/Unmonitored/Missing) conflated monitoring intent with file presence; replaced by independent Suivi and Fichier controls.
**Migration**: Use "Suivi and Fichier filter controls on the Films section"; the former Missing view is the "Manquants" shortcut (Surveillé + Absent).

### Requirement: État status filter control on the Séries section
**Reason**: The single État axis (Monitored/Unmonitored/Missing) conflated monitoring intent with file presence; replaced by independent Suivi and Fichier controls.
**Migration**: Use "Suivi and Fichier filter controls on the Séries section"; the former Missing view is the "Manquants" shortcut (Surveillé + Absent/Partiel).

### Requirement: Sidepanel header shows the selected movie's or series' État
**Reason**: État is split into Suivi and Fichier.
**Migration**: See "Sidepanel header shows the selected item's Suivi and Fichier".

### Requirement: Episode rows display their own État alongside their matched/unmatched status
**Reason**: État is split into Suivi and Fichier.
**Migration**: See "Episode rows display their own Suivi and Fichier".

### Requirement: État badge renders as a compact status dot on mobile viewports
**Reason**: The État badge no longer exists; its mobile compact rendering moves to the Suivi and Fichier badges.
**Migration**: See "Suivi and Fichier badges render as compact status dots on mobile viewports".
