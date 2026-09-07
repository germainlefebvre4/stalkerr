## 1. Header-only fixes (already-conformant requirements, missing top-level headers)

- [x] 1.1 In `openspec/specs/radarr-movie-matching/spec.md`, replace the opening `## ADDED Requirements` with `## Purpose` (a real 1-3 sentence paragraph on TVDB-ID-based movie matching) followed by `## Requirements`; requirements/scenarios below are untouched. Verify with `openspec validate radarr-movie-matching --type spec --strict` (no issues).
- [x] 1.2 In `openspec/specs/radarr-pagination/spec.md`, add a `## Purpose` paragraph and a `## Requirements` header before the existing `### Requirement:` blocks; requirements/scenarios untouched. Verify with `openspec validate radarr-pagination --type spec --strict` (no issues).
- [x] 1.3 In `openspec/specs/sonarr-pagination/spec.md`, add a `## Purpose` paragraph and a `## Requirements` header before the existing `### Requirement:` blocks; requirements/scenarios untouched. Verify with `openspec validate sonarr-pagination --type spec --strict` (no issues).
- [x] 1.4 In `openspec/specs/tmdb-rate-limiter/spec.md`, add a `## Purpose` paragraph and a `## Requirements` header before the existing `### Requirement:` blocks; requirements/scenarios untouched. Verify with `openspec validate tmdb-rate-limiter --type spec --strict` (no issues).

## 2. Purpose-only fixes (replace the "TBD - created by archiving/syncing ..." placeholder)

For each spec below: read its existing `## Requirements` section, then replace the placeholder `## Purpose` body with a real 1-3 sentence purpose statement grounded in those requirements. Requirements/scenarios stay untouched. Verify each with `openspec validate <name> --type spec --strict` (no issues).

- [x] 2.1 `openspec/specs/api-downloads-monitoring/spec.md`
- [x] 2.2 `openspec/specs/api-media-management/spec.md`
- [x] 2.3 `openspec/specs/api-processing-logs/spec.md`
- [x] 2.4 `openspec/specs/api-server-deployment/spec.md`
- [x] 2.5 `openspec/specs/configuration-management/spec.md`
- [x] 2.6 `openspec/specs/database-integration/spec.md`
- [x] 2.7 `openspec/specs/db-prune/spec.md`
- [x] 2.8 `openspec/specs/frontend-filters-management/spec.md`
- [x] 2.9 `openspec/specs/frontend-ihm-dashboard/spec.md`
- [x] 2.10 `openspec/specs/helm-chart-structure/spec.md`
- [x] 2.11 `openspec/specs/ingress-configuration/spec.md`
- [x] 2.12 `openspec/specs/local-dev/spec.md`
- [x] 2.13 `openspec/specs/m3u-playlist-details-sidepanel/spec.md`
- [x] 2.14 `openspec/specs/media-reset/spec.md`
- [x] 2.15 `openspec/specs/scheduled-jobs/spec.md`
- [x] 2.16 `openspec/specs/storage-configuration/spec.md`
- [x] 2.17 `openspec/specs/tmdb-manual-override/spec.md`

## 3. Partial reformat (some requirements untouched, legacy sections converted)

- [x] 3.1 In `openspec/specs/downloads-display-ui/spec.md`, insert the two Requirement blocks from this change's `specs/downloads-display-ui/spec.md` (Downloads Filter Option Sets; Graceful Handling of Missing Enrichment Fields), then delete the legacy "### Filter Requirements" and "### Non-Functional Requirements" sections they replace (moving the TS-interface-matching and HTML-select implementation notes into the existing "## Implementation Details" section instead of dropping them). Leave the 3 already-conformant requirements and everything below "## TypeScript Interfaces" untouched. Verify with `openspec validate downloads-display-ui --type spec --strict` (no issues).
- [x] 3.2 In `openspec/specs/downloads-enrichment-api/spec.md`, insert the 7 Requirement blocks from this change's `specs/downloads-enrichment-api/spec.md`, then delete the legacy "### Functional Requirements", "### Problem Filters", and "### Non-Functional Requirements" sections they replace (moving the GORM-consistency and pagination-pattern implementation notes into the existing "## Implementation Details" section instead of dropping them). Add a `## Purpose` paragraph (this spec currently opens with `## Description` instead). Leave the 4 already-conformant requirements (Multi-Value Problem Filter, Problem Filter Pagination Accuracy, Rename Target Folder Name, Target and Staging Path Fields) and everything from "## API Contract" onward untouched. Verify with `openspec validate downloads-enrichment-api --type spec --strict` (no issues).

## 4. Full rebuild

- [x] 4.1 Rewrite `openspec/specs/file-metadata-parsing/spec.md`: add a `## Purpose` paragraph, then replace the "## Requirements" body with the 7 Requirement blocks from this change's `specs/file-metadata-parsing/spec.md`. Move "no external dependencies (pure stdlib)" and the "≥ 95% coverage" goal out of Requirements into the existing "## Testing Requirements" / "## Implementation Notes" sections. Keep "## Examples", "## Implementation Notes", and "## Testing Requirements" as reference material below the requirements (already consistent with the new scenarios). Verify with `openspec validate file-metadata-parsing --type spec --strict` (no issues).

## 5. Final verification

- [x] 5.1 Run `openspec validate --all --strict` and confirm all 23 specs listed in this change's proposal now pass (the only remaining failure, if any, should be the unrelated `change/fix-year-mismatch-embedded-title-year` stub, which is explicitly out of scope for this change).
