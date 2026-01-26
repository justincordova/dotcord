package main

import (
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize DotCor repository",
	Long:  `Creates ~/.dotcor directory structure and initializes Git repository.`,
	Run:   runInit,
}

func init() {
	initCmd.Flags().Bool("apply", false, "Create symlinks from existing config (for new machine setup)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	// TODO: Implement
	// 1. Check if ~/.dotcor exists
	// 2. Create directories: ~/.dotcor/ and ~/.dotcor/files/
	// 3. Initialize Git repo in files/
	// 4. Create default config.yaml
	// 5. If --apply flag, create symlinks from config
}
