package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/glefebvre/stalkeer/internal/api"
	"github.com/glefebvre/stalkeer/internal/circuitbreaker"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/httpclient"
	"github.com/glefebvre/stalkeer/internal/external/jellyfin"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/notifier"
	"github.com/glefebvre/stalkeer/internal/policy"
	"github.com/glefebvre/stalkeer/internal/retry"
	"github.com/glefebvre/stalkeer/internal/scheduler"
	"github.com/glefebvre/stalkeer/internal/settings"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// newRunBreaker builds a circuit breaker for one Radarr/Sonarr client built
// for this download run, using the same defaults as the API server's
// long-lived breakers (see internal/api.newServiceBreaker) so a service
// failing partway through a run fails fast for the rest of it instead of
// letting every remaining worker independently wait out its own timeout.
func newRunBreaker() *circuitbreaker.CircuitBreaker {
	return circuitbreaker.New(circuitbreaker.Config{
		MaxFailures:         5,
		Timeout:             60 * time.Second,
		MaxHalfOpenRequests: 1,
		IsSuccessful:        httpclient.IsSuccessful,
	})
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download missing movies and TV episodes from Radarr and Sonarr",
	Long: `Fetch missing movies from Radarr and missing episodes from Sonarr within a single
run, schedule the work as season-local streams drained by a shared worker pool (so a
season in progress is never abandoned mid-way and a series' seasons are attempted in
ascending order), and download matched items from M3U playlist stream URLs.

Already-downloaded content eligible for a quality upgrade is folded into every run and
drawn occasionally alongside new content (see downloads.force_tier_probability), and any
incomplete/interrupted download is resumed as part of its own stream automatically.

This command replaces the removed "radarr" and "sonarr" commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		limit, _ := cmd.Flags().GetInt("limit")
		parallel, _ := cmd.Flags().GetInt("parallel")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if err := config.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		cfg := settings.Effective()

		if parallel <= 0 {
			parallel = cfg.Downloads.MaxParallel
		}
		if parallel <= 0 {
			parallel = 3
		}

		logger.InitializeLoggers(cfg.GetAppLogLevel(), cfg.GetDatabaseLogLevel())
		notif := notifier.FromConfig(cfg.Notifications)

		fmt.Println("=== Unified Download Command ===")
		if dryRun {
			fmt.Println("Mode: DRY RUN (no downloads will occur)")
		}
		if limit > 0 {
			fmt.Printf("Limit: %d work unit(s)\n", limit)
		}
		fmt.Printf("Parallel workers: %d\n", parallel)
		fmt.Println()

		if err := database.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		ctx := context.Background()
		db := database.Get()
		dl := downloader.New(time.Duration(cfg.Downloads.Timeout)*time.Second, cfg.Downloads.RetryAttempts, cfg.Downloads.MinFileSizeMB)

		policyEngine := newPolicyEngine(ctx, cfg)
		defer policyEngine.Stop()
		dl.SetPolicyEngine(policyEngine)

		var radarrClient scheduler.RadarrClient
		var radarrFullClient *radarr.Client
		var radarrRescanner movieRescanner
		if cfg.Radarr.URL != "" && cfg.Radarr.APIKey != "" {
			radarrFullClient = radarr.New(radarr.Config{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
				Timeout: time.Duration(cfg.Downloads.Timeout) * time.Second,
				Logger:  logger.AppLogger(),
				RetryConfig: retry.Config{
					MaxAttempts:       cfg.Downloads.RetryAttempts,
					InitialBackoff:    2 * time.Second,
					MaxBackoff:        30 * time.Second,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				Breaker: newRunBreaker(),
			})
			radarrClient = radarrFullClient
			radarrRescanner = radarrFullClient
		} else if verbose {
			fmt.Println("Radarr is not configured, skipping movie fetch")
		}

		var sonarrClient scheduler.SonarrClient
		var sonarrFullClient *sonarr.Client
		var sonarrRescanner seriesRescanner
		if cfg.Sonarr.URL != "" && cfg.Sonarr.APIKey != "" {
			sonarrFullClient = sonarr.New(sonarr.Config{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
				Timeout: time.Duration(cfg.Downloads.Timeout) * time.Second,
				Logger:  logger.AppLogger(),
				RetryConfig: retry.Config{
					MaxAttempts:       cfg.Downloads.RetryAttempts,
					InitialBackoff:    2 * time.Second,
					MaxBackoff:        30 * time.Second,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				Breaker: newRunBreaker(),
			})
			sonarrClient = sonarrFullClient
			sonarrRescanner = sonarrFullClient
		} else if verbose {
			fmt.Println("Sonarr is not configured, skipping episode fetch")
		}

		if radarrClient == nil && sonarrClient == nil {
			fmt.Fprintln(os.Stderr, "Error: neither Radarr nor Sonarr is configured")
			os.Exit(1)
		}

		fmt.Println("Reconciling completed download paths...")
		movieMonitored, seriesMonitored := reconcileDownloadPaths(ctx, db, radarrFullClient, sonarrFullClient, verbose)

		fmt.Println("Building stream set...")
		streams, err := scheduler.BuildStreams(ctx, scheduler.BuildDeps{
			Config:                 cfg,
			Radarr:                 radarrClient,
			Sonarr:                 sonarrClient,
			DB:                     db,
			StateManager:           dl.GetStateManager(),
			MonitoredMovieTMDBIDs:  movieMonitored,
			MonitoredSeriesTVDBIDs: seriesMonitored,
		})
		if err != nil {
			notifyBuildStreamsFailure(notif, err)
			fmt.Fprintf(os.Stderr, "Error building streams: %v\n", err)
			os.Exit(1)
		}

		streams = scheduler.ApplyLimit(streams, limit)

		totalItems := 0
		for _, s := range streams {
			totalItems += len(s.Items)
		}
		fmt.Printf("Built %d stream(s), %d item(s) total\n\n", len(streams), totalItems)

		if len(streams) == 0 {
			fmt.Println("Nothing to download!")
			return
		}

		if dryRun {
			printDryRunPlan(streams)
			return
		}

		sched := scheduler.NewScheduler(streams, cfg.Downloads.ForceTierProbability)
		stats := runDownloadWorkerPool(ctx, sched, dl, cfg, parallel, verbose, radarrRescanner, sonarrRescanner)

		notifyDownloadRunResult(notif, stats)
		notifyJellyfin(ctx, cfg, stats.changedPaths)

		fmt.Println("\n=== Download Summary ===")
		fmt.Printf("Total items:      %d\n", stats.Total)
		fmt.Printf("Downloaded:       %d\n", stats.Downloaded)
		fmt.Printf("Failed:           %d\n", stats.Failed)
	},
}

func init() {
	downloadCmd.Flags().Bool("dry-run", false, "preview planned streams without downloading")
	downloadCmd.Flags().Int("limit", 0, "maximum number of work units (movies/series) to consider (0 = no limit)")
	downloadCmd.Flags().Int("parallel", 0, "number of concurrent worker streams")
	downloadCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.AddCommand(downloadCmd)
}

// reconcileDownloadPaths fetches the current Radarr/Sonarr full libraries
// (one call each, distinct from the missing-content fetches BuildStreams
// performs) and corrects any completed download's stored download_path that
// has drifted from its movie's/series' current path. A fetch failure for
// either service is tolerated: that service's corrections are skipped for
// this run (logged, not fatal) rather than failing the run's primary
// scheduling work, and the run retries reconciliation next time it runs.
//
// It also returns a monitored-status snapshot (true/false keyed by
// Movie.TMDBID / TVDBID, entry present only for a fetched item) built from
// the exact same library fetch, for skip-monitored-check-on-resume's
// mergeIncompleteDownloads check — this adds no further Radarr/Sonarr API
// calls beyond what path reconciliation already makes.
func reconcileDownloadPaths(ctx context.Context, db *gorm.DB, radarrClient *radarr.Client, sonarrClient *sonarr.Client, verbose bool) (movieMonitored map[int]bool, seriesMonitored map[int]bool) {
	movieTMDBPaths := map[int]string{}
	movieMonitored = map[int]bool{}
	if radarrClient != nil {
		movies, err := radarrClient.GetAllMovies(ctx)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: failed to fetch Radarr library for path reconciliation, skipping: %v\n", err)
			}
		} else {
			for _, m := range movies {
				if m.Path != "" {
					movieTMDBPaths[m.TMDBID] = m.Path
				}
				movieMonitored[m.TMDBID] = m.Monitored
			}
		}
	}

	// GetAllSeries (unlike GetAllMonitoredSeries) returns every series
	// regardless of monitored status, so an unmonitored series can still be
	// recorded as monitored=false here instead of merely being absent.
	seriesTVDBPaths := map[int]string{}
	seriesMonitored = map[int]bool{}
	if sonarrClient != nil {
		series, err := sonarrClient.GetAllSeries(ctx)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: failed to fetch Sonarr library for path reconciliation, skipping: %v\n", err)
			}
		} else {
			for _, s := range series {
				if s.Path != "" {
					seriesTVDBPaths[s.TvdbID] = s.Path
				}
				seriesMonitored[s.TvdbID] = s.Monitored
			}
		}
	}

	if len(movieTMDBPaths) > 0 || len(seriesTVDBPaths) > 0 {
		api.ReconcileScheduledDownloadPaths(db, movieTMDBPaths, seriesTVDBPaths)
	}

	return movieMonitored, seriesMonitored
}

// newPolicyEngine builds the adaptive-download-throttling PolicyEngine for
// this run: the currently stored weekly bandwidth schedule, plus a Jellyfin
// active-playback checker whenever a Jellyfin base URL is configured (the
// engine itself only actually polls when jellyfin.playback_check_enabled is
// set - see policy.New). A failure to load the schedule is logged and
// treated as "no schedule defined" (effective action none), matching
// "Default Schedule Is Unrestricted" rather than aborting the run.
func newPolicyEngine(ctx context.Context, cfg *config.Config) *policy.Engine {
	windows, err := settings.ListScheduleWindows()
	if err != nil {
		logger.AppLogger().WithFields(map[string]interface{}{"error": err}).Warn("failed to load bandwidth schedule windows, proceeding without a schedule")
	}

	var checker policy.JellyfinChecker
	if cfg.Jellyfin.URL != "" {
		checker = jellyfin.New(jellyfin.Config{
			BaseURL: cfg.Jellyfin.URL,
			APIKey:  cfg.Jellyfin.APIKey,
			Logger:  logger.AppLogger(),
		})
	}

	engine := policy.New(cfg, windows, checker)
	pollInterval := time.Duration(cfg.Jellyfin.PlaybackPollIntervalSeconds) * time.Second
	engine.StartPolling(ctx, pollInterval)
	return engine
}

func printDryRunPlan(streams []*scheduler.Stream) {
	for _, s := range streams {
		tierLabel := "tier1"
		if s.Tier == scheduler.Tier2 {
			tierLabel = "tier2 (upgrade)"
		}
		fmt.Printf("[%s] %s\n", tierLabel, s.SourceKey)
		for _, item := range s.Items {
			resumeLabel := ""
			if item.ResumeInfo != nil {
				resumeLabel = " (resume)"
			}
			fmt.Printf("  Would download%s: %s\n", resumeLabel, item.DisplayName)
		}
	}
}

// downloadStats accumulates run statistics across concurrent workers.
type downloadStats struct {
	mu           sync.Mutex
	Total        int
	Downloaded   int
	Failed       int
	changedPaths []string

	// notifiedMovieIDs/notifiedSeriesIDs track, by Radarr movie ID / Sonarr
	// series ID, which distinct completed work units have already triggered a
	// post-download rescan notification this run (see shouldNotify).
	notifiedMovieIDs  map[int]struct{}
	notifiedSeriesIDs map[int]struct{}
}

// newDownloadStats builds a downloadStats ready to accumulate a run.
func newDownloadStats() *downloadStats {
	return &downloadStats{
		notifiedMovieIDs:  make(map[int]struct{}),
		notifiedSeriesIDs: make(map[int]struct{}),
	}
}

// recordItem records one item's outcome. When successful and changedPath is
// non-empty, the (movie or season) folder is added to this run's set of
// changed library paths for the later Jellyfin notification.
func (s *downloadStats) recordItem(success bool, changedPath string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Total++
	if success {
		s.Downloaded++
		if changedPath != "" {
			s.changedPaths = append(s.changedPaths, changedPath)
		}
	} else {
		s.Failed++
	}
}

// notifyKind distinguishes which per-run notified-ID set shouldNotify checks.
type notifyKind int

const (
	notifyKindMovie notifyKind = iota
	notifyKindSeries
)

// shouldNotify returns true and records id the first time it's called for
// that (kind, id) pair during this run, so a rescan is requested at most once
// per distinct completed movie/series regardless of how many of its items
// complete (see radarr-sonarr-post-download-rescan spec).
func (s *downloadStats) shouldNotify(kind notifyKind, id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	set := s.notifiedMovieIDs
	if kind == notifyKindSeries {
		set = s.notifiedSeriesIDs
	}
	if _, ok := set[id]; ok {
		return false
	}
	set[id] = struct{}{}
	return true
}

// movieRescanner is the subset of *radarr.Client the download command depends
// on to request a post-download rescan of a specific movie.
type movieRescanner interface {
	RescanMovie(ctx context.Context, movieID int) error
}

// seriesRescanner is the subset of *sonarr.Client the download command
// depends on to request a post-download rescan of a specific series.
type seriesRescanner interface {
	RescanSeries(ctx context.Context, seriesID int) error
}

// rescanNotifyTimeout bounds how long a single best-effort post-download
// rescan call may take, so an unreachable/slow Radarr or Sonarr cannot delay
// a worker beyond a bounded amount per completed item.
const rescanNotifyTimeout = 10 * time.Second

// runDownloadWorkerPool starts `parallel` goroutines that each claim a stream,
// drain it fully (in order), release it, and loop until the scheduler reports
// no streams remain. See media-download-scheduling spec: a claimed stream is
// drained to completion before its worker returns to the pool.
func runDownloadWorkerPool(ctx context.Context, sched *scheduler.Scheduler, dl *downloader.Downloader, cfg *config.Config, parallel int, verbose bool, radarr movieRescanner, sonarr seriesRescanner) *downloadStats {
	stats := newDownloadStats()

	var wg sync.WaitGroup
	for w := 0; w < parallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				// While the effective policy is "stop", no new stream is
				// claimed: the same as an empty queue. See "No New Claims
				// While Stopped".
				if dl.IsPolicyStopped() {
					return
				}

				stream, ok := sched.ClaimNext()
				if !ok {
					return
				}

				drainStream(ctx, dl, cfg, stream, stats, verbose, radarr, sonarr)

				if stream.SeriesID != 0 {
					sched.Release(stream.SeriesID)
				}
			}
		}()
	}
	wg.Wait()

	return stats
}

func drainStream(ctx context.Context, dl *downloader.Downloader, cfg *config.Config, stream *scheduler.Stream, stats *downloadStats, verbose bool, radarr movieRescanner, sonarr seriesRescanner) {
	for i := range stream.Items {
		// While the effective policy is "stop", no not-yet-started item is
		// begun; the remaining items in this stream are left for a later
		// run. See "No New Claims While Stopped".
		if dl.IsPolicyStopped() {
			return
		}

		outcome, changedPath := downloadItem(ctx, dl, cfg, &stream.Items[i], verbose)
		switch outcome {
		case itemDownloaded:
			stats.recordItem(true, changedPath)
			notifyRescanForCompletedItem(ctx, stats, stream, radarr, sonarr)
		case itemFailed:
			stats.recordItem(false, "")
		case itemPolicyStopped:
			// Not a failure: intentionally not recorded in stats (see
			// "Policy Abort Is Not a Failure"). The item was either not
			// started at all, or aborted mid-transfer; either way it
			// remains eligible for a later run once the policy clears.
			return
		}
	}
}

// notifyRescanForCompletedItem requests a targeted Radarr/Sonarr rescan for
// the movie/series that just had an item complete successfully, at most once
// per distinct movie/series for this run (see downloadStats.shouldNotify). A
// service that isn't configured for this run (nil) or whose ID is unknown
// (Stream.RadarrMovieID/SeriesID == 0, e.g. a stream synthesized without a
// fresh live lookup) is skipped. A rescan failure is logged and never affects
// stats or the run's exit status.
func notifyRescanForCompletedItem(ctx context.Context, stats *downloadStats, stream *scheduler.Stream, radarr movieRescanner, sonarr seriesRescanner) {
	if stream.RadarrMovieID != 0 && radarr != nil && stats.shouldNotify(notifyKindMovie, stream.RadarrMovieID) {
		rctx, cancel := context.WithTimeout(ctx, rescanNotifyTimeout)
		err := radarr.RescanMovie(rctx, stream.RadarrMovieID)
		cancel()
		if err != nil {
			logger.AppLogger().WithFields(map[string]interface{}{
				"radarr_movie_id": stream.RadarrMovieID,
				"error":           err,
			}).Warn("failed to notify radarr of completed download")
		}
	}

	if stream.SeriesID != 0 && sonarr != nil && stats.shouldNotify(notifyKindSeries, stream.SeriesID) {
		rctx, cancel := context.WithTimeout(ctx, rescanNotifyTimeout)
		err := sonarr.RescanSeries(rctx, stream.SeriesID)
		cancel()
		if err != nil {
			logger.AppLogger().WithFields(map[string]interface{}{
				"sonarr_series_id": stream.SeriesID,
				"error":            err,
			}).Warn("failed to notify sonarr of completed download")
		}
	}
}

// itemOutcome distinguishes a downloaded/failed item from one whose
// transfer was aborted by the adaptive-download-throttling effective
// policy - a distinct, non-failure outcome (see "Policy Abort Is Not a
// Failure").
type itemOutcome int

const (
	itemDownloaded itemOutcome = iota
	itemFailed
	itemPolicyStopped
)

// downloadItem attempts each candidate in quality-preference order and
// returns its outcome, and, when downloaded, the item's destination folder
// (movie folder, or season folder for a TV episode).
func downloadItem(ctx context.Context, dl *downloader.Downloader, cfg *config.Config, item *scheduler.Item, verbose bool) (itemOutcome, string) {
	db := database.Get()

	for j, candidate := range item.Candidates {
		if candidate.LineURL == nil || *candidate.LineURL == "" {
			continue
		}

		// While the effective policy is "stop", this not-yet-started
		// candidate is not attempted. See "No New Claims While Stopped".
		if dl.IsPolicyStopped() {
			return itemPolicyStopped, ""
		}

		fmt.Printf("Downloading: %s\n", item.DisplayName)

		if verbose {
			fmt.Printf("[%s] attempt %d/%d: %s\n", item.DisplayName, j+1, len(item.Candidates), *candidate.LineURL)
		}

		destPath := item.BaseDestDir + downloader.QualityTags(candidate.Resolution, candidate.Language, candidate.FrenchVariant)

		result, err := dl.Download(ctx, downloader.DownloadOptions{
			URL:             *candidate.LineURL,
			BaseDestPath:    destPath,
			TempDir:         cfg.Downloads.TempDir,
			ProcessedLineID: candidate.ID,
		})

		if err != nil {
			if errors.Is(err, policy.ErrStoppedByPolicy) {
				if verbose {
					fmt.Printf("[%s] attempt %d/%d aborted by download policy\n", item.DisplayName, j+1, len(item.Candidates))
				}
				return itemPolicyStopped, ""
			}

			db.Model(&models.ProcessedLine{}).Where("id = ?", candidate.ID).Update("state", models.StateFailed)
			if verbose {
				fmt.Printf("[%s] attempt %d/%d failed: %v\n", item.DisplayName, j+1, len(item.Candidates), err)
			}
			continue
		}

		fmt.Printf("Downloaded: %s -> %s (%s)\n", item.DisplayName, result.FilePath, formatBytes(result.FileSize))
		return itemDownloaded, filepath.Clean(filepath.Dir(result.FilePath))
	}

	return itemFailed, ""
}

// notifyBuildStreamsFailure sends a Critical notification identifying that
// this download run aborted because Radarr and/or Sonarr could not be
// reached while building the stream set. A delivery failure is logged and
// never affects the run's exit status.
func notifyBuildStreamsFailure(notif notifier.Notifier, err error) {
	event := notifier.Event{
		Severity: notifier.Critical,
		Title:    "Stalkeer: download run failed",
		Message:  fmt.Sprintf("Unable to reach Radarr/Sonarr while building the stream set: %v", err),
	}
	if notifyErr := notif.Notify(context.Background(), event); notifyErr != nil {
		logger.AppLogger().WithFields(map[string]interface{}{"error": notifyErr}).Warn("failed to send notification")
	}
}

// notifyDownloadRunResult sends a Warning notification summarizing this
// download run, only when at least one item failed to download. A run with
// zero failures sends nothing. A delivery failure is logged and never
// affects the run's exit status.
func notifyDownloadRunResult(notif notifier.Notifier, stats *downloadStats) {
	if stats.Failed == 0 {
		return
	}

	event := notifier.Event{
		Severity: notifier.Warning,
		Title:    "Stalkeer: download run completed with failures",
		Message:  fmt.Sprintf("Total: %d, Downloaded: %d, Failed: %d", stats.Total, stats.Downloaded, stats.Failed),
	}
	if err := notif.Notify(context.Background(), event); err != nil {
		logger.AppLogger().WithFields(map[string]interface{}{"error": err}).Warn("failed to send notification")
	}
}

// notifyJellyfin sends a single best-effort library-scan notification to
// Jellyfin for this run's deduplicated changed paths. Jellyfin is contacted
// only when it is enabled, configured with a base URL, and at least one path
// was collected. A failure to notify (misconfiguration, timeout, unreachable
// server, etc.) is logged as a warning and never affects the run's exit
// status or reported statistics.
func notifyJellyfin(ctx context.Context, cfg *config.Config, changedPaths []string) {
	if !cfg.Jellyfin.Enabled {
		return
	}
	if cfg.Jellyfin.URL == "" {
		logger.AppLogger().Warn("jellyfin integration is enabled but no url is configured, skipping library scan notification")
		return
	}

	paths := dedupePaths(changedPaths)
	if len(paths) == 0 {
		return
	}

	jellyfinClient := jellyfin.New(jellyfin.Config{
		BaseURL: cfg.Jellyfin.URL,
		APIKey:  cfg.Jellyfin.APIKey,
		Logger:  logger.AppLogger(),
	})

	if err := jellyfinClient.NotifyPathsUpdated(ctx, paths); err != nil {
		logger.AppLogger().WithFields(map[string]interface{}{
			"error":      err,
			"path_count": len(paths),
		}).Warn("failed to notify jellyfin of library changes")
	}
}

// dedupePaths cleans and deduplicates a set of collected paths, so a run with
// multiple episodes of the same season reports that season's folder once.
func dedupePaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(paths))
	deduped := make([]string, 0, len(paths))
	for _, p := range paths {
		clean := filepath.Clean(p)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		deduped = append(deduped, clean)
	}
	return deduped
}
