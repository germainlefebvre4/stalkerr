package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/glefebvre/stalkeer/internal/api"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/retry"
	"github.com/glefebvre/stalkeer/internal/scheduler"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

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
		cfg := config.Get()

		if parallel <= 0 {
			parallel = cfg.Downloads.MaxParallel
		}
		if parallel <= 0 {
			parallel = 3
		}

		logger.InitializeLoggers(cfg.GetAppLogLevel(), cfg.GetDatabaseLogLevel())

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

		var radarrClient scheduler.RadarrClient
		var radarrFullClient *radarr.Client
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
			})
			radarrClient = radarrFullClient
		} else if verbose {
			fmt.Println("Radarr is not configured, skipping movie fetch")
		}

		var sonarrClient scheduler.SonarrClient
		var sonarrFullClient *sonarr.Client
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
			})
			sonarrClient = sonarrFullClient
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
		stats := runDownloadWorkerPool(ctx, sched, dl, cfg, parallel, verbose)

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
	mu         sync.Mutex
	Total      int
	Downloaded int
	Failed     int
}

func (s *downloadStats) recordItem(success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Total++
	if success {
		s.Downloaded++
	} else {
		s.Failed++
	}
}

// runDownloadWorkerPool starts `parallel` goroutines that each claim a stream,
// drain it fully (in order), release it, and loop until the scheduler reports
// no streams remain. See media-download-scheduling spec: a claimed stream is
// drained to completion before its worker returns to the pool.
func runDownloadWorkerPool(ctx context.Context, sched *scheduler.Scheduler, dl *downloader.Downloader, cfg *config.Config, parallel int, verbose bool) *downloadStats {
	stats := &downloadStats{}

	var wg sync.WaitGroup
	for w := 0; w < parallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				stream, ok := sched.ClaimNext()
				if !ok {
					return
				}

				drainStream(ctx, dl, cfg, stream, stats, verbose)

				if stream.SeriesID != 0 {
					sched.Release(stream.SeriesID)
				}
			}
		}()
	}
	wg.Wait()

	return stats
}

func drainStream(ctx context.Context, dl *downloader.Downloader, cfg *config.Config, stream *scheduler.Stream, stats *downloadStats, verbose bool) {
	for i := range stream.Items {
		success := downloadItem(ctx, dl, cfg, &stream.Items[i], verbose)
		stats.recordItem(success)
	}
}

func downloadItem(ctx context.Context, dl *downloader.Downloader, cfg *config.Config, item *scheduler.Item, verbose bool) bool {
	db := database.Get()

	for j, candidate := range item.Candidates {
		if candidate.LineURL == nil || *candidate.LineURL == "" {
			continue
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
			db.Model(&models.ProcessedLine{}).Where("id = ?", candidate.ID).Update("state", models.StateFailed)
			if verbose {
				fmt.Printf("[%s] attempt %d/%d failed: %v\n", item.DisplayName, j+1, len(item.Candidates), err)
			}
			continue
		}

		fmt.Printf("Downloaded: %s -> %s (%s)\n", item.DisplayName, result.FilePath, formatBytes(result.FileSize))
		return true
	}

	return false
}
