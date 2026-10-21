# Proposal

## Why

The arr-suite Radarr/Sonarr tabs collapse two independent facts into one "État" label (Monitored / Unmonitored / Missing): whether Radarr/Sonarr is *watching* an item (monitored) and whether the file is actually *on disk* (downloaded). A monitored movie that is fully downloaded just reads "Monitored", an unmonitored movie that has its file is indistinguishable from one that does not, and "Missing" is merely the combination "monitored AND no file" presented as if it were a third peer state. Users cannot tell at a glance — or filter by — "Radarr/Sonarr already has this".

## What Changes

- Split the single État axis into two independent axes, each with its own badge and its own filter:
  - **Suivi** (Surveillé / Non surveillé): Radarr/Sonarr's own `monitored` flag. Default filter selection: Surveillé (today's behavior).
  - **Fichier** (Téléchargé / Partiel / Absent): a new derived `file_state` (`downloaded` | `partial` | `none`). Default filter selection: none selected (everything shown). "Partiel" never applies to a movie (a movie has one file).
- Expose `file_state` on Radarr movie entries, Sonarr series entries, and Sonarr per-episode entries, plus the series' downloaded/total episode counts so the UI can render a ratio (e.g. 18/28) on the Partiel badge.
- **BREAKING** (internal, single consumer — the arr-suite tab): remove "Missing" as a stored/returned/filterable state. The `missing` field disappears from the three response shapes and the `missing` value disappears from the `status` query parameter; `status` accepts only `monitored` / `unmonitored`. "Missing" stays reachable as the derived view Surveillé + (Absent or Partiel).
- Add a repeatable `file` query parameter (`downloaded` / `partial` / `none`) to the Radarr movies and Sonarr series listing endpoints. Values within one axis combine as logical OR (they are mutually exclusive facts); the two axes combine as logical AND. This also removes the former "contradictory combination yields an empty result" corner case.
- Replace the État dropdown with two multi-select dropdowns (Suivi, Fichier) on both the Films and Séries sections, plus a one-click "Manquants" shortcut that sets Suivi = Surveillé and Fichier = Absent + Partiel. Both selections persist in the URL like the existing filters.
- Badges: a Suivi column and a Fichier badge in the Films/Séries tables, in the sidepanel header, and per episode in the season/episode breakdown. Fichier colors: Téléchargé = green, Partiel = orange (series show the ratio), Absent = red when the item is monitored (Radarr/Sonarr is waiting for it) and neutral grey when it is not (intentional). On mobile both render as colored dots via the existing `renderStatusIndicator` convention.
- Series file-state denominators SHALL exclude episodes that have not aired yet, so a continuing series whose aired episodes are all downloaded reads Téléchargé rather than permanently Partiel. This also corrects the former series-level "Missing" definition, which counted unaired episodes. The Sonarr client's series model and parsing of the upstream counters is verified against a real payload as part of this change.
- Vocabulary (FR/EN locales): Surveillé / Non surveillé, Téléchargé / Partiel / Absent, "Manquants" shortcut.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-api`: status filter loses `missing`; new `file` filter and `file_state` field (plus series episode counts) on movies, series and episodes; `missing` field removed; series file-state definition excludes unaired episodes.
- `radarr-sonarr-monitoring-view`: the État filter and badge are replaced by Suivi and Fichier filters/badges (table, sidepanel header, episode rows, mobile dots) and a "Manquants" shortcut.

## Impact

- Backend: `internal/api/radarr_sonarr_handlers.go` (response structs, `statusFilterSet`/`matchesStatusFilter`, new `file` parsing and file-state derivation, episode list), `internal/external/sonarr/sonarr.go` (series model/parsing of upstream counters, `GetMissingSeries` reuse), plus related tests.
- Frontend: `frontend/src/types.ts`, `frontend/src/services/api.ts`, `frontend/src/hooks/useRadarrSonarr.ts` (URL state `filmsStatus`/`seriesStatus` narrowed to Suivi; new `filmsFile`/`seriesFile`), `frontend/src/components/RadarrSonarrTab.tsx`, `frontend/src/components/EtatFilterDropdown.tsx` (generalized to a reusable multi-select dropdown), `frontend/src/locales/{fr,en}/common.json`, and tests.
- Old URLs carrying `filmsStatus=missing` / `seriesStatus=missing` must degrade gracefully (invalid value falls back to the default).
- No database changes, no new upstream calls (the data already comes from the existing catalog/episode fetches).
