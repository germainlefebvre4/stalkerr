## Context

The `problem` filter and `fileparser.FileInfo` diagnostics already exist and are exercised today by the Downloads tab's `problemFilter` dropdown (single value at a time, `internal/api/handlers_frontend.go`, `matchesProblem`). See `proposal.md` - Why for the motivation. This design covers how the new tab reuses that machinery instead of duplicating it, and how the frontend adds a fifth-ish desktop-only tab on top of the existing tab-array + `768px` breakpoint mechanism (`frontend/src/App.tsx`, `useMediaQuery`).

## Goals / Non-Goals

**Goals:**
- Reuse the existing `GET /api/v1/downloads` enrichment endpoint and its `file_info`/`content` shape as the sole data source — no new endpoint, no new DB columns.
- Keep the multi-value `problem` filter change backward compatible with the single-value behavior the Downloads tab already depends on.
- Keep the new tab's frontend code additive (new components + one new entry in the tab list/breakpoint gating), not a refactor of `DownloadsTab`.

**Non-Goals:**
- Persisting a new "invalid"/"error" status or reason on any DB row (per proposal scope, `uncategorized`/no-match items stay out of scope for this change).
- Any corrective action (rename, move, associate, force-download, or re-parsing) from the new tab.
- Changing the existing Downloads tab's own `problemFilter` dropdown or sidepanel — they continue to work as today, single-value only, unaffected by this change.

## Decisions

**Reuse `GET /api/v1/downloads` with an extended `problem` param, instead of a new endpoint.**
The data needed (content + file_info) is already exactly what this endpoint returns; the only gap is that `problem` only ever matched one value. Extending it to accept a comma-separated OR-list is a small, backward-compatible change to `matchesProblem`'s call site, versus standing up and maintaining a parallel endpoint. Alternative considered: a dedicated `/api/v1/downloads/errors` endpoint - rejected because it would duplicate the enrichment/pagination logic that already exists and would need to be kept in sync with `downloads-enrichment-api` as it evolves.

**Frontend fetches the combined filter (`missing_year,year_mismatch,unknown_format`) as the tab's default, and a single value when a specific reason is selected.**
This maps directly onto the reason filter UI (`Tous` vs. one reason) without the frontend needing to merge multiple requests or de-duplicate rows client-side. Alternative considered: always fetch all four problem types unfiltered and compute which reasons apply client-side (as `DownloadsTab` already does) - rejected because the whole point of this change is server-side, paginated aggregation across the full catalog (per earlier scoping decision), not a page's worth of client-side derivation.

**Reason badges are derived from the same `file_info` flags already in the response, not a new "reasons" array field.**
The frontend already knows how to read `has_year_in_path`, `year_mismatch`, and `is_valid_format` (see `DownloadsTab.tsx`'s existing derivation). The new table computes its badges the same way per row. Alternative considered: have the backend add an explicit `reasons: string[]` field to `DownloadEnrichedResponse` - simpler for the frontend, but changes the shared DTO used by the existing Downloads tab too; not needed since the flags are already boolean and self-explanatory, so it's deferred rather than done speculatively.

**New tab is gated by the existing `useMediaQuery` breakpoint hook, added to the desktop tab list only.**
Consistent with how `frontend-responsive-layout` already draws the line at `768px`. The tab is simply never added to the mobile bottom tab bar's tab array, and an effect mirrors the existing pattern used elsewhere in the app for falling back off a tab that becomes unavailable (e.g. when a tab is removed from `VALID_TABS` at runtime) if the viewport narrows while it's active.

**Sidepanel is a new component, not a variant of `MediaOccurrenceDrawer` or the Downloads details sidepanel.**
Per the scoping conversation, the information hierarchy needs to be diagnostic-first, which is a different section order than both existing drawers. Building a small dedicated component avoids threading a "diagnostic-first" mode flag through either existing drawer.

## Risks / Trade-offs

- [Comma-separated `problem` values is a new micro-syntax on a query param] -> Mitigation: it's purely additive (single value keeps working unchanged, per the `Multi-Value Problem Filter` requirement's own scenario), and matches a pattern already familiar from other list-style query params elsewhere in HTTP APIs generally.
- [Two sidepanel components now exist for closely related download data (Downloads tab's and this tab's)] -> Mitigation: acceptable per explicit product decision to keep them separate; if they drift in content over time that's a future consolidation call, not a correctness risk today.
- [Filtering out `low_quality` from this tab means a download can have a real quality problem invisible here] -> Mitigation: intentional per scope (`low_quality` is a quality concern, not a naming/parsing error); it remains visible via the existing Downloads tab filter.
