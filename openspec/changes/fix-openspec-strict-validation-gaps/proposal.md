## Why

`openspec validate --all --strict` currently fails on 24 of 54 items (46 pass without `--strict`). Two distinct gaps cause this: six capability specs still use a pre-strict free-form format (no `## Purpose`, requirements written as prose instead of `### Requirement:` / `#### Scenario:` blocks, in one case zero requirements/scenarios at all), and seventeen capability specs still carry the literal placeholder text `openspec archive`/`openspec sync` writes into `## Purpose` for a newly created capability ("TBD - created by archiving/syncing change X. Update Purpose after archive.") that nobody ever replaced. No product behavior is wrong or missing — only the spec documents describing it are non-compliant with the strict schema, which blocks `--strict` from being a reliable CI/pre-merge gate.

## What Changes

- Add a real `## Purpose` paragraph (replacing the `TBD - created by ...` placeholder) to the 17 capability specs where it was left unfilled after an `archive`/`sync` operation.
- Add a real `## Purpose` paragraph and a `## Requirements` header to 4 capability specs that already have fully-formed `### Requirement:` / `#### Scenario:` content but are missing these two top-level headers (one of them currently opens with a delta-style `## ADDED Requirements` header left over from its original change, which must become a plain `## Requirements` header).
- Convert the 2 non-conformant sections of `downloads-display-ui` ("Filter Requirements", "Non-Functional Requirements") into proper `### Requirement:` / `#### Scenario:` blocks; the rest of that spec is already conformant and untouched.
- Rewrite the free-form "Functional Requirements" / "Problem Filters" prose in `downloads-enrichment-api` into discrete `### Requirement:` / `#### Scenario:` blocks; its already-conformant requirements (Multi-Value Problem Filter, Problem Filter Pagination Accuracy, Rename Target Folder Name, Target and Staging Path Fields) are untouched.
- Rebuild `file-metadata-parsing` from scratch into `## Purpose` / `## Requirements` with `### Requirement:` / `#### Scenario:` blocks, deriving requirements and scenarios from its existing "Functional/Non-Functional Requirements" prose and its 5 worked "Examples".
- No production code, API, or runtime behavior changes. This is a documentation/spec-format-only change.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities

Purpose-only rewrite (placeholder text replaced with a real Purpose paragraph, requirements/scenarios untouched):
- `api-downloads-monitoring`, `api-media-management`, `api-processing-logs`, `api-server-deployment`, `configuration-management`, `database-integration`, `db-prune`, `frontend-filters-management`, `frontend-ihm-dashboard`, `helm-chart-structure`, `ingress-configuration`, `local-dev`, `m3u-playlist-details-sidepanel`, `media-reset`, `scheduled-jobs`, `storage-configuration`, `tmdb-manual-override`

Header-only fix (add `## Purpose` + `## Requirements`, requirements/scenarios untouched):
- `radarr-movie-matching`, `radarr-pagination`, `sonarr-pagination`, `tmdb-rate-limiter`

Partial reformat (some requirements untouched, some sections converted to proper Requirement/Scenario blocks):
- `downloads-display-ui`: convert "Filter Requirements" and "Non-Functional Requirements" sections
- `downloads-enrichment-api`: convert "Functional Requirements" and "Problem Filters" sections

Full rebuild into Purpose/Requirements/Scenario structure:
- `file-metadata-parsing`

## Impact

- Affected: 23 files under `openspec/specs/**/spec.md` (documentation only).
- Not affected: application code, APIs, database schema, deployment config, tests — the requirements these specs describe are not changing, only how they are written down.
- Unblocks: `openspec validate --all --strict` as a trustworthy pass/fail gate (e.g. for CI or pre-merge checks) going forward.
