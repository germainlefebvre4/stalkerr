package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Stalkeer",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(versionString())
	},
}

// versionString formats the binary's embedded version, commit, and build
// date (see the version/commit/date vars in main.go).
func versionString() string {
	return fmt.Sprintf("Stalkeer %s\ncommit: %s\nbuilt:  %s", version, commit, date)
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
