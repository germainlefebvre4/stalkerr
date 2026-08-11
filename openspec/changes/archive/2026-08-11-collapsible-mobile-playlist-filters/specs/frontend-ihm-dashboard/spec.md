## ADDED Requirements

### Requirement: Mobile Collapsible Advanced Filters
On mobile viewports, the frontend SHALL render the Playlist advanced filter grid (media name search, group search, TMDB enrichment, pipeline state) inside a disclosure section that is collapsed by default, so the table/card list below it is visible without additional scrolling. The content-type filter (`all`, `movies`, `tvshows`) SHALL remain outside this disclosure and always visible on mobile. The disclosure's header SHALL display a count of advanced filters currently set to a non-default value (i.e. not `all` and not empty). The disclosure's collapsed/expanded state SHALL NOT be persisted across page loads or reflected in the URL or `localStorage`; it SHALL always start collapsed when the Playlist tab is mounted on a mobile viewport, regardless of whether an advanced filter is already active from a restored URL. On non-mobile viewports, the advanced filter grid SHALL continue to render inline and always visible, with no disclosure control.

#### Scenario: Advanced filters are collapsed on mobile by default
- **WHEN** the user opens the Playlist tab on a mobile viewport with no active advanced filters
- **THEN** the frontend SHALL render the content-type filter always visible, render the advanced filter grid collapsed inside a disclosure showing no active-filter count, and render the table/card list immediately below without requiring the user to scroll past the advanced filter grid

#### Scenario: Expanding the disclosure reveals the advanced filter grid
- **WHEN** the user taps the collapsed advanced-filters disclosure header on a mobile viewport
- **THEN** the frontend SHALL expand the disclosure to show the media name search, group search, TMDB enrichment, and pipeline state controls, without changing the content-type filter's visibility or position

#### Scenario: Active filter count is visible while collapsed
- **WHEN** the pipeline state filter is set to `failed` and the TMDB enrichment filter is set to `yes`, and the advanced-filters disclosure is collapsed
- **THEN** the frontend SHALL display a count of `2` active advanced filters on the disclosure header

#### Scenario: Disclosure state resets on remount regardless of restored filters
- **WHEN** the user navigates to the Playlist tab on a mobile viewport with a pipeline state filter of `failed` restored from the URL query string
- **THEN** the frontend SHALL still render the advanced-filters disclosure collapsed by default, while the restored `failed` filter SHALL remain applied to the fetched playlist items and reflected in the disclosure's active-filter count

#### Scenario: Desktop rendering is unaffected
- **WHEN** the user opens the Playlist tab on a non-mobile viewport
- **THEN** the frontend SHALL render the advanced filter grid inline and fully visible, with no disclosure control and no collapse/expand behavior
