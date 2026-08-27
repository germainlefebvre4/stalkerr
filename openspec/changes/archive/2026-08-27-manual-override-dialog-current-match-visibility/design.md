## Context

`ManualOverrideDialog` (`frontend/src/components/ManualOverrideDialog.tsx`) receives `overrideItemData: PlaylistItem` when opened. That object already carries `movie`/`tvshow` (populated whenever `MovieID`/`TVShowID` is set, from `toItemResponse` in `internal/api/handlers.go`), plus `override_by`/`override_at`. The dialog currently only reads `overrideItemData.tvshow?.tmdb_year` (to pre-fill the search year) and never surfaces the rest. The bulk candidate list added by `bulk-associate-tvshow-episodes` (see archived change) reuses the same `playlist` array already loaded by `App.tsx`/`PlaylistTab.tsx`, so each `OverrideCandidate.item` is also a full `PlaylistItem` with the same `movie`/`tvshow`/`tvg_name` fields available client-side, no new fetch required. See proposal.md - Why for the user-facing problem.

## Goals / Non-Goals

**Goals:**
- Surface, with data already in hand, what an item (opened item or bulk candidate) is currently matched to, or would be matched to, before the user commits an override.
- Keep the detected-season/episode extraction rule (frontend regex `S(\d+)[\s-]*E(\d+)`, case-insensitive) as the single source of truth for "detected from title" previews, reusing it for both the opened item and each candidate rather than introducing a second implementation.

**Non-Goals:**
- Reconciling the frontend's detection regex with the backend classifier's broader pattern set (`internal/classifier/classifier.go`) - out of scope for this change; the backend remains the source of truth at submit time via its own fallback extraction, this change only improves what is previewed client-side before submit.
- Any new backend endpoint or response field - everything needed is already returned by the existing playlist listing and single-item override responses.

## Decisions

**The "current match" panel reads directly from `overrideItemData`/`candidate.item`, no new state.**
`overrideItemData.tvshow` / `.movie` and `override_by`/`override_at` are already present on the prop; the panel is a pure render of existing fields, shown whenever `overrideItemData.tvshow || overrideItemData.movie` is truthy. It does not depend on `selectedResult`, so it stays visible through the whole search/selection flow instead of disappearing once the user starts picking a new match.

**Detected season/episode preview is decoupled from `selectedResult`.**
The existing `useEffect` that extracts season/episode from `overrideItemData.tvg_name` (today gated behind `overrideMediaType === 'tvshow' && selectedResult` for *rendering* the inputs) keeps its extraction logic unchanged; only the rendering condition drops the `selectedResult` requirement so the fields (and their pre-populated values) show as soon as `"tvshow"` mode is active. This changes visibility only - the values submitted on "Forcer l'association" are unaffected.

**Candidate rows compute their own preview client-side using the same `cleanRawTitle`-adjacent extraction helper, not a new API round trip.**
Each candidate is a full `PlaylistItem` already loaded in `playlist`. For a candidate with an existing `tvshow`, the row shows that association's season/episode directly. For a candidate without one, the row runs the same season/episode regex already used for the opened item against `candidate.item.tvg_name` and shows the result (or nothing, matching the existing "leave empty rather than guess" rule). This preview is display-only: the bulk submit path continues to omit `season`/`episode` for candidates (see archived design.md - "Decisions") so the backend's own (broader) extraction remains authoritative at submit time; the frontend preview and the backend's actual extraction can occasionally disagree for titles the frontend regex doesn't recognize but the backend classifier does; that gap is accepted per Non-Goals.

## Risks / Trade-offs

- [Risk] The frontend preview (narrower regex) can under-detect season/episode for a title format the backend classifier recognizes (e.g. `1x05`, `Saison 1 Épisode 5`), showing "not detected" for a candidate the backend will actually match correctly → Mitigation: the preview is explicitly a preview, not a guarantee; per Non-Goals, aligning the two extractors is left for a follow-up change if this proves confusing in practice.
- [Risk] Showing a persistent current-match panel plus a detected-preview row adds vertical space to an already content-dense modal → Mitigation: keep the panel compact (a single-line summary, consistent with the existing raw-title/group info block) and scope its fields to what's already decided as useful (title, year, season/episode, override_by/at).
