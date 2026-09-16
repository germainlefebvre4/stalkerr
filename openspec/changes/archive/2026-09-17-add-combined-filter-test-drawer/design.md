## Context

See proposal.md - Why. Today, dry-run testing is duplicated across three call sites with independent state: `CreateFilterDialog` (inline results below its form, `frontend/src/components/CreateFilterDialog.tsx:230-238`), and one `FilterCardTester` instance per filter card in `FiltersSection` (`frontend/src/components/FiltersSection.tsx:30-103`), each owning its own source selection and dry-run state, rendering `FilterDryRunPanel` inline. The backend endpoint (`internal/api/filter_dryrun_handlers.go`) evaluates exactly one attribute (`group_title` or `tvg_name`) per request; the real ingestion pipeline only keeps a line that passes both (`internal/filter/filter.go`'s `Manager.ShouldProcess`/`MatchesItem`), which today's dry-run cannot reproduce.

The codebase already has a proven pattern for a shared, reusable side panel used from multiple embedding contexts: `MediaOccurrenceDrawer.tsx` splits `MediaOccurrenceDrawerBody` (pure content) from `MediaOccurrenceDrawer` (the `Dialog.Root` + `.drawer-content` shell), and `RunItemsDialog.tsx:41-61,125-130` shows the exact mechanics for mounting that drawer as a non-modal sibling of another still-open dialog (`modal={false}`, `withOverlay={false}`, and a `justClosedDrawerRef` to suppress the outside-click-closes-both-layers issue Radix's portal-based dismissable layer would otherwise cause).

## Goals / Non-Goals

**Goals:**
- One shared test drawer and one shared panel body, reused from both `CreateFilterDialog` and `FiltersSection`, replacing all three current inline testers.
- A new combined test mode that reproduces the real AND-logic between Group Title and TVG Name, with a cause-ventilated result.
- Preserve the presentational work already shipped by `improve-filter-dryrun-results-layout` (two-column top-values grid, height-capped scrollable search table) inside the new drawer body.

**Non-Goals:**
- No change to how filters are created, saved, or deleted (`Create/Delete Filter Configuration` requirements are untouched).
- No live/streaming re-test as patterns are typed - testing stays an explicit, user-triggered action (click "Tester"), same as today.
- No attempt to let the combined test use in-progress, unsaved patterns from `CreateFilterDialog` for the *other* attribute - it always uses the effective (override-or-origin) patterns for both attributes, per the confirmed design decision.

## Decisions

**Component split: `FilterTestPanelBody` + `FilterTestDrawer`, mirroring `MediaOccurrenceDrawerBody`/`MediaOccurrenceDrawer`.**
`FilterTestPanelBody` is pure content (no `Dialog.Root`): the M3U source selector, the "Tester" trigger, and the mode-dependent result display (summary, top-value columns, search field, results table). `FilterTestDrawer` is the `Dialog.Root` + `.drawer-content` shell, rendering a contextual title derived from the target and `<FilterTestPanelBody>` inside. This lets `CreateFilterDialog` and `FiltersSection` each mount the shell with different `modal`/`withOverlay` settings while sharing all the actual testing logic and layout - the same reason this split already exists for `MediaOccurrenceDrawer`.

**A single `target` prop drives both what is shown and whether the drawer is open**, instead of a separate open/closed boolean:
```ts
type FilterTestTarget =
  | { mode: 'single'; attribute: 'group_title' | 'tvg_name'; includePatterns: string; excludePatterns: string; label: string }
  | { mode: 'combined'; groupTitleInclude: string; groupTitleExclude: string; tvgNameInclude: string; tvgNameExclude: string; label: string };
```
The drawer is open exactly when `target !== null` (mirrors `MediaOccurrenceDrawer`'s `item: PlaylistItem | null` prop). Closing the drawer (X, Escape, overlay click) sets `target` back to `null` in the owning component, which discards the in-progress/last dry-run result - matching the confirmed answer that this is acceptable. This refines the earlier "open iff status !== idle" idea: since source selection now lives inside the drawer (decision below), the drawer must already carry a `target` to know what to render and title, so reusing that same value as the open/closed signal avoids a second, redundant boolean while keeping the same "no extra state" property that was approved.

**Source selector moves into the drawer; each trigger site only supplies a `target`.**
- `CreateFilterDialog` keeps its own local `testTarget: FilterTestTarget | null` (always `mode: 'single'`), set when "Tester" is clicked, cleared by the existing `wasOpen` reset effect when the dialog itself closes (`CreateFilterDialog.tsx:38-52`). It mounts `<FilterTestDrawer modal={false} withOverlay={false} ...>` as a sibling of its own `Dialog.Content`, with the same `justClosedDrawerRef` suppression as `RunItemsDialog`.
- `FiltersSection` lifts a single `testTarget: FilterTestTarget | null` for the whole section (replacing each card's own `FilterCardTester` state). A card's "Tester" button becomes a plain trigger that calls `setTestTarget({mode: 'single', ...})`; a new "Tester l'ensemble" button next to "Configurer un filtre" resolves the effective patterns for both attributes (override if `filters.find(...)` exists, else `filterOrigin.find(...)`, exactly the logic already inlined at `FiltersSection.tsx:176-177,205-206`) and calls `setTestTarget({mode: 'combined', ...})`. It mounts exactly one `<FilterTestDrawer modal ...>` for the section.

**Backend request generalizes via an explicit `attributes` list, not inference from non-empty patterns.**
Inferring "which attribute(s) are being tested" purely from whether their pattern fields are non-empty would misfire: testing `group_title` with a deliberately empty include/exclude (a valid "what does unfiltered group_title look like" test, allowed today) is indistinguishable from "no attribute selected." Instead, the request explicitly lists which attribute(s) are in scope:
```go
type FilterDryRunRequest struct {
    SourceName        string   `json:"source_name" binding:"required"`
    Attributes        []string `json:"attributes" binding:"required"` // ["group_title"], ["tvg_name"], or both
    GroupTitleInclude  string  `json:"group_title_include_patterns"`
    GroupTitleExclude  string  `json:"group_title_exclude_patterns"`
    TvgNameInclude     string  `json:"tvg_name_include_patterns"`
    TvgNameExclude     string  `json:"tvg_name_exclude_patterns"`
    Search             string  `json:"search"`
    SearchAttribute    string  `json:"search_attribute"` // required when len(Attributes) == 2 and Search != ""
}
```
`len(Attributes) == 1` reproduces today's single-attribute request/response exactly (the other attribute imposes no filtering). `len(Attributes) == 2` selects combined mode. This is an internal-only endpoint (no external consumers), and frontend + backend ship together, so the request/response shape can change directly without versioning.

**Combined response is a distinct shape, not an overloaded single-attribute one**, so existing single-attribute callers' response parsing is untouched:
```go
// unchanged: FilterDryRunSummaryResponse{NoArchive, TotalLines, MatchedCount, ExcludedCount, TopMatched, TopExcluded}

type FilterDryRunCombinedSummaryResponse struct {
    NoArchive                bool                     `json:"no_archive"`
    TotalLines               int                      `json:"total_lines"`
    KeptCount                int                      `json:"kept_count"`
    ExcludedByGroupTitleOnly int                      `json:"excluded_by_group_title_only"`
    ExcludedByTvgNameOnly    int                      `json:"excluded_by_tvg_name_only"`
    ExcludedByBoth           int                      `json:"excluded_by_both"`
    GroupTitleTopMatched     []FilterDryRunValueCount `json:"group_title_top_matched"`
    GroupTitleTopExcluded    []FilterDryRunValueCount `json:"group_title_top_excluded"`
    TvgNameTopMatched        []FilterDryRunValueCount `json:"tvg_name_top_matched"`
    TvgNameTopExcluded       []FilterDryRunValueCount `json:"tvg_name_top_excluded"`
}
```
The frontend already knows which mode it requested, so it can deserialize the matching shape without a runtime discriminator field. The combined-mode drawer view renders each attribute's top-value columns as its own two-column block (reusing the existing grid), stacked one attribute above the other, plus the four cause counts as badges above them.

**Combined search result line gets a `verdict` instead of overloading `matched`:**
```go
type FilterDryRunResultLine struct {
    GroupTitle string `json:"group_title"`
    TvgName    string `json:"tvg_name"`
    Matched    bool   `json:"matched"`           // populated in single-attribute mode
    Verdict    string `json:"verdict,omitempty"` // combined mode only: "kept" | "excluded_by_group_title" | "excluded_by_tvg_name" | "excluded_by_both"
}
```
`SearchAttribute` picks which value (`group_title` or `tvg_name`) the search substring is matched against; `Verdict` still reflects both attributes' patterns regardless of which one was searched.

**Evaluation logic mirrors `filter.Manager.ShouldProcess`, not a new independent implementation** - the handler compiles both attributes' patterns (when both are supplied) and, per line, checks each independently before combining, exactly as production filtering already does, so the dry-run and the real pipeline can never disagree on what "combined" means.

## Risks / Trade-offs

- [Reintroducing the Radix nested-dialog outside-click bug if the non-modal drawer wiring in `CreateFilterDialog` isn't copied exactly] → Mitigation: reuse the `justClosedDrawerRef` pattern verbatim from `RunItemsDialog.tsx:41-61`, already proven in this codebase for this exact scenario.
- [Four top-value lists (2 attributes x matched/excluded) in combined mode could feel cluttered in a single drawer] → Mitigation: stack the two attributes' existing two-column blocks vertically rather than inventing a new dense layout; the drawer's own height (100vh, `overflow-y: auto`) already absorbs this without the double-scroll concerns a dialog would have.
- [Changing the dry-run endpoint's request/response shape is a breaking change to that contract] → Mitigation: no external consumers exist (internal-only endpoint used solely by this frontend); frontend and backend are built and deployed together.
- [`FiltersSection` lifting state removes each card's previously-independent test state (e.g. a source picked for one card no longer persists to another)] → Mitigation: intentional per the confirmed design - one shared drawer, one shared source selection, consistent with "test everything" needing a single source choice anyway.

## Open Questions

- Exact JSON field names for the combined request/response (`Attributes` list vs. an alternative shape) are pinned above as a concrete proposal but can still be adjusted during implementation without changing the specs or the overall approach.
- Exact drawer title copy and button labels (e.g. "Tester l'ensemble") are placeholders pending final i18n wording; they don't affect behavior.
