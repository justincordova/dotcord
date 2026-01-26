package main

import (
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Commit all changes and push to remote",
	Run:   runSync,
}

func init() {
	syncCmd.Flags().Bool("no-push", false, "Commit but don't push to remote")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) {
	// TODO: Implement
	// 1. Check for deleted files
	// 2. Prompt to remove from config
	// 3. Commit all changes
	// 4. Push to remote (unless --no-push)
}
