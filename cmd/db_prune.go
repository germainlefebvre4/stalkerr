package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/parser"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// errDryRun is returned from within the pruning transaction to force a
// rollback when --dry-run is set, after the transaction's queries have
// already computed the counts that would otherwise have been applied.
var errDryRun = errors.New("dry run: rolling back")

var dbPruneCmd = &cobra.Command{
	Use:   "db-prune",
	Short: "Prune expired M3U stream URLs and orphaned metadata",
	Long: `Scan the database against every configured M3U source's currently downloaded
file, prune each source's expired processed lines (either soft or hard mode)
independently, and clean up orphaned movie and TV show metadata.`,
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

		// Get flags
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		hard, _ := cmd.Flags().GetBool("hard")

		fmt.Println("=== Database Pruning ===")
		if dryRun {
			fmt.Println("Mode: DRY RUN (no modifications will be made)")
		}
		if hard {
			fmt.Println("Pruning Type: HARD (will prune downloaded/downloading history)")
		} else {
			fmt.Println("Pruning Type: SOFT (will preserve downloaded/downloading history)")
		}
		fmt.Println()

		// Initialize database
		if err := database.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()
		db := database.Get()

		result, unprunableSources := pruneAllSources(db, log, cfg.M3U.Sources, hard, dryRun)

		if dryRun {
			fmt.Println("=== Simulation Results ===")
			fmt.Printf("Processed Lines to prune:  %d\n", result.linesPruned)
			fmt.Printf("Orphaned Movies to prune:   %d\n", result.moviesPruned)
			fmt.Printf("Orphaned TV Shows to prune: %d\n", result.tvShowsPruned)
			fmt.Println("\nPruning simulation complete. No records were deleted.")
		} else {
			fmt.Println("=== Pruning Complete ===")
			fmt.Printf("Pruned Processed Lines: %d\n", result.linesPruned)
			fmt.Printf("Pruned Movies:          %d\n", result.moviesPruned)
			fmt.Printf("Pruned TV Shows:        %d\n", result.tvShowsPruned)
			fmt.Println("\nDatabase pruned and metadata cleaned up successfully!")
		}

		if unprunableSources > 0 {
			fmt.Fprintf(os.Stderr, "\nError: %d source(s) could not be pruned\n", unprunableSources)
			os.Exit(1)
		}
	},
}

// pruneResult aggregates the counts of a db-prune run across every
// configured source.
type pruneResult struct {
	linesPruned   int64
	moviesPruned  int64
	tvShowsPruned int64
}

// pruneAllSources evaluates every configured M3U source independently,
// pruning each source's stale processed_lines against that source's own
// current file content, then cleans up any resulting orphaned metadata.
//
// A source whose file is missing or parses to zero entries is skipped with a
// warning rather than aborting the run; the number of such unprunable
// sources is returned alongside the aggregated result so the caller can
// decide the process exit code only after every source has been attempted.
//
// When dryRun is true, every query still runs against the real database
// (inside a transaction that is always rolled back), so the reported counts
// reflect exactly what a real run would prune without persisting any change.
func pruneAllSources(db *gorm.DB, log *logger.Logger, sources []config.M3USourceConfig, hard, dryRun bool) (pruneResult, int) {
	var result pruneResult
	unprunableSources := 0

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, source := range sources {
			fmt.Printf("--- Source %q ---\n", source.Name)
			fmt.Printf("M3U Playlist: %s\n", source.FilePath)

			activeHashes, ok := activeHashesForSource(source, log)
			if !ok {
				unprunableSources++
				fmt.Println()
				continue
			}
			fmt.Printf("Found %d active streams.\n", len(activeHashes))

			query := tx.Model(&models.ProcessedLine{}).
				Where("source_name = ?", source.Name).
				Where("line_hash NOT IN ?", activeHashes)
			if !hard {
				query = query.Where("state NOT IN ?", []string{
					string(models.StateDownloaded),
					string(models.StateDownloading),
					string(models.StateOrganizing),
				})
			}

			deleteResult := query.Delete(&models.ProcessedLine{})
			if deleteResult.Error != nil {
				return fmt.Errorf("failed to prune processed lines for source %q: %w", source.Name, deleteResult.Error)
			}
			fmt.Printf("Pruned lines: %d\n\n", deleteResult.RowsAffected)
			result.linesPruned += deleteResult.RowsAffected
		}

		prunedMovies, prunedTVShows, err := database.CleanupOrphanedMetadata(tx)
		if err != nil {
			return fmt.Errorf("failed to clean up orphaned metadata: %w", err)
		}
		result.moviesPruned = prunedMovies
		result.tvShowsPruned = prunedTVShows

		if dryRun {
			return errDryRun
		}
		return nil
	})

	if err != nil && !errors.Is(err, errDryRun) {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	return result, unprunableSources
}

// activeHashesForSource parses a configured source's current M3U file to
// collect its active line hashes. It returns ok=false (after printing a
// warning) when the file is missing, fails to parse, or parses to zero
// entries - any of which make this source unsafe to prune against right now.
func activeHashesForSource(source config.M3USourceConfig, log *logger.Logger) ([]string, bool) {
	if _, err := os.Stat(source.FilePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: M3U file '%s' for source %q does not exist, skipping. Run 'stalkeer m3u-download' first.\n", source.FilePath, source.Name)
		return nil, false
	}

	p := parser.NewParserWithLogger(source.FilePath, source.Name, log)
	lines, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error parsing M3U file for source %q, skipping: %v\n", source.Name, err)
		return nil, false
	}

	activeHashes := make([]string, len(lines))
	for i, line := range lines {
		activeHashes[i] = line.LineHash
	}

	if len(activeHashes) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: source %q's M3U file has 0 entries, skipping to prevent emptying its history\n", source.Name)
		return nil, false
	}

	return activeHashes, true
}

func init() {
	dbPruneCmd.Flags().Bool("dry-run", false, "simulate pruning and display metrics without deleting")
	dbPruneCmd.Flags().Bool("hard", false, "force delete downloaded and downloading stream records")
	rootCmd.AddCommand(dbPruneCmd)
}
