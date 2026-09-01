# playlist-pipeline-state-display Specification

## Purpose

Defines how the Playlist page's desktop table and mobile list card display an item's processing status (enrichment) and download status independently of one another, so neither fact is hidden or overwritten by the other.

## Requirements

### Requirement: Independent Processing and Download Status Indicators
The Playlist page SHALL display an item's processing status (`pending`/`processed`) and its download status (`not_downloaded`/`downloading`/`organizing`/`downloaded`/`failed`) as two independent visual indicators, derived from the item's pipeline state, such that a download outcome (in progress, completed, or failed) never hides or overwrites the fact that the item was already successfully processed.

On viewports at or above the mobile breakpoint, each indicator SHALL render as a distinct compact text badge. Below the mobile breakpoint, each indicator SHALL render as a small colored icon accompanied by an accessible text label (`title`/`aria-label`) carrying the full status name, so the mobile list card stays compact while remaining screen-reader accessible.

An item that has been processed but for which no download has been attempted yet (no download record exists) SHALL show a neutral/gray "not downloaded" download-status indicator, visually distinct from both the "downloaded" (success) and "failed" (error) indicators.

#### Scenario: Desktop table shows two independent badges
- **WHEN** the Playlist table (desktop) renders a row for an item that has been processed and whose forced download subsequently failed
- **THEN** the row SHALL display a "processed" processing-status badge and, separately, a "failed" download-status badge, both visible at the same time

#### Scenario: Mobile list card shows two independent icon indicators
- **WHEN** the Playlist mobile list card renders an item that has been processed and whose forced download subsequently failed
- **THEN** the card SHALL display two small icon indicators — one for processing status, one for download status — each with its own color and an accessible label naming the full status, instead of a single combined text badge

#### Scenario: Failed download does not hide successful processing
- **WHEN** an item's processing succeeded (matched to a movie or TV show) and a later download attempt for that item fails
- **THEN** the Playlist page SHALL continue to show the processing-status indicator as "processed", independently of the download-status indicator showing "failed"

#### Scenario: Processed item with no download attempted shows a neutral download indicator
- **WHEN** an item has been processed but no download has ever been attempted for it
- **THEN** the download-status indicator SHALL render in a neutral/gray style distinct from "downloaded" and "failed", rather than being blank or omitted
