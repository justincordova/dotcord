package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	version = "0.1.0"
)

func init() {
	// Initialize viper for later use
	viper.SetDefault("version", version)
}

var rootCmd = &cobra.Command{
	Use:   "dotcor",
	Short: "A simple, fast dotfile manager with symlinks and Git automation",
	Long: `DotCor combines the simplicity of GNU Stow with automatic Git commits.

Manage your dotfiles with symlinks - edit files directly, changes instantly
appear in your repository. Built-in Git automation handles commits and sync.`,
	Version: version,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
