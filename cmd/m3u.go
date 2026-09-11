package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/m3udownloader"
	"github.com/glefebvre/stalkeer/internal/notifier"
	"github.com/spf13/cobra"
)

var downloadM3UCmd = &cobra.Command{
	Use:   "m3u-download",
	Short: "Download M3U playlist from remote URL",
	Long: `Download M3U playlist file(s) from the configured URL(s) and save them to the
configured file path(s). Every configured source (see m3u.sources) is downloaded
and archived independently within this single run: a failure on one source is
logged and does not prevent the other sources from being attempted.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		if err := config.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		cfg := config.Get()

		// Initialize logger
		logger.InitializeLoggersWithFormat(cfg.GetAppLogLevel(), cfg.GetDatabaseLogLevel(), cfg.Logging.Format)
		log := logger.AppLogger()

		notif := notifier.FromConfig(cfg.Notifications)

		// Get flags
		urlOverride, _ := cmd.Flags().GetString("url")
		noArchive, _ := cmd.Flags().GetBool("no-archive")

		failures := downloadAllSources(cfg, log, notif, urlOverride, noArchive)
		if failures > 0 {
			fmt.Fprintf(os.Stderr, "\nError: %d source(s) failed to download\n", failures)
			os.Exit(1)
		}
	},
}

// downloadAllSources attempts to download and archive every configured M3U
// source independently. A failure on one source is logged and does not
// prevent the remaining sources from being attempted. Returns the number of
// sources that failed, so the caller can decide the process exit code only
// after every source has been attempted.
func downloadAllSources(cfg *config.Config, log *logger.Logger, notif notifier.Notifier, urlOverride string, noArchive bool) int {
	sources := cfg.M3U.Sources
	failures := 0

	for i := range sources {
		source := &sources[i]
		url := source.Download.URL
		// The --url flag only makes sense when there is exactly one source to
		// apply it to; with multiple configured sources it is ambiguous and is
		// ignored in favor of each source's own configured URL.
		if urlOverride != "" && len(sources) == 1 {
			url = urlOverride
		}

		if url == "" {
			fmt.Fprintf(os.Stderr, "Error: M3U download URL must be provided via --url flag or m3u.download.url in config for source %q\n", source.Name)
			failures++
			continue
		}

		destPath, archiveDir := m3udownloader.SourcePaths(source.FilePath, source.Download.ArchiveDir, source.Name)
		if destPath == "" {
			fmt.Fprintf(os.Stderr, "Error: M3U file path must be configured for source %q\n", source.Name)
			failures++
			continue
		}

		downloadCfg := source.Download
		downloadCfg.ArchiveDir = archiveDir

		fmt.Printf("Downloading M3U playlist for source %q...\n", source.Name)
		fmt.Printf("  Source URL:      %s\n", url)
		fmt.Printf("  Destination:     %s\n", destPath)
		if !noArchive {
			fmt.Printf("  Archive dir:     %s\n", archiveDir)
		}
		fmt.Println()

		dl := m3udownloader.NewDownloader(&downloadCfg, log)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(downloadCfg.TimeoutSeconds)*time.Second)
		var err error
		if noArchive {
			err = dl.Download(ctx, url, destPath)
		} else {
			err = dl.DownloadAndArchive(ctx, url, destPath)
		}
		cancel()

		if err != nil {
			notifyPlaylistFetchFailure(notif, source.Name, err)
			fmt.Fprintf(os.Stderr, "Error: Download failed for source %q: %v\n\n", source.Name, err)
			log.WithFields(map[string]interface{}{
				"source": source.Name,
				"error":  err,
			}).Error("M3U download failed for source", err)
			failures++
			continue
		}

		fmt.Printf("✓ M3U playlist downloaded successfully for source %q\n", source.Name)
		if info, err := os.Stat(destPath); err == nil {
			fmt.Printf("  Size: %s\n", formatBytes(info.Size()))
			fmt.Printf("  Modified: %s\n", info.ModTime().Format(time.RFC3339))
		}
		fmt.Println()
	}

	return failures
}

// notifyPlaylistFetchFailure sends a Critical notification identifying that
// an m3u-download run failed to fetch the configured playlist for the given
// source. A delivery failure is logged and never affects the run's exit
// status.
func notifyPlaylistFetchFailure(notif notifier.Notifier, sourceName string, err error) {
	event := notifier.Event{
		Severity: notifier.Critical,
		Title:    "Stalkeer: m3u-download run failed",
		Message:  fmt.Sprintf("Failed to fetch playlist for source %q: %v", sourceName, err),
	}
	if notifyErr := notif.Notify(context.Background(), event); notifyErr != nil {
		logger.AppLogger().WithFields(map[string]interface{}{"error": notifyErr}).Warn("failed to send notification")
	}
}

var listM3UArchivesCmd = &cobra.Command{
	Use:   "m3u-list-archives",
	Short: "List archived M3U playlist files",
	Long:  `Display a list of archived M3U playlist files with timestamps and sizes, for every configured source.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		if err := config.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		cfg := config.Get()

		// Initialize logger
		logger.InitializeLoggersWithFormat(cfg.GetAppLogLevel(), cfg.GetDatabaseLogLevel(), cfg.Logging.Format)
		log := logger.AppLogger()

		for _, entry := range listAllSourceArchives(cfg, log) {
			if entry.Err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to list archives for source %q: %v\n", entry.SourceName, entry.Err)
				continue
			}

			if len(entry.Archives) == 0 {
				fmt.Printf("No archived M3U files found for source %q in %s\n\n", entry.SourceName, entry.ArchiveDir)
				continue
			}

			fmt.Printf("Archived M3U files for source %q (%s):\n\n", entry.SourceName, entry.ArchiveDir)
			fmt.Printf("%-40s %-12s %s\n", "Filename", "Size", "Modified")
			fmt.Println(strings.Repeat("-", 80))

			for _, archive := range entry.Archives {
				fmt.Printf("%-40s %-12s %s\n",
					archive.Name,
					formatBytes(archive.SizeBytes),
					archive.ModTime.Format("2006-01-02 15:04:05"),
				)
			}

			fmt.Printf("\nTotal: %d archived files\n\n", len(entry.Archives))
		}
	},
}

// sourceArchives holds the archive listing result for one configured source.
type sourceArchives struct {
	SourceName string
	ArchiveDir string
	Archives   []m3udownloader.ArchiveInfo
	Err        error
}

// listAllSourceArchives lists archived M3U files for every configured
// source, each under its own per-source archive subdirectory.
func listAllSourceArchives(cfg *config.Config, log *logger.Logger) []sourceArchives {
	sources := cfg.M3U.Sources
	results := make([]sourceArchives, 0, len(sources))

	for i := range sources {
		source := &sources[i]
		_, archiveDir := m3udownloader.SourcePaths(source.FilePath, source.Download.ArchiveDir, source.Name)
		archiveManager := m3udownloader.NewArchiveManager(archiveDir, log)
		archives, err := archiveManager.ListArchiveFiles()
		results = append(results, sourceArchives{
			SourceName: source.Name,
			ArchiveDir: archiveDir,
			Archives:   archives,
			Err:        err,
		})
	}

	return results
}

var cleanupM3UArchivesCmd = &cobra.Command{
	Use:   "m3u-cleanup-archives",
	Short: "Clean up old M3U archive files",
	Long:  `Manually trigger rotation of M3U archive files for every configured source, keeping only each source's configured retention count.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		if err := config.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		cfg := config.Get()

		// Initialize logger
		logger.InitializeLoggersWithFormat(cfg.GetAppLogLevel(), cfg.GetDatabaseLogLevel(), cfg.Logging.Format)
		log := logger.AppLogger()

		// Get retention count override from flag (-1 means: use each source's own config value)
		retentionOverride, _ := cmd.Flags().GetInt("retention")

		if cleanupAllSourceArchives(cfg, log, retentionOverride) > 0 {
			os.Exit(1)
		}
	},
}

// cleanupAllSourceArchives rotates archived M3U files for every configured
// source independently. A failure rotating one source's archives is logged
// and does not prevent the remaining sources from being attempted. Returns
// the number of sources that failed to rotate.
func cleanupAllSourceArchives(cfg *config.Config, log *logger.Logger, retentionOverride int) int {
	sources := cfg.M3U.Sources
	failures := 0

	for i := range sources {
		source := &sources[i]
		retentionCount := source.Download.RetentionCount
		if retentionOverride >= 0 {
			retentionCount = retentionOverride
		}

		_, archiveDir := m3udownloader.SourcePaths(source.FilePath, source.Download.ArchiveDir, source.Name)
		archiveManager := m3udownloader.NewArchiveManager(archiveDir, log)

		archives, err := archiveManager.ListArchiveFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to list archives for source %q: %v\n", source.Name, err)
			failures++
			continue
		}

		fmt.Printf("Source %q archive directory: %s\n", source.Name, archiveDir)
		fmt.Printf("Current files:     %d\n", len(archives))
		fmt.Printf("Retention count:   %d\n", retentionCount)

		if len(archives) <= retentionCount {
			fmt.Println("No cleanup needed - archive count within retention limit")
			fmt.Println()
			continue
		}

		toDelete := len(archives) - retentionCount
		fmt.Printf("Files to delete:   %d\n", toDelete)

		if err := archiveManager.RotateArchive(retentionCount); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Cleanup failed for source %q: %v\n", source.Name, err)
			failures++
			continue
		}

		fmt.Printf("✓ Archive cleanup completed successfully for source %q\n\n", source.Name)
	}

	return failures
}

func init() {
	downloadM3UCmd.Flags().String("url", "", "M3U playlist URL (overrides config, single-source only)")
	downloadM3UCmd.Flags().Bool("no-archive", false, "skip creating archive copy")

	cleanupM3UArchivesCmd.Flags().Int("retention", -1, "number of archives to keep per source (default: use each source's config value)")

	rootCmd.AddCommand(downloadM3UCmd)
	rootCmd.AddCommand(listM3UArchivesCmd)
	rootCmd.AddCommand(cleanupM3UArchivesCmd)
}
