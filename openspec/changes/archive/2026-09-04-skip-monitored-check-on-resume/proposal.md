## Why

`mergeIncompleteDownloads` (`internal/scheduler/build.go`) resumes any `download_info` row left in an incomplete/interrupted state (e.g. `downloading`, `paused`, stale-locked `failed`/`retrying`) purely from local database state, without ever re-checking Radarr's/Sonarr's *current* monitored status for the movie/series it belongs to. If a download attempt gets interrupted (crashed process, killed container, etc.) and the user subsequently un-monitors that movie/series in Radarr/Sonarr — explicitly deciding they no longer want it — the next scheduled `download` run has no way to know that: it finds the stale lock, clears it, and blindly resumes the leftover attempt as if nothing had changed. This was observed in production: a movie flagged unmonitored in Radarr was re-downloaded anyway because an interrupted attempt from days earlier got picked back up.

Radarr's/Sonarr's `Monitored` field already reflects the user's current intent and is already fetched in bulk during the scheduled run (via `GetAllMovies`/`GetAllMonitoredSeries`, added by the `reconcile-radarr-sonarr-download-paths` change) or resolvable per-item; nothing today reconnects that signal to the resume path.

## What Changes

- Before `mergeIncompleteDownloads` folds an incomplete movie/series download back into a claimable stream, it SHALL check whether that movie/series is still monitored in Radarr/Sonarr, using the same live/bulk lookup data already available to the run.
- An incomplete download whose movie/series is confirmed unmonitored SHALL be left out of this run's stream set (not resumed, not silently deleted) — the scheduler simply stops trying it going forward, the same way it would if it had never seen it as missing/wanted in the first place.
- An incomplete download whose movie/series cannot be confirmed either way (no Radarr/Sonarr match at all, e.g. a TMDB-only/M3U-matched item, or Radarr/Sonarr temporarily unreachable) SHALL still be resumed as today — this change only acts on an explicit, confirmed `monitored: false`, never on absence of information.
- No change to how an incomplete download is detected or to the stale-lock cleanup step itself — only to whether a detected incomplete download is folded into this run's stream set.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `media-download-scheduling`: the "Interrupted downloads resume within their own stream" requirement gains a precondition — resumption is skipped when the owning movie/series is confirmed unmonitored in Radarr/Sonarr.

## Impact

- **Backend**: `internal/scheduler/build.go` (`mergeIncompleteDownloads`), and its `BuildDeps`/`RadarrClient`/`SonarrClient` interfaces if a new lookup is needed there.
- **No impact** on stale-lock cleanup, tier-1/tier-2 stream classification, or any other scheduling requirement — this only narrows which incomplete downloads get resumed.
