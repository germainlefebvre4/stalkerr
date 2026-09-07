## Context

See `proposal.md` for motivation. This is a docs-only change: 23 spec files under `openspec/specs/` need fixing so `openspec validate --all --strict` stops failing on them, but no described behavior changes. `openspec`'s own tooling constrains *how* that fix can be authored:

- A change's delta spec (`specs/<capability>/spec.md` under this change) is only for `## ADDED/MODIFIED/REMOVED/RENAMED Requirements` — i.e. requirement/scenario content. Per `openspec instructions specs`, a delta for an **existing** capability must never carry `## Purpose` — the main spec already has one, and a delta's is silently ignored. Any Purpose fix (including a leftover `TBD` placeholder) has to be an edit to `<root>/openspec/specs/<capability-path>/spec.md` directly.
- The three specs that also need actual restructuring (free-form prose becoming `### Requirement:` / `#### Scenario:` blocks) have those legacy sections named things like "Functional Requirements" or "Filter Requirements" — not a `### Requirement: <name>` the delta grammar can target with `MODIFIED`/`REMOVED`. There is no delta operation that deletes an arbitrary untracked heading. Removing the superseded prose can only happen as a direct edit to the main spec too.

So this change has no code to implement in the usual sense — "implementation" is editing `openspec/specs/**/spec.md` directly for every one of the 23 capabilities. The 3 delta files already written under this change (`downloads-display-ui`, `downloads-enrichment-api`, `file-metadata-parsing`) are the authored, reviewed source of the new Requirement/Scenario content for those three specs; tasks.md treats them as "copy this into the main spec, then delete the prose it replaces" rather than relying on `openspec archive`'s automatic delta merge.

## Goals / Non-Goals

**Goals:**
- `openspec validate --all --strict` passes on all 23 targeted specs with zero remaining issues.
- Every scenario/example already documented (GIVEN/WHEN/THEN prose, worked `Examples`, bullet lists) survives the rewrite as an equivalent `#### Scenario:` — no behavior silently dropped.
- Already-conformant requirements in partially-legacy specs (`downloads-display-ui`, `downloads-enrichment-api`) are left untouched, not rewritten wholesale.

**Non-Goals:**
- No production code, API, schema, or config changes.
- No new capabilities and no requirement behavior changes — only how existing behavior is written down.
- Not attempting to also fix `openspec validate`'s other unrelated finding, the empty `fix-year-mismatch-embedded-title-year` change stub — that is a separate change, out of scope here.

## Decisions

**1. Direct main-spec edits for Purpose-only and header-only fixes (20 of 23 specs), not delta files.**
17 specs need only their `## Purpose` placeholder replaced; 4 specs (`radarr-movie-matching`, `radarr-pagination`, `sonarr-pagination`, `tmdb-rate-limiter`) need only `## Purpose` + a `## Requirements` header added (one of them currently opens with a stray delta-style `## ADDED Requirements` header that must become `## Requirements`). None of these 20 have any requirement/scenario change, so a delta file would have no valid `ADDED`/`MODIFIED` content to carry ("do not invent a requirement just to satisfy validation") — these are edited directly on `openspec/specs/<capability>/spec.md`.

**2. Delta files only where real Requirement/Scenario content is being created**, i.e. formalizing previously-informal prose into new `### Requirement:` blocks. Per the `MODIFIED` pitfall guidance ("if adding new concerns without changing existing behavior, use ADDED instead"), all 3 are written as `## ADDED Requirements` — they add requirement/scenario structure that did not exist in that form, without changing what the system does.

**3. Non-behavioral bullets are relocated, not turned into fake requirements.** Items like "unit test coverage ≥ 95%", "no external dependencies (pure stdlib)", or "handler SHALL use GORM patterns consistent with the codebase" are implementation/process concerns, not externally observable behavior. Where the target spec already has a non-Requirements section for this (`file-metadata-parsing` has `## Testing Requirements` / `## Implementation Notes`; `downloads-enrichment-api` has `## Implementation Details`), the direct-edit task moves these bullets there instead of inventing a Requirement/Scenario for them. Genuinely observable constraints (response-time budget, concurrency safety, graceful handling of missing fields) do get a real Requirement.

**4. Execution order, cheapest/lowest-risk first:** header-only (4) → Purpose-only (17) → partial reformat (2: `downloads-display-ui`, `downloads-enrichment-api`) → full rebuild (1: `file-metadata-parsing`). Each spec is independently verifiable with `openspec validate <name> --type spec --strict` right after its own edit, so a mistake in one spec never blocks or gets confused with another.

## Risks / Trade-offs

- [Risk] Writing a Purpose paragraph without actually reading that capability's Requirements risks a generic or inaccurate Purpose → Mitigation: task per spec explicitly requires reading the existing Requirements section first; Purpose is 1-3 sentences of pure "what this is for", no new claims.
- [Risk] For the 3 reformatted specs, applying the new Requirement blocks but forgetting to delete the legacy prose section it replaces leaves the main spec with both the old freeform heading (still failing strict validation) and duplicated content → Mitigation: each reformat task is a single edit that both inserts the new blocks and removes the exact legacy section, verified by re-running `openspec validate --strict` on that one spec immediately after.
- [Risk] `downloads-enrichment-api` and `downloads-display-ui` have large non-Requirements sections (API contract, TypeScript interfaces, ASCII mockups, implementation notes) below/after the parts being touched → Mitigation: tasks reference the exact legacy section names to replace; everything else in the file is explicitly out of scope for this change.
- [Risk] `file-metadata-parsing`'s full rebuild is the highest-effort item and the only one with zero pre-existing Requirement/Scenario structure to anchor against → Mitigation: the delta already written under this change (`specs/file-metadata-parsing/spec.md`) was derived line-by-line from the current prose and its 5 worked `Examples`, so applying it is a direct copy rather than a fresh rewrite at task time.

## Migration Plan

No rollback concerns beyond normal git revert (docs-only, no runtime impact). Order: apply the direct edits and the 3 delta-derived rewrites to `openspec/specs/**/spec.md` per tasks.md, validating each spec individually as it's fixed; once all 23 are fixed, run `openspec validate --all --strict` for the final confirmation; then archive this change as usual.
