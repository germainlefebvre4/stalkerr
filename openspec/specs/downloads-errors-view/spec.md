# Downloads Errors View Specification

## Purpose

Provides a desktop-only "Erreurs" tab that surfaces every downloaded item whose organized file has a naming problem — invalid extension, missing year, or a year inconsistent with its matched TMDB year — in one triageable list, separate from the general Downloads tab.

## Requirements

### Requirement: Errors Tab Desktop-Only Visibility
The frontend SHALL add an "Erreurs" tab to the tab navigation, rendered only on viewports at or above the `768px` mobile breakpoint (capability `frontend-responsive-layout`). Below that breakpoint, the tab SHALL NOT appear in the desktop segmented tabs and SHALL NOT be added to the mobile bottom tab bar. If the "Erreurs" tab is the active tab and the viewport narrows below the breakpoint, the frontend SHALL fall back to another visible tab rather than rendering the Erreurs tab or a related empty/broken layout.

#### Scenario: Tab visible on desktop
- **WHEN** the viewport width is at or above `768px`
- **THEN** the frontend SHALL show an "Erreurs" entry among the tabs, selectable like the existing tabs

#### Scenario: Tab hidden on mobile
- **WHEN** the viewport width is below `768px`
- **THEN** the frontend SHALL NOT show an "Erreurs" entry in the segmented tabs or in the mobile bottom tab bar

#### Scenario: Narrowing the viewport while Erreurs is active falls back
- **WHEN** the "Erreurs" tab is currently active and the viewport is resized to below `768px`
- **THEN** the frontend SHALL switch the active tab to another available tab instead of continuing to render the Erreurs tab

### Requirement: Errors Table Lists Problematic Downloads
The Erreurs tab SHALL fetch from `GET /api/v1/downloads` using a combined problem filter matching `missing_year`, `year_mismatch`, or `unknown_format` by default (capability `downloads-enrichment-api`), and SHALL render one table row per matching download. Each row SHALL show: the content type icon and title with year (falling back to the file name or folder name when no `content.title` is available, consistent with the existing Downloads tab fallback), one badge per applicable reason (an item with more than one problem SHALL show one badge per problem, not just the first), the detected year and the expected/matched year when a year-related reason applies, the detected file extension, the file's folder location (`file_info.folder_name`), and the download's completion date. The table SHALL NOT include `low_quality` results, since resolution is a quality concern rather than a naming/parsing error.

#### Scenario: Default view combines all three error reasons
- **WHEN** the Erreurs tab loads with no reason filter selected
- **THEN** the frontend SHALL request downloads matching `missing_year`, `year_mismatch`, or `unknown_format` and list every match, excluding `low_quality`-only results

#### Scenario: An item with multiple problems shows multiple badges
- **WHEN** a listed download has both `file_info.has_year_in_path` false and `file_info.is_valid_format` false
- **THEN** its row SHALL show both a "missing year" badge and an "invalid extension" badge

#### Scenario: Title falls back when content title is unavailable
- **WHEN** a listed download has no `content.title`
- **THEN** the row SHALL display the file name or folder name derived from `file_info`, instead of leaving the title blank

### Requirement: Errors Tab Reason Filter
The Erreurs tab SHALL offer a reason filter with the options "Tous", "Extension invalide", "Année manquante", and "Année incohérente". Selecting a specific reason SHALL re-fetch using that single `problem` value (`unknown_format`, `missing_year`, or `year_mismatch` respectively) and reset pagination to the first page. Selecting "Tous" SHALL restore the combined three-reason filter.

#### Scenario: Filtering to a single reason
- **WHEN** the user selects "Extension invalide" in the reason filter
- **THEN** the frontend SHALL re-fetch `GET /api/v1/downloads?problem=unknown_format` (plus any active pagination parameters reset to page 1) and show only those results

#### Scenario: Returning to the combined view
- **WHEN** the user selects "Tous" after having a specific reason selected
- **THEN** the frontend SHALL re-fetch using the combined `missing_year,year_mismatch,unknown_format` filter and reset to page 1

### Requirement: Errors Tab Pagination
The Erreurs tab SHALL paginate its list using the same `limit`, `offset`, `total`, and `total_pages` fields returned by `GET /api/v1/downloads`, with page navigation controls and an items-per-page selector, consistent with the existing Downloads tab pagination pattern.

#### Scenario: Navigating pages
- **WHEN** the user navigates to another page in the Erreurs tab
- **THEN** the frontend SHALL re-fetch with the corresponding `offset` under the currently active reason filter and update the list

### Requirement: Dedicated Diagnostic Sidepanel
Clicking a row in the Erreurs table SHALL open a sidepanel component dedicated to this tab (distinct from the Downloads tab's details sidepanel, capability `downloads-details-sidepanel`), which SHALL lead with a diagnostic section listing which reason(s) apply and their relevant values (e.g. detected vs. expected year, detected extension), followed by the file location detail, then the matched content detail (title, year, genres, season/episode when applicable), then the download's status and dates, then the source `url`. The sidepanel SHALL NOT expose the "Déplacer" or "Renommer" actions, or any other corrective action.

#### Scenario: Diagnostic section renders first
- **WHEN** the user clicks a row in the Erreurs table
- **THEN** the sidepanel SHALL open with the diagnostic reason(s) and their values as its first visible section, before file, content, and status detail

#### Scenario: No corrective actions are offered
- **WHEN** the sidepanel is open for any download, regardless of its status
- **THEN** the sidepanel SHALL NOT render "Déplacer", "Renommer", "Associer", "Forcer le téléchargement", or any other action control
