package main

import (
	"fmt"
	"os"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/parser"
	"github.com/spf13/cobra"
)

var dbPruneCmd = &cobra.Command{
	Use:   "db-prune",
	Short: "Prune expired M3U stream URLs and orphaned metadata",
	Long: `Scan the database against the currently active M3U file, prune expired 
processed lines (either soft or hard mode), and clean up orphaned movie and TV show metadata.`,
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

		filePath := cfg.M3U.FilePath
		if filePath == "" {
			fmt.Fprintln(os.Stderr, "Error: M3U file path must be configured in m3u.file_path")
			os.Exit(1)
		}

		// Check if M3U file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: M3U file '%s' does not exist. Run 'stalkeer m3u-download' first.\n", filePath)
			os.Exit(1)
		}

		fmt.Println("=== Database Pruning ===")
		fmt.Printf("M3U Playlist: %s\n", filePath)
		if dryRun {
			fmt.Println("Mode: DRY RUN (no modifications will be made)")
		}
		if hard {
			fmt.Println("Pruning Type: HARD (will prune downloaded/downloading history)")
		} else {
			fmt.Println("Pruning Type: SOFT (will preserve downloaded/downloading history)")
		}
		fmt.Println()

		// Parse the M3U file to extract active hashes
		fmt.Println("Parsing active M3U playlist to collect current hashes...")
		p := parser.NewParserWithLogger(filePath, log)
		lines, err := p.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing M3U: %v\n", err)
			os.Exit(1)
		}

		activeHashes := make([]string, len(lines))
		for i, line := range lines {
			activeHashes[i] = line.LineHash
		}
		fmt.Printf("Found %d active streams in playlist.\n\n", len(activeHashes))

		if len(activeHashes) == 0 {
			fmt.Fprintln(os.Stderr, "Error: Active M3U file has 0 entries. Aborting prune to prevent empty database.")
			os.Exit(1)
		}

		// Initialize database
		if err := database.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()
		db := database.Get()

		if dryRun {
			// Dry run: simulate counts
			var lineCount int64
			var movieCount int64
			var tvshowCount int64

			// Count lines that would be deleted
			query := db.Table("processed_lines").Where("line_hash NOT IN ?", activeHashes)
			if !hard {
				query = query.Where("state NOT IN ?", []string{"downloaded", "downloading", "organizing"})
			}
			if err := query.Count(&lineCount).Error; err != nil {
				fmt.Fprintf(os.Stderr, "Error counting lines: %v\n", err)
				os.Exit(1)
			}

			// Count movies that would become orphaned and deleted
			var movieQuery string
			if hard {
				movieQuery = `
					SELECT COUNT(*) FROM movies 
					WHERE id NOT IN (
						SELECT DISTINCT movie_id 
						FROM processed_lines 
						WHERE movie_id IS NOT NULL AND line_hash IN (?)
					)
				`
			} else {
				movieQuery = `
					SELECT COUNT(*) FROM movies 
					WHERE id NOT IN (
						SELECT DISTINCT movie_id 
						FROM processed_lines 
						WHERE movie_id IS NOT NULL 
						  AND (line_hash IN (?) OR state IN ('downloaded', 'downloading', 'organizing'))
					)
				`
			}
			if err := db.Raw(movieQuery, activeHashes).Scan(&movieCount).Error; err != nil {
				fmt.Fprintf(os.Stderr, "Error counting orphaned movies: %v\n", err)
				os.Exit(1)
			}

			// Count tvshows that would become orphaned and deleted
			var tvshowQuery string
			if hard {
				tvshowQuery = `
					SELECT COUNT(*) FROM tvshows 
					WHERE id NOT IN (
						SELECT DISTINCT tv_show_id 
						FROM processed_lines 
						WHERE tv_show_id IS NOT NULL AND line_hash IN (?)
					)
				`
			} else {
				tvshowQuery = `
					SELECT COUNT(*) FROM tvshows 
					WHERE id NOT IN (
						SELECT DISTINCT tv_show_id 
						FROM processed_lines 
						WHERE tv_show_id IS NOT NULL 
						  AND (line_hash IN (?) OR state IN ('downloaded', 'downloading', 'organizing'))
					)
				`
			}
			if err := db.Raw(tvshowQuery, activeHashes).Scan(&tvshowCount).Error; err != nil {
				fmt.Fprintf(os.Stderr, "Error counting orphaned TV shows: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("=== Simulation Results ===")
			fmt.Printf("Processed Lines to prune:  %d\n", lineCount)
			fmt.Printf("Orphaned Movies to prune:   %d\n", movieCount)
			fmt.Printf("Orphaned TV Shows to prune: %d\n", tvshowCount)
			fmt.Println("\nPruning simulation complete. No records were deleted.")
			return
		}

		// Perform actual prune
		prunedLines, err := database.PruneProcessedLines(db, activeHashes, hard)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error pruning processed lines: %v\n", err)
			os.Exit(1)
		}

		prunedMovies, prunedTVShows, err := database.CleanupOrphanedMetadata(db)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning up metadata: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("=== Pruning Complete ===")
		fmt.Printf("Pruned Processed Lines: %d\n", prunedLines)
		fmt.Printf("Pruned Movies:          %d\n", prunedMovies)
		fmt.Printf("Pruned TV Shows:        %d\n", prunedTVShows)
		fmt.Println("\nDatabase pruned and metadata cleaned up successfully!")
	},
}

func init() {
	dbPruneCmd.Flags().Bool("dry-run", false, "simulate pruning and display metrics without deleting")
	dbPruneCmd.Flags().Bool("hard", false, "force delete downloaded and downloading stream records")
	rootCmd.AddCommand(dbPruneCmd)
}
