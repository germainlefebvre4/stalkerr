package main

import (
	"fmt"
	"os"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Surgically reset specific movie or TV show stream states",
	Long:  `Surgically reset individual movie or TV show entries by deleting their associated processed lines, allowing them to be re-downloaded/re-evaluated upon the next playlist import.`,
}

var resetMovieCmd = &cobra.Command{
	Use:   "movie",
	Short: "Reset a movie's stream state by ID",
	Long:  `Delete all processed line stream records for a specific movie, keeping its TMDB metadata record intact.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetUint("id")

		initDatabase()
		defer database.Close()
		db := database.Get()

		fmt.Printf("Resetting stream state for Movie ID: %d...\n", id)
		rows, err := database.ResetMovie(db, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resetting movie: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully reset movie ID %d. Removed %d stream line(s).\n", id, rows)
	},
}

var resetTVShowCmd = &cobra.Command{
	Use:   "tvshow",
	Short: "Reset a TV show's stream state by ID",
	Long:  `Delete all processed line stream records for a specific TV show, keeping its TMDB metadata record intact.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetUint("id")

		initDatabase()
		defer database.Close()
		db := database.Get()

		fmt.Printf("Resetting stream state for TV Show ID: %d...\n", id)
		rows, err := database.ResetTVShow(db, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resetting TV show: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully reset TV show ID %d. Removed %d stream line(s).\n", id, rows)
	},
}

func initDatabase() {
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Get()
	logger.InitializeLoggersWithFormat(cfg.GetAppLogLevel(), cfg.GetDatabaseLogLevel(), cfg.Logging.Format)

	if err := database.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	resetMovieCmd.Flags().Uint("id", 0, "ID of the movie to reset")
	_ = resetMovieCmd.MarkFlagRequired("id")

	resetTVShowCmd.Flags().Uint("id", 0, "ID of the TV show to reset")
	_ = resetTVShowCmd.MarkFlagRequired("id")

	resetCmd.AddCommand(resetMovieCmd)
	resetCmd.AddCommand(resetTVShowCmd)
	rootCmd.AddCommand(resetCmd)
}
