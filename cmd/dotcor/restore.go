package main

import (
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore a dotfile from Git history",
	Args:  cobra.ExactArgs(1),
	Run:   runRestore,
}

func init() {
	restoreCmd.Flags().String("to", "HEAD", "Git reference to restore from (e.g., HEAD~5, abc123)")
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) {
	// TODO: Implement
	// 1. Validate file is managed
	// 2. Get --to flag value
	// 3. Restore file from Git history
	// 4. Display success message
}
