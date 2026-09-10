package main

import (
	"fmt"
	"os"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/m3udownloader"
	"github.com/glefebvre/stalkeer/internal/processor"
	"github.com/spf13/cobra"
)

var processCmd = &cobra.Command{
	Use:   "process [m3u-file]",
	Short: "Process M3U file and store to database",
	Long: `Parse M3U playlist file, classify content, and store entries to the database.
This command performs full processing including content type detection and metadata
extraction. With no [m3u-file] argument, every configured M3U source (see m3u.sources)
is processed in turn, each tagged with its own source name; a missing source file is
skipped with a warning rather than aborting the run. Passing [m3u-file] bypasses the
configured source list entirely for a manual one-off run against that explicit file.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		if err := config.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		cfg := config.Get()

		// Initialize loggers with configured levels and format
		logger.InitializeLoggersWithFormat(cfg.GetAppLogLevel(), cfg.GetDatabaseLogLevel(), cfg.Logging.Format)
		log := logger.AppLogger()

		// Warn about legacy logging configuration
		if cfg.IsUsingLegacyLogging() {
			log.Warn("Using deprecated 'logging.level' configuration. Please migrate to 'logging.app.level' and 'logging.database.level' for better control.")
		}

		force, _ := cmd.Flags().GetBool("force")
		limit, _ := cmd.Flags().GetInt("limit")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		progress, _ := cmd.Flags().GetInt("progress")
		skipTMDB, _ := cmd.Flags().GetBool("skip-tmdb")
		tmdbLanguage, _ := cmd.Flags().GetString("tmdb-language")

		opts := processor.ProcessOptions{
			Force:            force,
			Limit:            limit,
			BatchSize:        batchSize,
			ProgressInterval: progress,
			SkipTMDB:         skipTMDB,
			TMDBLanguage:     tmdbLanguage,
		}

		if force {
			fmt.Println("Force mode: will re-process existing entries")
		}
		if limit > 0 {
			fmt.Printf("Processing limit: %d entries\n", limit)
		}
		if skipTMDB {
			fmt.Println("TMDB enrichment: disabled")
		} else if tmdbLanguage != "" {
			fmt.Printf("TMDB language: %s\n", tmdbLanguage)
		}
		fmt.Println()

		// Initialize database
		if err := database.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		// A single positional file path bypasses the configured source list
		// entirely for a manual one-off run against an explicit file, unchanged
		// from before multi-source support existed.
		if len(args) > 0 {
			filePath := args[0]

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: file '%s' does not exist\n", filePath)
				os.Exit(1)
			}

			fmt.Printf("Processing M3U file: %s\n", filePath)

			stats, err := runProcessTarget(filePath, "default", opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error processing file: %v\n", err)
				os.Exit(1)
			}

			printProcessStats("default", stats, skipTMDB)
			fmt.Println("\nProcessing completed successfully!")
			return
		}

		// Otherwise, process every configured source, isolating each source's
		// failure from the others.
		if processConfiguredSources(cfg.M3U.Sources, opts, log) {
			os.Exit(1)
		}

		fmt.Println("\nProcessing completed successfully!")
	},
}

// runProcessTarget creates a Processor for the given file/source and runs it.
func runProcessTarget(filePath, sourceName string, opts processor.ProcessOptions) (*processor.Statistics, error) {
	proc, err := processor.NewProcessor(filePath, sourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}
	return proc.Process(opts)
}

// processConfiguredSources processes every configured source independently.
// A source whose downloaded file is missing is skipped with a warning rather
// than aborting the run; a source that fails to process is logged and does
// not prevent the remaining sources from being attempted. Returns true if
// any source failed to process (a missing file is skipped, not a failure).
func processConfiguredSources(sources []config.M3USourceConfig, opts processor.ProcessOptions, log *logger.Logger) bool {
	hadError := false

	for i := range sources {
		source := &sources[i]
		filePath, _ := m3udownloader.SourcePaths(source.FilePath, source.Download.ArchiveDir, source.Name)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: file '%s' for source %q does not exist, skipping\n", filePath, source.Name)
			log.WithFields(map[string]interface{}{
				"source": source.Name,
				"file":   filePath,
			}).Warn("source file missing, skipping")
			continue
		}

		fmt.Printf("Processing M3U file: %s (source: %s)\n", filePath, source.Name)

		stats, err := runProcessTarget(filePath, source.Name, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing source %q: %v\n", source.Name, err)
			log.WithFields(map[string]interface{}{
				"source": source.Name,
				"error":  err,
			}).Error("failed to process source", err)
			hadError = true
			continue
		}

		printProcessStats(source.Name, stats, opts.SkipTMDB)
	}

	return hadError
}

// printProcessStats prints the processing summary for one source.
func printProcessStats(sourceName string, stats *processor.Statistics, skipTMDB bool) {
	fmt.Printf("\n=== Processing Complete (source: %s) ===\n", sourceName)
	fmt.Printf("Total lines in file:  %d\n", stats.TotalLines)
	fmt.Printf("Successfully processed: %d\n", stats.Processed)
	fmt.Printf("Duplicates skipped:   %d\n", stats.DuplicatesFound)
	fmt.Printf("Filtered out:         %d\n", stats.FilteredOut)
	fmt.Printf("Errors:               %d\n", stats.Errors)
	fmt.Printf("\nContent breakdown:\n")
	fmt.Printf("  Movies:        %d\n", stats.Movies)
	fmt.Printf("  TV Shows:      %d\n", stats.TVShows)
	fmt.Printf("  Channels:      %d\n", stats.Channels)
	fmt.Printf("  Uncategorized: %d\n", stats.Uncategorized)

	if !skipTMDB {
		fmt.Printf("\nTMDB Enrichment:\n")
		fmt.Printf("  Matched:       %d\n", stats.TMDBMatched)
		fmt.Printf("  Not found:     %d\n", stats.TMDBNotFound)
		fmt.Printf("  Errors:        %d\n", stats.TMDBErrors)
		if stats.TMDBMatched+stats.TMDBNotFound > 0 {
			matchRate := float64(stats.TMDBMatched) / float64(stats.TMDBMatched+stats.TMDBNotFound) * 100
			fmt.Printf("  Match rate:    %.1f%%\n", matchRate)
		}
		if stats.MetadataBackfilled > 0 || stats.MetadataBackfillErrors > 0 {
			fmt.Printf("  Metadata backfilled: %d\n", stats.MetadataBackfilled)
			fmt.Printf("  Metadata backfill errors: %d\n", stats.MetadataBackfillErrors)
		}
	}

	fmt.Printf("\nProcessing time: %v\n", stats.Duration)

	if stats.Errors > 0 {
		fmt.Printf("\nErrors encountered:\n")
		for i, msg := range stats.ErrorMessages {
			if i >= 10 {
				fmt.Printf("  ... and %d more errors\n", len(stats.ErrorMessages)-10)
				break
			}
			fmt.Printf("  - %s\n", msg)
		}
	}
}

func init() {
	processCmd.Flags().Bool("force", false, "re-process existing entries")
	processCmd.Flags().Int("limit", 0, "maximum number of items to process (0 = no limit)")
	processCmd.Flags().Int("batch-size", 100, "batch size for database inserts")
	processCmd.Flags().Int("progress", 1000, "show progress every N entries")
	processCmd.Flags().Bool("skip-tmdb", false, "skip TMDB metadata enrichment")
	processCmd.Flags().String("tmdb-language", "", "TMDB API language (e.g., 'en-US', 'fr-FR')")
	rootCmd.AddCommand(processCmd)
}
