## 1. Backend: `sonarr_matched` in the stats endpoint

- [ ] 1.1 Add `SonarrMatched *int` to `RadarrSonarrStatsResponse` (`internal/api/radarr_sonarr_handlers.go`) and verify the package builds
- [ ] 1.2 In `listRadarrSonarrStats`, after fetching `allSeries` via `GetAllMonitoredSeries`, resolve each series' matched status concurrently via `sonarrMatchCache.matchedStatus` (one goroutine per series, `sync.WaitGroup`, mirroring the concurrency pattern already used in `listSonarrMonitoredSeries`'s filtered path), and count how many are matched
- [ ] 1.3 On any per-series `matchedStatus` error, set `resp.SonarrError = "sonarr_unreachable"` and leave both `SonarrMonitored` and `SonarrMatched` `nil` (no partial result), and verify with a unit test that one failing series fails the whole Sonarr section while leaving the Radarr section unaffected
- [ ] 1.4 On success, set `resp.SonarrMonitored` and `resp.SonarrMatched` and verify with a unit test (extend `TestListRadarrSonarrStats_FullCatalogCounts` or add a sibling test) that `sonarr_matched` reflects the count of series with at least one matched monitored episode, using pre-seeded local `TVShow` records for a mix of matched/unmatched series
- [ ] 1.5 Add a unit test asserting a warm cache (pre-populated via a prior filtered `listSonarrMonitoredSeries` call, or by calling `sonarrMatchCache` directly in the test) is reused by the stats endpoint without additional `GetEpisodesBySeriesID` calls (spy/count assertion, mirroring `TestListSonarrMonitoredSeries_PageSizeBoundsEpisodeFetches`'s call-counting approach)

## 2. Frontend: types and Home tab

- [ ] 2.1 Add `sonarr_matched: number | null` to `RadarrSonarrStats` in `frontend/src/types.ts`
- [ ] 2.2 In `HomeTab.tsx`, compute `sonarrMatchedRatio` from `radarrSonarrStats.sonarr_matched`/`sonarr_monitored` using the same guarded-division logic as `radarrMatchedRatio`, and render a `<Progress.Root>`/`<Progress.Indicator>` pair plus a matched count in the Sonarr subsection, mirroring the existing Radarr subsection's markup
- [ ] 2.3 Update `HomeTab.test.tsx` to cover: matched count and progress bar render when `sonarr_matched` is present; Sonarr section still renders its existing error/loading states unchanged when `sonarr_error` is set or `sonarr_matched` is `null`/absent

## 3. Frontend: arr-suite Résumé sub-tab

- [ ] 3.1 In `RadarrSonarrTab.tsx`, compute `sonarrMatchedRatio` from `stats.sonarr_matched`/`stats.sonarr_monitored`, mirroring `radarrMatchedRatio`
- [ ] 3.2 Add the matched/unmatched secondary-grid row (reusing the `resume.matched`/`resume.unmatched` i18n keys already used by the Radarr card) and a `<Progress.Root>`/`<Progress.Indicator>` pair to the Sonarr card in the "resume" sub-tab, mirroring the Radarr card's layout
- [ ] 3.3 Update `RadarrSonarrTab.test.tsx` to cover: the Sonarr résumé card renders matched/unmatched counts and a progress bar when `sonarr_matched` is present; falls back to the current total-only display when `sonarr_matched` is `null`/absent, without breaking the existing Radarr card assertions

## 4. Verification

- [ ] 4.1 Run the backend test suite (`go test ./internal/api/...`) and confirm all `radarr_sonarr_handlers_test.go` tests pass, including the new ones from section 1
- [ ] 4.2 Run the frontend test suite for the touched components (`HomeTab.test.tsx`, `RadarrSonarrTab.test.tsx`) and confirm they pass
- [ ] 4.3 Manually verify in the running app: Home tab and arr-suite Résumé sub-tab both show a Sonarr matched-vs-monitored progress bar next to the existing Radarr one, and that a Sonarr error state still renders correctly (no progress bar, existing error message)
