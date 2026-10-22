package main

import (
	"fmt"
	"os"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Initialize or update database schema to the latest version`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		if err := config.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Running database migrations...")

		mustLoadEffectiveConfig(true)
		defer database.Close()

		fmt.Println("Database migrations completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
