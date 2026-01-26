package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/justincordova/dotcor/internal/config"
	"github.com/justincordova/dotcor/internal/core"
	"github.com/justincordova/dotcor/internal/fs"
	"github.com/justincordova/dotcor/internal/git"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of managed dotfiles and repository",
	Long: `Show comprehensive status of your DotCor setup.

Displays:
- Symlink health for each managed file
- Git repository status (uncommitted changes, remote sync)
- Overall statistics

Examples:
  dotcor status                # Show full status
  dotcor status --quick        # Show summary only
  dotcor status --problems     # Show only files with issues`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().BoolP("quick", "q", false, "Show summary only")
	statusCmd.Flags().Bool("problems", false, "Show only files with problems")
	statusCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	quick, _ := cmd.Flags().GetBool("quick")
	problemsOnly, _ := cmd.Flags().GetBool("problems")
	jsonFormat, _ := cmd.Flags().GetBool("json")

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w\nRun 'dotcor init' first", err)
	}

	// Collect status
	status := collectStatus(cfg)

	// Output
	if jsonFormat {
		return outputStatusJSON(status)
	}

	if quick {
		return outputStatusQuick(status)
	}

	return outputStatusFull(status, problemsOnly)
}

// StatusReport contains all status information
type StatusReport struct {
	Files      []FileStatus
	GitStatus  GitStatusInfo
	Statistics StatusStats
}

// FileStatus represents the status of a single managed file
type FileStatus struct {
	SourcePath string
	RepoPath   string
	Status     string
	Problem    string
}

// GitStatusInfo contains git-related status
type GitStatusInfo struct {
	IsRepo         bool
	HasUncommitted bool
	Branch         string
	AheadBy        int
	BehindBy       int
	RemoteExists   bool
}

// StatusStats contains summary statistics
type StatusStats struct {
	TotalFiles     int
	HealthyFiles   int
	ProblematicFiles int
}

// collectStatus gathers all status information
func collectStatus(cfg *config.Config) StatusReport {
	report := StatusReport{}

	// Get managed files
	files := cfg.GetManagedFilesForPlatform()
	report.Statistics.TotalFiles = len(files)

	// Check each file
	for _, f := range files {
		fs := checkFileStatus(cfg, f)
		report.Files = append(report.Files, fs)

		if fs.Status == "ok" {
			report.Statistics.HealthyFiles++
		} else {
			report.Statistics.ProblematicFiles++
		}
	}

	// Get git status
	repoPath, err := config.ExpandPath(cfg.RepoPath)
	if err == nil && git.IsGitInstalled() && git.IsRepo(repoPath) {
		gitStatus, _ := git.GetStatus(repoPath)
		report.GitStatus = GitStatusInfo{
			IsRepo:         true,
			HasUncommitted: gitStatus.HasUncommitted,
			Branch:         gitStatus.Branch,
			AheadBy:        gitStatus.AheadBy,
			BehindBy:       gitStatus.BehindBy,
			RemoteExists:   gitStatus.RemoteExists,
		}
	}

	return report
}

// checkFileStatus checks the status of a single managed file
func checkFileStatus(cfg *config.Config, mf config.ManagedFile) FileStatus {
	status := FileStatus{
		SourcePath: mf.SourcePath,
		RepoPath:   mf.RepoPath,
	}

	// Expand paths
	sourcePath, err := config.ExpandPath(mf.SourcePath)
	if err != nil {
		status.Status = "error"
		status.Problem = "invalid source path"
		return status
	}

	repoPath, err := config.GetRepoFilePath(cfg, mf.RepoPath)
	if err != nil {
		status.Status = "error"
		status.Problem = "invalid repo path"
		return status
	}

	// Check if repo file exists
	if !fs.FileExists(repoPath) {
		status.Status = "missing-repo"
		status.Problem = "file missing from repository"
		return status
	}

	// Check if source path exists
	if !fs.PathExists(sourcePath) {
		status.Status = "missing-source"
		status.Problem = "symlink missing"
		return status
	}

	// Check if it's a symlink
	isLink, err := fs.IsSymlink(sourcePath)
	if err != nil {
		status.Status = "error"
		status.Problem = fmt.Sprintf("error checking symlink: %v", err)
		return status
	}

	if !isLink {
		status.Status = "not-symlink"
		status.Problem = "source is a regular file, not a symlink"
		return status
	}

	// Check if symlink is valid
	valid, err := fs.IsValidSymlink(sourcePath)
	if err != nil {
		status.Status = "error"
		status.Problem = fmt.Sprintf("error validating symlink: %v", err)
		return status
	}

	if !valid {
		status.Status = "broken"
		status.Problem = "symlink target does not exist"
		return status
	}

	// Check if symlink points to correct target
	target, err := fs.ReadSymlink(sourcePath)
	if err != nil {
		status.Status = "error"
		status.Problem = fmt.Sprintf("error reading symlink: %v", err)
		return status
	}

	// Get expected target
	expectedRel, _ := config.ComputeRelativeSymlink(sourcePath, repoPath)

	// Compare (allowing both relative and absolute)
	if target != expectedRel && target != repoPath {
		// Try resolving relative path
		resolvedTarget := resolvePath(getDir(sourcePath), target)
		if resolvedTarget != repoPath {
			status.Status = "wrong-target"
			status.Problem = fmt.Sprintf("points to %s instead of repo file", target)
			return status
		}
	}

	status.Status = "ok"
	return status
}

// outputStatusFull outputs detailed status
func outputStatusFull(status StatusReport, problemsOnly bool) error {
	// Header
	fmt.Println("DotCor Status")
	fmt.Println("=============")
	fmt.Println("")

	// Files section
	if len(status.Files) > 0 {
		fmt.Println("Managed Files:")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		hasProblems := false
		for _, f := range status.Files {
			if problemsOnly && f.Status == "ok" {
				continue
			}

			icon := getStatusIcon(f.Status)
			if f.Status == "ok" {
				fmt.Fprintf(w, "  %s %s\tok\n", icon, f.SourcePath)
			} else {
				fmt.Fprintf(w, "  %s %s\t%s\n", icon, f.SourcePath, f.Problem)
				hasProblems = true
			}
		}

		w.Flush()

		if problemsOnly && !hasProblems {
			fmt.Println("  All files are healthy!")
		}

		fmt.Println("")
	}

	// Git section
	if status.GitStatus.IsRepo {
		fmt.Println("Git Repository:")

		if status.GitStatus.Branch != "" {
			fmt.Printf("  Branch: %s\n", status.GitStatus.Branch)
		}

		if status.GitStatus.HasUncommitted {
			fmt.Println("  ⚠ Uncommitted changes")
		} else {
			fmt.Println("  ✓ Working tree clean")
		}

		if status.GitStatus.RemoteExists {
			if status.GitStatus.AheadBy > 0 {
				fmt.Printf("  ↑ %d commit(s) ahead of remote\n", status.GitStatus.AheadBy)
			}
			if status.GitStatus.BehindBy > 0 {
				fmt.Printf("  ↓ %d commit(s) behind remote\n", status.GitStatus.BehindBy)
			}
			if status.GitStatus.AheadBy == 0 && status.GitStatus.BehindBy == 0 && !status.GitStatus.HasUncommitted {
				fmt.Println("  ✓ In sync with remote")
			}
		} else {
			fmt.Println("  - No remote configured")
		}

		fmt.Println("")
	}

	// Summary
	fmt.Printf("Summary: %d files managed", status.Statistics.TotalFiles)
	if status.Statistics.ProblematicFiles > 0 {
		fmt.Printf(", %d with issues", status.Statistics.ProblematicFiles)
	}
	fmt.Println("")

	// Suggestions
	if status.Statistics.ProblematicFiles > 0 {
		fmt.Println("")
		fmt.Println("Run 'dotcor doctor' for detailed diagnostics and repair suggestions.")
	}

	return nil
}

// outputStatusQuick outputs summary only
func outputStatusQuick(status StatusReport) error {
	// One-line summary
	if status.Statistics.ProblematicFiles == 0 {
		fmt.Printf("✓ %d files managed, all healthy\n", status.Statistics.TotalFiles)
	} else {
		fmt.Printf("⚠ %d files managed, %d with issues\n",
			status.Statistics.TotalFiles, status.Statistics.ProblematicFiles)
	}

	if status.GitStatus.IsRepo && status.GitStatus.HasUncommitted {
		fmt.Println("⚠ Uncommitted changes in repository")
	}

	return nil
}

// outputStatusJSON outputs status as JSON
func outputStatusJSON(status StatusReport) error {
	fmt.Println("{")
	fmt.Printf("  \"total_files\": %d,\n", status.Statistics.TotalFiles)
	fmt.Printf("  \"healthy_files\": %d,\n", status.Statistics.HealthyFiles)
	fmt.Printf("  \"problematic_files\": %d,\n", status.Statistics.ProblematicFiles)

	if status.GitStatus.IsRepo {
		fmt.Println("  \"git\": {")
		fmt.Printf("    \"branch\": \"%s\",\n", status.GitStatus.Branch)
		fmt.Printf("    \"uncommitted\": %t,\n", status.GitStatus.HasUncommitted)
		fmt.Printf("    \"ahead\": %d,\n", status.GitStatus.AheadBy)
		fmt.Printf("    \"behind\": %d,\n", status.GitStatus.BehindBy)
		fmt.Printf("    \"remote_exists\": %t\n", status.GitStatus.RemoteExists)
		fmt.Println("  },")
	}

	fmt.Println("  \"files\": [")
	for i, f := range status.Files {
		comma := ","
		if i == len(status.Files)-1 {
			comma = ""
		}
		problem := f.Problem
		if problem == "" {
			problem = "none"
		}
		fmt.Printf("    {\"source\": \"%s\", \"status\": \"%s\", \"problem\": \"%s\"}%s\n",
			f.SourcePath, f.Status, problem, comma)
	}
	fmt.Println("  ]")
	fmt.Println("}")

	return nil
}

// getStatusIcon returns an icon for the given status
func getStatusIcon(status string) string {
	switch status {
	case "ok":
		return "✓"
	case "missing-repo", "missing-source", "broken", "not-symlink", "wrong-target":
		return "✗"
	default:
		return "?"
	}
}

// CheckLockStatus checks if there's a stale lock (used by doctor)
func CheckLockStatus() (bool, *core.LockInfo, error) {
	info, err := core.GetLockInfo()
	if err != nil {
		return false, nil, err
	}

	if info == nil {
		return false, nil, nil // No lock
	}

	// Check if it's our own lock
	if info.PID == os.Getpid() {
		return false, info, nil // Our own lock
	}

	// Check if it's stale
	lockPath, _ := getLockPathForCheck()
	if lockPath != "" {
		stale, _ := core.IsStale(lockPath)
		if stale {
			return true, info, nil // Stale lock
		}
	}

	return false, info, nil // Active lock from another process
}

// getLockPathForCheck returns lock path for checking (internal use)
func getLockPathForCheck() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return configDir + "/.lock", nil
}
