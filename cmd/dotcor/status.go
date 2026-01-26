package main

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of managed dotfiles and repository",
	Run:   runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	// TODO: Implement
	// 1. Load config
	// 2. For each managed file, check symlink status
	// 3. Get Git repository status
	// 4. Display results
}
