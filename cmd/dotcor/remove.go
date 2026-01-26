package main

import (
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [file]",
	Short: "Stop managing a dotfile",
	Args:  cobra.ExactArgs(1),
	Run:   runRemove,
}

func init() {
	removeCmd.Flags().Bool("keep-file", false, "Keep file at source location after removing symlink")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) {
	// TODO: Implement
	// 1. Validate file is managed
	// 2. Prompt to remove symlink
	// 3. Prompt to delete from repo
	// 4. Copy file back if keeping
	// 5. Remove from config
	// 6. Git commit
}
