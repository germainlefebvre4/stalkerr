# Spec Delta

## ADDED Requirements

### Requirement: Filter bar groups controls by concern on Films and Séries
On both the Films and the Séries sections, the title search field SHALL sit alone on its own line, and below it the filter controls SHALL be arranged in labelled groups: a Radarr/Sonarr group (Suivi, Fichier, "Manquants") and a Playlist group (match-status, labelled "Correspondance"). All controls in the bar SHALL share the same height and visual style.

#### Scenario: Search field on its own line
- **WHEN** the user opens the Films or the Séries section
- **THEN** the search field SHALL be displayed alone on the first line of the filter area, with the filter groups on the line(s) below

#### Scenario: Groups are labelled
- **WHEN** the filter bar is displayed
- **THEN** the Suivi, Fichier and "Manquants" controls SHALL appear together under a Radarr/Sonarr group label, and the match-status control SHALL appear under a separate Playlist group label with its own "Correspondance" label

#### Scenario: Controls stay on one row when space allows
- **WHEN** the viewport is wide enough to hold the group controls
- **THEN** no dropdown SHALL stretch to a full row of its own

#### Scenario: Mobile layout
- **WHEN** the section is displayed on a mobile viewport
- **THEN** the search field SHALL take the full width and the filter controls SHALL wrap into a two-column layout without horizontal page scroll

#### Scenario: Filtering behavior is unchanged
- **WHEN** the user changes any control of the reorganised bar
- **THEN** the filtering semantics, defaults and URL persistence SHALL be identical to those before the reorganisation

### Requirement: Réinitialiser control restores the default filters
Both the Films and the Séries sections SHALL offer a "Réinitialiser" control in the filter bar that restores that section's Suivi to Surveillé only, Fichier to no selection, and match-status to "Tous", in a single update, and returns the table to its first page. It SHALL NOT alter the search field. It SHALL be disabled while the three filters already hold their default values.

#### Scenario: Resetting after changing filters
- **WHEN** the user has changed Suivi, Fichier or match-status in a section and clicks "Réinitialiser"
- **THEN** that section's Suivi SHALL be Surveillé only, Fichier SHALL have no selection, match-status SHALL be "Tous", and the table SHALL show the first page of the unfiltered-by-Fichier default view

#### Scenario: Reset is disabled at defaults
- **WHEN** a section's three filters already hold their default values
- **THEN** its "Réinitialiser" control SHALL be disabled

#### Scenario: Reset keeps the search term
- **WHEN** the user has typed a search term and clicks "Réinitialiser"
- **THEN** the search field SHALL keep its content and the search SHALL stay applied

#### Scenario: Reset is scoped to its own section
- **WHEN** the user clicks "Réinitialiser" in the Films section
- **THEN** the Séries section's filters SHALL be left unchanged

#### Scenario: Reset persists in the URL
- **WHEN** the user clicks "Réinitialiser" and then refreshes the page
- **THEN** the default filters SHALL remain applied after the page reloads
