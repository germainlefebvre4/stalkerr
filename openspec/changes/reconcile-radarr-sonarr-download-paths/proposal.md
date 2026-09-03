## Why

`download_info.download_path` is written once (on completion, or when the user explicitly uses the app's own "Renommer"/"Déplacer" actions) and never revisited afterward. When a completed download's parent folder is renamed outside those two paths — e.g. directly on disk, with Radarr/Sonarr's own `Path` updated to match afterward because Radarr/Sonarr's own rename failed for that item — `download_path` keeps pointing at the old name forever. Since the Erreurs tab's `missing_year`/`year_mismatch`/`unknown_format` checks (`fileparser.Parse`) run a regex against this stored string rather than the real file, a folder that now correctly contains a year still reports "année manquante" indefinitely, and the only way to clear it today is to redo the same rename through the app's UI.

Radarr and Sonarr already expose the answer to "where does this item actually live now" through `movie.Path`/`series.Path` on their APIs (the same authoritative root the scheduler already trusts when deciding new download destinations, per `sonarr-series-path-routing`). Nothing today reconnects that live source back to an already-completed download's stored path.

## What Changes

- During the existing scheduled `download` run, fetch Radarr's and Sonarr's current full libraries (`GetAllMovies`/`GetAllMonitoredSeries` — a new, single per-run call each, distinct from the `GetMissingMovies`/`GetMissingEpisodes` calls the run already makes for scheduling) and compare each completed download's stored root directory against the matching movie's/series' current `Path`, rewriting `download_info.download_path` (preserving the season/filename segment stalkeer's own naming convention already produced) when they differ. This adds exactly one new full-library fetch per Radarr/Sonarr per run — no per-item API calls.
- Add a manual "Resynchroniser" action, scoped to a single `DownloadInfo` and available next to the existing "Déplacer"/"Renommer" actions in the Downloads tab's details sidepanel, that performs the same live-lookup-and-correct check on demand for one item instead of waiting for the next scheduled run.
- When the manual action's live lookup finds no matching movie/series in Radarr/Sonarr, the system SHALL report this explicitly (e.g. "non géré par Radarr/Sonarr") rather than silently doing nothing — this is also how an item outside this change's scope is distinguished from one where nothing needed correcting.
- **Explicitly out of scope**: downloads whose content was matched directly against TMDB by the IPTV/M3U pipeline (`processor.go`) without a confirmed corresponding entry in the user's own Radarr/Sonarr instance. For these, `download_path` remains the only record of the file's location and this change does not attempt to validate or correct it — a TMDB-derived `TVDBID` alone does not establish that the title is actually managed by Radarr/Sonarr, so both mechanisms above rely on a live "found" lookup, not on the presence of that field, to decide whether an item is in scope.
- No change to how `download_path` is computed for a download in progress, and no change to the Erreurs tab's read logic (`fileparser.Parse` keeps operating on `download_path` as before) — this change corrects the stored value at its source instead of adding a live join on every read, keeping `downloads-enrichment-api`'s existing `< 500ms` response-time requirement intact.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `sonarr-series-path-routing`: gains requirements for re-deriving and correcting an already-completed download's stored destination root from the current `movie.Path`/`series.Path`, both opportunistically during the scheduled `download` run and via an on-demand live lookup, scoped to items with a confirmed Radarr/Sonarr match.
- `api-media-management`: gains a new endpoint for the per-download on-demand resync action, alongside the existing move/rename endpoints, including the explicit "not managed by Radarr/Sonarr" outcome.
- `downloads-details-sidepanel`: gains a "Resynchroniser" action alongside the existing "Déplacer"/"Renommer" actions, available only for completed downloads.

## Impact

- **Backend**: `internal/scheduler/build.go` (or a new step in the `download` run) to fetch the full Radarr/Sonarr libraries (`GetAllMovies`/`GetAllMonitoredSeries`) and compare/correct stored paths against them; `internal/downloader/destpath.go` helpers reused to preserve the season/filename segment; new handler + route in `internal/api/handlers_frontend.go`/`internal/api/api.go` for the manual resync endpoint, reusing `radarr.Client.GetMovieByTMDBID`/`sonarr.Client.GetSeriesByTVDBID`.
- **Database**: updates to `download_info.download_path` for affected rows; no schema change anticipated beyond what design.md may identify for tracking outcome/last-checked state.
- **Frontend**: new sidepanel action in `frontend/src/components` (details sidepanel, not the read-only Erreurs sidepanel), new API call in `frontend/src/services/api.ts`.
- **No impact** on `downloads-enrichment-api`, `downloads-errors-view`, or `fileparser` — the Erreurs tab keeps reading `download_path` exactly as today.
