## Context

See `proposal.md` - Why/What Changes for motivation and scope. Relevant current state, confirmed by reading the code:

- `internal/api/radarr_sonarr_handlers.go`'s `listRadarrMonitoredMovies` hard-filters to `Monitored` (`if !m.Monitored { continue }`) before matching/pagination; `listSonarrMonitoredSeries` calls `sonarr.GetAllMonitoredSeries` (itself a client-side-filtered wrapper around `GetAllSeries`, which already exists and returns every series unfiltered).
- `radarr.Movie` already carries `Monitored`/`HasFile`/`SizeOnDisk` from the single `/api/v3/movie` fetch; `sonarr.Series` already carries `Monitored`/`EpisodeFileCount`/`TotalEpisodeCount` from the single `/api/v3/series` fetch. Both are populated today and simply dropped before the JSON response.
- `sonarr.GetMissingSeries` already implements the series-level "missing" predicate used by an older feature: `Monitored && EpisodeFileCount < TotalEpisodeCount`. This is the established convention for series-level missing in this codebase and this change reuses it rather than inventing a new one.
- The Sonarr per-episode list (fetched via `GetEpisodesBySeriesID`, exposed through `/api/v1/sonarr/series/{id}/episodes` and consumed by the drawer's season/episode tree) already carries each `Episode`'s own `Monitored`/`HasFile`, used today only for local-playlist matching, not exposed to the client.
- `listRadarrSonarrStats` (Résumé sub-tab) has its own independent fetch/filter logic and is not touched by this change.
- The existing match-status filter (`matched`/`no_match`) is a single-select `<select>`, persisted via `useURLState` as `filmsFilter`/`seriesFilter`. No multi-select control exists anywhere in the frontend today, and no `@radix-ui/react-dropdown-menu` (or any multi-select primitive) is currently a dependency - only `@radix-ui/react-select` (single-select), `-dialog`, `-tabs`, `-progress`, `-tooltip`.
- Mobile compactness convention: `renderStatusIndicator(isMobile, badgeClass, label)` already swaps a full-text badge for a small colored dot on mobile, used today by the Résumé cards.

## Goals / Non-Goals

**Goals:**
- Define the exact response fields and query-parameter shape for the new État (Monitored/Unmonitored/Missing) status.
- Define where the "missing" predicate is computed (server-side, once) and what the client receives (pre-computed booleans, not raw counts).
- Pick a concrete UI approach for the new cumulative multi-select filter control, given none exists in the codebase yet.
- Confirm the default-preserving migration path for removing the current hard monitored-only server filter.

**Non-Goals:**
- No changes to the Résumé sub-tab or its stats endpoint.
- No change to the existing match-status (Matched/No match) filter's own semantics - it keeps its current single-select, OR-with-itself-not-applicable behavior; the new État filter is additive and independent.
- No generic, reusable multi-select component for other filters elsewhere in the app - scoped to this one control; extraction is a future concern if a second use case appears.
- No change to the Sonarr match-status cache or its per-series episode fan-out logic - the new État data comes from fields already present in the base catalog fetch, not from that cache.

## Decisions

### 1. Server computes and exposes booleans, not raw counts
Each Radarr movie entry gains `monitored: bool` and `missing: bool` (in addition to the existing `has_file`); each Sonarr series entry gains `monitored: bool` and `missing: bool`; each Sonarr episode entry (in the per-episode list) gains `monitored: bool` and `missing: bool`. The client never receives or recomputes `episodeFileCount`/`totalEpisodeCount`.

**Rationale**: keeps the "missing" derivation in exactly one place (Go), reused identically by the new `status` filter and by the fields returned to the client - no risk of frontend/backend drift on the definition of "missing". A single tiny frontend helper (`etat(monitored, missing) => 'monitored' | 'unmonitored' | 'missing'`) then covers all three item types (movie, series, episode) identically.
**Alternative considered**: expose raw `hasFile`/`episodeFileCount`/`totalEpisodeCount` and let the frontend derive "missing" - rejected, duplicates logic across two languages/layers for no benefit.

### 2. Sonarr series-level "missing" reuses the existing `GetMissingSeries` formula, at the whole-series granularity
`missing = Monitored && EpisodeFileCount < TotalEpisodeCount`, using the aggregate counters Sonarr itself returns in the base `/api/v3/series` listing - no per-series episode fan-out, no extra upstream calls, independent of and cheaper than the playlist match-status computation (which does fan out per series and is cached for that reason).

**Rationale**: reuses an established, already-tested formula in this codebase rather than inventing a parallel one; keeps the État filter's cost model flat regardless of catalog size, satisfying the "no additional upstream calls" scenario in the API spec.
**Trade-off to note explicitly**: `TotalEpisodeCount` as Sonarr reports it is not scoped to "monitored episodes" the way the playlist match-status ratio is - the two "series is (mostly) complete" signals (missing-file vs matched-in-playlist) measure different things and can legitimately disagree for the same series. This is intentional and documented in the spec/proposal, not a bug to reconcile.

### 3. New cumulative filter control: adopt `@radix-ui/react-dropdown-menu` with `CheckboxItem`s
Add `@radix-ui/react-dropdown-menu` as a new dependency and build a small `EtatFilterDropdown` component (trigger button + `DropdownMenu.Content` with three `DropdownMenu.CheckboxItem`s) local to the arr-suite tab.

**Rationale**: the codebase already standardizes on unstyled Radix primitives for exactly this class of problem (`Dialog`, `Tabs`, `Select`, `Tooltip`); `react-dropdown-menu` is the matching primitive for "trigger + multi-checkable menu" and comes with focus management, keyboard navigation, and outside-click/Escape handling already solved and tested - the same guarantees the rest of the tab already relies on for its Dialog-based drawer.
**Alternative considered**: hand-roll a popover with a controlled boolean + a manual outside-click listener (no new dependency) - rejected: this class of bug (focus trap, Escape handling, outside-click edge cases with the already-present nested drawer/dialog) is exactly what the project already delegates to Radix elsewhere; re-implementing it by hand for one control adds maintenance risk for no real benefit.
**Alternative considered**: reuse the existing `@radix-ui/react-select` - rejected: Radix `Select` is fundamentally single-value; forcing multi-select through it would require fighting its own state model.

### 4. Query parameter: repeated `status` values, AND semantics, default `monitored`
Both listing endpoints accept a repeatable `status` query parameter (`?status=monitored&status=missing`), read via Gin's `QueryArray`. Absent parameter is treated as `status=monitored`. Multiple values combine as a logical AND (see proposal/spec for the confirmed rationale and the accepted contradictory-combination-yields-empty behavior).

**Rationale**: repeated query parameters are the idiomatic Gin/REST way to pass a multi-value filter and avoid hand-rolling comma-split parsing (and its edge cases) for no benefit.
**Alternative considered**: single comma-separated value (`?status=monitored,missing`) - rejected as unnecessary manual parsing with no advantage over native repeated-param support.

### 5. Frontend URL persistence mirrors the existing filter pattern
New URL-persisted state `filmsStatus`/`seriesStatus`, stored as a comma-joined string (e.g. `monitored,missing`) via the existing `useURLState` hook, parsed to/from a `Set<'monitored'|'unmonitored'|'missing'>` in the component. Default (key absent from URL) is `monitored` alone - identical to today's implicit server-side default, so a first-time visit or a link without the param renders exactly what the tab renders today.

**Rationale**: reuses the exact mechanism already proven for `filmsFilter`/`seriesFilter`, keeping the two filters symmetric and independently shareable via URL.

### 6. Badge styling reuses the existing vocabulary; mobile reuses the existing dot convention
Missing -> `badge-failed`; Monitored (not missing) -> `badge-success`; Unmonitored -> `badge-neutral`. On mobile-width viewports, the new État badge renders via the existing `renderStatusIndicator`-style compact dot + short label instead of the full badge, everywhere the new badge appears (table row / mobile card / sidepanel header / episode row).

**Rationale**: consistent with every other status badge already in this tab; avoids inventing a fourth badge color family.

## Risks / Trade-offs

- [Risk] Dropping the hard monitored-only server filter changes what the listing endpoints can return by default. -> Mitigated by keeping the exact same default (monitored-only) when `status` is omitted, and by this being an internal-only API with a single consumer that always sends the filter explicitly once implemented.
- [Risk] AND semantics across the three État values can surprise a user used to checkbox groups meaning OR. -> Accepted explicitly by the user during exploration; the tri-state visual distinction and documented empty-result behavior for contradictory combinations make the outcome legible rather than a silent bug.
- [Risk] Adding a new Radix package increases dependency surface. -> Small, same vendor family already used pervasively, actively maintained, tree-shakeable; judged lower long-term risk than a hand-rolled popover.
- [Risk] Sonarr's own "missing" signal (file-presence based) and the existing playlist match-status ratio can disagree for the same series (different questions). -> Documented explicitly in this design and the proposal so implementers and future readers don't try to reconcile the two into one number.

## Migration Plan

- Backend change is purely additive to the response shape (`monitored`/`missing` fields) plus a new optional `status` query parameter with a default-preserving fallback - safe to deploy on its own, no data migration, no schema change.
- Frontend change ships in the same release as the backend fields it depends on (single monorepo/deploy unit); no feature flag needed given the low blast radius (one tab, one internal consumer, default-preserving behavior).
- Rollback is a plain revert of the change's commit(s); no persisted state to unwind (per the API spec's existing "no caching or persistence" requirement, unaffected by this change).

## Open Questions

- Exact French copy for the three État labels and the Missing badge text (e.g. "Manquant") is a wording detail that can be finalized during implementation without affecting the approach, specs, or task breakdown.
