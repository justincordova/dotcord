package main

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all managed dotfiles",
	Aliases: []string{"ls"},
	Run:     runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	// TODO: Implement
	// 1. Load config
	// 2. Display table of managed files
	// 3. Show count
}
