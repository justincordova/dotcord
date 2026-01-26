package main

import (
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [file]",
	Short: "Show Git history for a dotfile",
	Args:  cobra.ExactArgs(1),
	Run:   runHistory,
}

func init() {
	historyCmd.Flags().Int("n", 10, "Number of commits to show")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) {
	// TODO: Implement
	// 1. Validate file is managed
	// 2. Get Git log for file
	// 3. Display formatted history
}
