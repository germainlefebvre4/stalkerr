## Why

The only download triggers today are the scheduled `radarr`/`sonarr` CLI jobs, and they source their work list from Radarr/Sonarr's own "missing" endpoints (items with zero file). Once a single occurrence of a movie or episode has been downloaded and imported, Radarr/Sonarr permanently drops it from that list — the existing `--force` flag only bypasses stalkeer's own per-media dedup, not that upstream gate. A user who spots, in the ingested playlist, a second interesting occurrence of a title already partially satisfied (e.g. a movie downloaded in SD while an HD occurrence also exists in the M3U) has no way to fetch that specific occurrence today, short of a destructive full reset of the media that discards all history and re-triggers reprocessing of every occurrence from scratch.

## What Changes

- New POST endpoint(s) to force the download of one specific `ProcessedLine` occurrence, independently of its sibling occurrences for the same movie/TV episode and independently of Radarr/Sonarr's "missing" status.
- The request is rejected (fail-closed) unless a live Radarr (movie) or Sonarr (episode) lookup, performed synchronously at click time, confirms the associated media exists in that library — no pre-check when the sidepanel simply opens, since most opens never lead to a click. That same lookup resolves the destination root folder (`movie.Path` / series path), which is not persisted anywhere today.
- New Radarr and Sonarr client methods to look up a movie/series by TMDB/TVDB ID directly, independent of the "missing" list (Radarr's `GET /movie?tmdbId=`, Sonarr's equivalent lookup) — needed because the target occurrence's title can already have a file.
- The forced download itself runs asynchronously (background job) through the existing `DownloadInfo`/state machinery, so progress and completion surface through the existing Downloads screens without new UI plumbing for progress tracking.
- Destination filenames produced by a forced download are suffixed with the occurrence's resolution (e.g. `Title (Year) [1080p].mkv`) so they cannot silently overwrite a sibling occurrence already sitting in the same Radarr/Sonarr folder. This suffixing applies **only** to forced downloads; the existing automatic pipeline's naming is unchanged.
- New "Forcer le téléchargement" action in the playlist sidepanel (drawer), scoped to the single occurrence currently open. It is unavailable when the occurrence isn't matched to a movie/TV show, or is already downloaded/downloading; otherwise it triggers the new endpoint and reflects queued/in-progress/error feedback after the click.

## Capabilities

### New Capabilities
- `force-download-media-occurrence`: eligibility rules (media must exist live in Radarr/Sonarr, checked at trigger time), the async force-download trigger endpoint(s) for a single occurrence, the new Radarr/Sonarr existence+path lookup, and resolution-suffixed destination naming for forced downloads.

### Modified Capabilities
- `m3u-playlist-details-sidepanel`: adds the "Forcer le téléchargement" action to the drawer for the currently-open occurrence, including its available/unavailable and post-click feedback states.

## Impact

- Backend: `internal/external/radarr` and `internal/external/sonarr` (new lookup-by-ID client methods), `internal/downloader` (resolution-suffixed destination naming variant), `internal/api` (new handler/route for the force-download trigger), `internal/matcher` (reuse of `FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates` style eligibility, now exercised outside the CLI for the first time).
- Frontend: `frontend/src/components/PlaylistTab.tsx` (new drawer section/button and its states), `frontend/src/services/api.ts` (new API call), i18n resources for the new drawer strings (fr/en).
- No breaking changes; purely additive. No changes to the existing automatic `radarr`/`sonarr` CLI download flow or its file naming.
