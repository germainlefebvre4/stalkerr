# Design

## Context

See `proposal.md` (Why / What Changes). Current state, confirmed in the code:

- `internal/api/radarr_sonarr_handlers.go` returns `monitored` + `missing` (and `has_file` for movies only) on movie, series and episode entries. `missing` is computed in three places: movies `Monitored && !HasFile`, series `Monitored && EpisodeFileCount < TotalEpisodeCount`, episodes `Monitored && !HasFile`. `statusFilterSet`/`matchesStatusFilter` combine the repeatable `status` values (monitored/unmonitored/missing) as an AND. Default is `{monitored}`.
- The frontend models this as one `EtatFilter` (`Set<'monitored'|'unmonitored'|'missing'>`), URL-persisted as `filmsStatus`/`seriesStatus`, rendered by `EtatFilterDropdown` (Radix `DropdownMenu` checkbox items) and `etatBadgeClass`/`etatLabel` in `RadarrSonarrTab.tsx`. Mobile uses `renderStatusIndicator` (dot + label).
- `sonarr.Series` reads `episodeFileCount` and `totalEpisodeCount` from the top level of the `/api/v3/series` JSON. It does not read `episodeCount`. Existing tests only build the struct directly, so no fixture shows the real upstream shape.
- No Sonarr instance or payload is available to the author of this design; the upstream shape therefore has to be checked during implementation (Decision 2).

## Goals / Non-Goals

**Goals:**
- Two orthogonal axes (Suivi, Fichier) end to end: API fields, filters, URL state, badges.
- One server-side definition of `file_state`, used both for the returned field and for the `file` filter.
- Keep the "no extra upstream call" cost model of the existing filters.

**Non-Goals:**
- No change to the match-status filter, the Résumé sub-tab/stats endpoint, the Sonarr match-status cache, or the local-playlist matching.
- No generic multi-select component library; only generalize the existing dropdown enough to serve both axes.
- No per-episode `partial` state, no change to which episodes the sidepanel lists.

## Decisions

### 1. `file_state` is computed once in Go and returned as a string enum
Entries gain `file_state: "downloaded" | "partial" | "none"`; series entries also gain `episode_file_count` and `episode_count` (downloaded / aired) for the ratio. `has_file` and `missing` are dropped from the responses. A single helper derives the state: for movies and episodes `hasFile ? downloaded : none`; for series `fileCount <= 0 ? none : fileCount >= aired ? downloaded : partial`, with `aired == 0` giving `none`.
**Why**: keeps the definition in one place and lets the client stay a pure renderer (it never recomputes from counts). **Alternative**: keep sending booleans and let the client derive the tri-state — rejected, it duplicates the rule in two languages.

### 2. Series denominator is "aired episodes"; the upstream shape is verified first
Series `file_state` uses Sonarr's `episodeCount` (episodes that have aired and are monitored per Sonarr's own statistics) as the denominator instead of `totalEpisodeCount`, so continuing series are not stuck on Partiel. `fileCount > aired` can occur (files for unmonitored episodes); it is treated as downloaded.
Sonarr v3/v4 normally nests these counters in a `statistics` object on each series; the current struct reads them at the top level. Implementation therefore starts with a check against a real `/api/v3/series` payload (task 1.1) and, if needed, makes `sonarr.Series` parse `statistics.episodeFileCount`/`statistics.episodeCount` (top-level kept as fallback for older shapes). If the counters were always zero upstream, today's series-level Missing was silently never true; this change fixes that as a side effect and the release note should say so.
**Alternative**: keep `totalEpisodeCount` — rejected, it counts future episodes. **Alternative**: fan out per series to count episodes — rejected, breaks the "no extra upstream calls" requirement.

### 3. Filter semantics: OR within an axis, AND across axes
`status` (monitored|unmonitored) and `file` (downloaded|partial|none) are repeatable Gin `QueryArray` parameters. Within one parameter, values are OR (they are mutually exclusive facts, so AND would only ever yield empty). Between parameters, AND. Unrecognized values are ignored; `status` absent or reduced to nothing recognized defaults to `{monitored}`, `file` absent means no restriction. `status=missing` (old bookmarks or old frontend) is therefore silently treated like an absent `status`, never an error.
**Alternative**: keep AND within `status` — rejected, it only existed to make `missing` composable and produced empty-result corner cases.

### 4. Frontend state: two URL params per section, narrowed Suivi, new File
`filmsStatus`/`seriesStatus` keep their keys but accept only `monitored`/`unmonitored` (default `monitored`); `filmsFile`/`seriesFile` are new (default empty). Validation falls back to the default on any invalid string, which covers old URLs containing `missing`. The two values are sent to the API as repeated `status` / `file` params. Changing either resets pagination, as the existing filters do.

### 5. One dropdown component for both axes; "Manquants" is a preset button
`EtatFilterDropdown` is generalized (generic option type, no État-specific naming) and instantiated twice per section (Suivi: 2 options; Fichier: 2 options for Films, 3 for Séries). "Manquants" is a small button that calls both setters at once (`status={monitored}`, `file={none}` for Films, `{none,partial}` for Séries) and renders as active when the current selections equal the preset. It owns no state of its own.
**Why**: avoids a third bespoke control and keeps the preset a pure function of the two filters, so it can never disagree with them.

### 6. Badges: Suivi column + Fichier badge; color depends on monitored for Absent
Fichier badge classes: downloaded `badge-success`, partial `badge-pending` (orange; reuse the existing pending/warning style rather than adding a color family), none `badge-failed` when monitored else `badge-neutral`. The Suivi indicator is a neutral/info badge (Surveillé) or neutral (Non surveillé), kept visually quieter than Fichier so Fichier carries the signal. A series Partiel badge appends `episode_file_count/episode_count`. On mobile both go through `renderStatusIndicator` as two dots. Sidepanel header and episode rows reuse the same two small render helpers, so table, header and episode rows cannot drift.

### 7. Locale keys
In `frontend/src/locales/{fr,en}/radarrSonarr.json`, replace the `etat` keys (`filterStatus.etat.*`, the Films/Séries table column keys, and the drawer key) with Suivi/Fichier equivalents (FR: Surveillé, Non surveillé, Téléchargé, Partiel, Absent, Manquants).

## Risks / Trade-offs

- [Upstream counters nested under `statistics`, so today's series Missing is always false] → verify first (task 1.1), fix parsing if needed, mention in the changelog; Radarr movies and Sonarr episodes are unaffected.
- [`episodeCount` semantics differ across Sonarr versions or with unmonitored seasons] → document the definition in the spec ("aired episodes"), cover the boundary cases in unit tests (`aired == 0`, `files > aired`, continuing series), and fall back to `totalEpisodeCount` only if `episodeCount` is absent from the payload.
- [Removing `missing` from the API/URL breaks old bookmarks] → unknown values are ignored on both sides and fall back to defaults; single internal consumer.
- [Series Fichier and playlist matched-ratio measure different things and can disagree] → intentional, as with the former Missing; they are separate badges.
- [Two dropdowns plus a preset widen the filter bar on mobile] → check the existing responsive filter row; allow wrapping rather than horizontal scroll.

## Migration Plan

Backend and frontend ship together (one deploy unit). The API change is a deliberate removal of `missing`/`has_file` and of the `missing` status value, tolerated by the lenient parsing above. No schema or persisted data is involved. Rollback is a plain revert.

## Open Questions

- Exact orange class: reuse `badge-pending` unless it reads as "in progress" rather than "partial" once seen in the UI (cosmetic, no spec impact).
