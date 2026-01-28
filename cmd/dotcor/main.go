package main

import (
	"fmt"
	"os"

	"github.com/justincordova/dotcor/internal/config"
	"github.com/justincordova/dotcor/internal/git"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "0.1.0"
)

const banner = `
     _       _
  __| | ___ | |_ ___ ___  _ __
 / _` + "`" + ` |/ _ \| __/ __/ _ \| '__|
| (_| | (_) | || (_| (_) | |
 \__,_|\___/ \__\___\___/|_|
`

func init() {
	viper.SetDefault("version", version)
}

var rootCmd = &cobra.Command{
	Use:   "dotcor",
	Short: "A simple, fast dotfile manager with symlinks and Git automation",
	Long: `DotCor combines the simplicity of GNU Stow with automatic Git commits.

Manage your dotfiles with symlinks - edit files directly, changes instantly
appear in your repository. Built-in Git automation handles commits and sync.`,
	Version: version,
	Run:     runRoot,
}

func runRoot(cmd *cobra.Command, args []string) {
	// Print banner
	fmt.Print("\033[36m") // Cyan color
	fmt.Print(banner)
	fmt.Print("\033[0m") // Reset color
	fmt.Printf("  v%s - Symlink-based dotfile manager\n\n", version)

	// Try to load config and show status
	cfg, err := config.LoadConfig()
	if err != nil {
		// Not initialized
		fmt.Println("\033[33m⚠ DotCor is not initialized\033[0m")
		fmt.Println("")
		fmt.Println("Get started:")
		fmt.Println("  dotcor init          Initialize DotCor")
		fmt.Println("  dotcor --help        Show all commands")
		fmt.Println("")
		return
	}

	// Show quick status
	showQuickStatus(cfg)
}

func showQuickStatus(cfg *config.Config) {
	files := cfg.GetManagedFilesForPlatform()
	totalFiles := len(files)

	// Count problems
	problemCount := 0
	for _, f := range files {
		fs := checkFileStatus(cfg, f)
		if fs.Status != "ok" {
			problemCount++
		}
	}

	// Files status
	if totalFiles == 0 {
		fmt.Println("📁 No files managed yet")
		fmt.Println("")
		fmt.Println("Get started:")
		fmt.Println("  dotcor add ~/.zshrc    Add a dotfile")
		fmt.Println("  dotcor list            List managed files")
		fmt.Println("")
	} else {
		if problemCount == 0 {
			fmt.Printf("\033[32m✓ %d file(s) managed, all healthy\033[0m\n", totalFiles)
		} else {
			fmt.Printf("\033[33m⚠ %d file(s) managed, %d with issues\033[0m\n", totalFiles, problemCount)
		}
	}

	// Git status
	repoPath, err := config.ExpandPath(cfg.RepoPath)
	if err == nil && git.IsGitInstalled() && git.IsRepo(repoPath) {
		gitStatus, err := git.GetStatus(repoPath)
		if err == nil {
			if gitStatus.HasUncommitted {
				fmt.Println("\033[33m⚠ Uncommitted changes\033[0m")
			} else {
				fmt.Println("\033[32m✓ Working tree clean\033[0m")
			}

			if gitStatus.RemoteExists {
				if gitStatus.AheadBy > 0 {
					fmt.Printf("\033[36m↑ %d commit(s) to push\033[0m\n", gitStatus.AheadBy)
				}
				if gitStatus.BehindBy > 0 {
					fmt.Printf("\033[36m↓ %d commit(s) to pull\033[0m\n", gitStatus.BehindBy)
				}
			}
		}
	}

	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  dotcor status     Full status")
	fmt.Println("  dotcor add        Add dotfiles")
	fmt.Println("  dotcor sync       Commit & push")
	fmt.Println("  dotcor --help     All commands")
	fmt.Println("")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
