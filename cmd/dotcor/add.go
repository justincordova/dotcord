package main

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [file]...",
	Short: "Add a dotfile or directory to DotCord",
	Args:  cobra.MinimumNArgs(1),
	Run:   runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) {
	// TODO: Implement
	// For each file in args:
	// 1. Validate file exists
	// 2. Normalize source path
	// 3. Check if already managed
	// 4. Generate repo path
	// 5. Move file to repo
	// 6. Create symlink
	// 7. Add to config
	// 8. Git commit
}
