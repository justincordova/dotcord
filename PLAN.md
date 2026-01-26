# DotCor Implementation Plan

A detailed, step-by-step implementation plan for building DotCor - a symlink-based dotfile manager with Git automation.

---

## Overview

DotCor is a CLI-first dotfile manager built in Go. It uses **symlinks** to keep your dotfiles in a Git repository while making them accessible in their original locations.

**Core Design Decisions:**
- **Storage:** Symlink-based (files live in repo, symlinks point to them)
- **Git:** Automatic commits after every operation
- **Conflicts:** Let Git handle merges (we provide helpful messages)
- **Config:** YAML format with Viper
- **Paths:** Normalize to `~` for portability
- **Cross-platform:** macOS/Linux native, Windows with fallback to copy

**Key Differentiator:** GNU Stow's simplicity + Git automation built-in

---

## Project Structure

```
dotcor/
├── cmd/
│   └── dotcor/
│       ├── main.go           # Entry point + Cobra root command
│       ├── init.go           # dotcor init
│       ├── add.go            # dotcor add <file>
│       ├── remove.go         # dotcor remove <file>
│       ├── list.go           # dotcor list
│       ├── status.go         # dotcor status
│       ├── sync.go           # dotcor sync
│       ├── restore.go        # dotcor restore <file>
│       └── history.go        # dotcor history <file>
│
├── internal/
│   ├── config/
│   │   ├── config.go         # Config struct, Load/Save operations
│   │   └── paths.go          # Path normalization utilities
│   │
│   ├── core/
│   │   ├── linker.go         # Symlink creation/removal logic
│   │   └── validator.go      # File/path validation
│   │
│   ├── fs/
│   │   ├── fs.go             # File operations (move, copy fallback)
│   │   └── symlink.go        # Cross-platform symlink handling
│   │
│   └── git/
│       └── git.go            # Git command wrapper
│
├── go.mod
├── go.sum
├── PLAN.md                    # This file
└── README.md                  # Project documentation
```

---

## Data Model

### Config File Structure

**Location:** `~/.dotcor/config.yaml`

```yaml
repo_path: ~/.dotcor/files
git_enabled: true
git_remote: ""  # Optional remote URL

managed_files:
  - source_path: ~/.zshrc
    repo_path: shell/zshrc
    added_at: 2025-01-04T10:30:00Z
    platforms: []  # Empty = all platforms, or ["darwin", "linux", "windows"]

  - source_path: ~/.config/nvim/init.vim
    repo_path: nvim/init.vim
    added_at: 2025-01-04T10:31:00Z
    platforms: []
```

### Directory Layout

**Location:** `~/.dotcor/`

```
~/.dotcor/
├── config.yaml              # Metadata: which files are managed
└── files/                   # Git repository with actual dotfiles
    ├── .git/
    ├── shell/
    │   ├── zshrc            ← actual file
    │   └── bashrc           ← actual file
    └── nvim/
        └── init.vim         ← actual file

# System locations (symlinks):
~/.zshrc                     → symlink to ~/.dotcor/files/shell/zshrc
~/.bashrc                    → symlink to ~/.dotcor/files/shell/bashrc
~/.config/nvim/init.vim      → symlink to ~/.dotcor/files/nvim/init.vim
```

---

## Implementation Phases

### Phase 1: Core Infrastructure (Build First)

#### 1.1 Config Management (`internal/config/config.go`)

**Data structures:**

```go
package config

import "time"

type Config struct {
    RepoPath     string        `yaml:"repo_path"`      // ~/.dotcor/files
    GitEnabled   bool          `yaml:"git_enabled"`
    GitRemote    string        `yaml:"git_remote"`
    ManagedFiles []ManagedFile `yaml:"managed_files"`
}

type ManagedFile struct {
    SourcePath string    `yaml:"source_path"`  // ~/.zshrc (normalized, with ~)
    RepoPath   string    `yaml:"repo_path"`    // shell/zshrc (relative to files/)
    AddedAt    time.Time `yaml:"added_at"`
    Platforms  []string  `yaml:"platforms"`    // ["darwin", "linux"] or empty for all
}
```

**Required functions:**

```go
// LoadConfig loads config from ~/.dotcor/config.yaml
func LoadConfig() (*Config, error)

// SaveConfig writes config to ~/.dotcor/config.yaml
func (c *Config) SaveConfig() error

// AddManagedFile adds a new managed file
func (c *Config) AddManagedFile(mf ManagedFile) error

// RemoveManagedFile removes a managed file by source path
func (c *Config) RemoveManagedFile(sourcePath string) error

// GetManagedFile retrieves managed file by source path
func (c *Config) GetManagedFile(sourcePath string) (*ManagedFile, error)

// IsManaged checks if a file is already managed
func (c *Config) IsManaged(sourcePath string) bool

// GetManagedFilesForPlatform returns files that should be linked on current platform
func (c *Config) GetManagedFilesForPlatform() []ManagedFile
```

**Implementation notes:**
- Use Viper for YAML parsing
- Default repo path: `~/.dotcor/files`
- Handle missing config gracefully
- Platform detection: use `runtime.GOOS` (darwin, linux, windows)

---

#### 1.2 Path Utilities (`internal/config/paths.go`)

**Required functions:**

```go
// NormalizePath converts absolute path to ~ notation
// Example: /Users/you/.zshrc -> ~/.zshrc
func NormalizePath(path string) (string, error)

// ExpandPath converts ~ notation to absolute path
// Example: ~/.zshrc -> /Users/you/.zshrc
// Also handles environment variables: $XDG_CONFIG_HOME, %APPDATA%, etc.
func ExpandPath(path string) (string, error)

// GetRepoFilePath returns full path to file in repo
// Example: shell/zshrc -> ~/.dotcor/files/shell/zshrc
func GetRepoFilePath(config *Config, repoPath string) (string, error)

// GenerateRepoPath creates repo path from source path
// Example: ~/.config/nvim/init.vim -> nvim/init.vim
// Example: ~/.zshrc -> shell/zshrc
func GenerateRepoPath(sourcePath string) (string, error)

// GetCurrentPlatform returns current OS identifier
// Returns: "darwin", "linux", or "windows"
func GetCurrentPlatform() string

// ShouldApplyOnPlatform checks if file should be linked on current platform
func ShouldApplyOnPlatform(platforms []string) bool
```

**Repo path generation algorithm:**

```
Input: ~/.zshrc
1. Strip home directory: .zshrc
2. Determine category from filename: "shell"
3. Strip leading dot: zshrc
4. Output: shell/zshrc

Input: ~/.config/nvim/init.vim
1. Strip home directory: .config/nvim/init.vim
2. Strip .config/: nvim/init.vim
3. Output: nvim/init.vim

Input: ~/.gitconfig
1. Strip home directory: .gitconfig
2. Determine category from filename: "git"
3. Strip leading dot: gitconfig
4. Output: git/gitconfig
```

**Category mapping (for top-level dotfiles):**
```go
var categoryMap = map[string]string{
    ".zshrc":      "shell",
    ".bashrc":     "shell",
    ".bash_profile": "shell",
    ".gitconfig":  "git",
    ".gitignore":  "git",
    ".vimrc":      "vim",
    ".tmux.conf":  "tmux",
    // Add more as needed
}
```

**Implementation notes:**
- Use `os.UserHomeDir()` for home directory
- Use `filepath.Clean()` to normalize separators (handles / vs \ automatically)
- Support env variables: `os.ExpandEnv()` handles $VAR and %VAR%
- Handle edge cases: no home dir, already absolute, etc.

---

#### 1.3 Symlink Operations (`internal/fs/symlink.go`)

**Required functions:**

```go
// CreateSymlink creates a symlink from link to target
// On Windows: falls back to copying if symlink fails (no admin/dev mode)
func CreateSymlink(target, link string) error

// RemoveSymlink removes a symlink
func RemoveSymlink(link string) error

// IsSymlink checks if path is a symlink
func IsSymlink(path string) (bool, error)

// ReadSymlink reads the target of a symlink
func ReadSymlink(link string) (string, error)

// IsValidSymlink checks if symlink exists and points to existing target
func IsValidSymlink(link string) (bool, error)

// SupportsSymlinks checks if current platform supports symlinks
// Windows: requires admin rights or developer mode
func SupportsSymlinks() bool
```

**Cross-platform symlink handling:**

```go
func CreateSymlink(target, link string) error {
    // Ensure parent directory exists
    if err := EnsureDir(filepath.Dir(link)); err != nil {
        return err
    }

    // Try creating symlink
    err := os.Symlink(target, link)
    if err != nil {
        // On Windows, if symlink fails, fall back to copy
        if runtime.GOOS == "windows" {
            fmt.Println("⚠ Symlink failed, copying file instead")
            fmt.Println("  Enable Developer Mode for symlink support")
            return CopyFile(target, link)
        }
        return err
    }
    return nil
}
```

---

#### 1.4 File Operations (`internal/fs/fs.go`)

**Required functions:**

```go
// MoveFile moves a file from src to dst
func MoveFile(src, dst string) error

// CopyFile copies file with permissions preserved (Windows fallback)
func CopyFile(src, dst string) error

// FileExists checks if file exists
func FileExists(path string) bool

// EnsureDir creates directory if it doesn't exist
func EnsureDir(path string) error

// IsDirectory checks if path is a directory
func IsDirectory(path string) (bool, error)
```

**Implementation notes:**
- Preserve file permissions when copying/moving
- Create parent directories as needed
- Use `io.Copy` for efficient file copying

---

#### 1.5 Validation (`internal/core/validator.go`)

**Required functions:**

```go
// ValidateSourceFile checks if source file is valid for adding
func ValidateSourceFile(path string) error

// ValidateRepoPath checks if repo path is valid
func ValidateRepoPath(path string) error

// ValidateNotAlreadyManaged checks if file is not already managed
func ValidateNotAlreadyManaged(config *Config, sourcePath string) error
```

**Validation rules:**
- File must exist
- File must be readable
- Can be file or directory
- Path must be absolute or start with ~ or contain env variables
- Must not already be a symlink pointing to our repo

---

#### 1.6 Git Wrapper (`internal/git/git.go`)

**Required functions:**

```go
// InitRepo initializes git repository in directory
func InitRepo(repoPath string) error

// AutoCommit stages all changes and commits with message
func AutoCommit(repoPath, message string) error

// Sync commits all changes and pushes to remote (if configured)
func Sync(repoPath string) error

// HasChanges checks if working tree has uncommitted changes
func HasChanges(repoPath string) (bool, error)

// SetRemote configures git remote
func SetRemote(repoPath, remoteName, remoteURL string) error

// GetStatus returns git status information
func GetStatus(repoPath string) (StatusInfo, error)

// GetFileHistory returns git log for specific file
func GetFileHistory(repoPath, filePath string) ([]CommitInfo, error)

// RestoreFile restores file from git history
func RestoreFile(repoPath, filePath, ref string) error

// IsGitInstalled checks if git command is available
func IsGitInstalled() bool
```

**Status information:**

```go
type StatusInfo struct {
    HasUncommitted bool
    AheadBy        int
    BehindBy       int
    Branch         string
}

type CommitInfo struct {
    Hash      string
    Author    string
    Date      time.Time
    Message   string
}
```

**Implementation notes:**
- Use `os/exec.Command("git", ...)`
- Run git commands in repo directory with `cmd.Dir = repoPath`
- Capture stderr for error messages
- Check if git is installed at startup
- Auto-commit should be silent (don't spam output)

---

### Phase 2: Commands (Implement in Order)

#### 2.1 `dotcor init` (`cmd/dotcor/init.go`)

**What it does:**
1. Check if `~/.dotcor` already exists (prevent re-init)
2. Create directory structure:
   - `~/.dotcor/`
   - `~/.dotcor/files/`
3. Initialize Git repository in `~/.dotcor/files/`
4. Create default `config.yaml`
5. Optionally create symlinks if config already exists (from clone)
6. Success message

**Command definition:**

```go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize DotCor repository",
    Long:  `Creates ~/.dotcor directory structure and initializes Git repository.`,
    Run:   runInit,
}

func init() {
    initCmd.Flags().Bool("apply", false, "Create symlinks from existing config (for new machine setup)")
}
```

**Output example:**

```
✓ Created ~/.dotcor/
✓ Created ~/.dotcor/files/
✓ Initialized Git repository
✓ Created config.yaml

DotCor is ready! Next steps:
  dotcor add ~/.zshrc
  dotcor list
```

**With `--apply` flag (new machine setup):**

```
$ dotcor init --apply

✓ Found existing config
✓ Creating symlinks...
  ✓ ~/.zshrc → shell/zshrc
  ✓ ~/.bashrc → shell/bashrc
  ✓ ~/.config/nvim/init.vim → nvim/init.vim

DotCor setup complete!
```

**Error handling:**
- Already initialized: friendly message, exit
- Permission denied: clear error message
- Git not installed: warning (can work without Git, just no auto-commits)

---

#### 2.2 `dotcor add <file>` (`cmd/dotcor/add.go`)

**What it does:**
1. Validate file exists and is readable
2. Normalize source path (convert to ~ notation)
3. Check if already managed
4. Generate repo path
5. Move file to `~/.dotcor/files/{repo_path}` (or copy directory)
6. Create symlink from original location to repo
7. Add to config.yaml
8. Git commit: "Add {source_path}"
9. Success message

**Command definition:**

```go
var addCmd = &cobra.Command{
    Use:   "add [file]...",
    Short: "Add a dotfile or directory to DotCor",
    Args:  cobra.MinimumNArgs(1),
    Run:   runAdd,
}
```

**Output example:**

```
$ dotcor add ~/.zshrc

✓ Added ~/.zshrc
  Moved to: ~/.dotcor/files/shell/zshrc
  Symlink: ~/.zshrc → shell/zshrc
✓ Committed to Git

$ dotcor add ~/.config/nvim

✓ Added ~/.config/nvim
  Moved to: ~/.dotcor/files/nvim
  Symlink: ~/.config/nvim → nvim
✓ Committed to Git
```

**Error handling:**
- File doesn't exist
- File not readable
- Already managed (offer to update)
- Git commit fails (warn but don't fail)
- Symlink fails on Windows (fall back to copy, show warning)

---

#### 2.3 `dotcor list` (`cmd/dotcor/list.go`)

**What it does:**
1. Load config
2. Display managed files in table format
3. Show count and platform info

**Command definition:**

```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all managed dotfiles",
    Aliases: []string{"ls"},
    Run:   runList,
}
```

**Output example:**

```
Managed dotfiles (3):

SOURCE PATH                     REPO PATH              ADDED AT          PLATFORMS
~/.zshrc                        shell/zshrc            Jan 04 10:30      all
~/.config/nvim/init.vim         nvim/init.vim          Jan 04 10:31      all
~/Library/Preferences/foo.plist foo.plist              Jan 04 10:32      darwin
```

**Error handling:**
- No managed files: friendly message
- Config not found: suggest running `dotcor init`

---

#### 2.4 `dotcor status` (`cmd/dotcor/status.go`)

**What it does:**
1. Load config
2. For each managed file:
   - Check if symlink exists
   - Check if symlink target exists
   - Check if symlink points to correct location
3. Show Git repository status
4. Display status for each file + overall repo status

**Command definition:**

```go
var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show status of managed dotfiles and repository",
    Run:   runStatus,
}
```

**Output example:**

```
Symlinks:
✓ ~/.zshrc                 → shell/zshrc
✓ ~/.bashrc                → shell/bashrc
✗ ~/.vimrc                 → vim/vimrc (broken: target missing)
! ~/.config/nvim/init.vim  → (not linked - file exists but not a symlink)

Repository:
● 2 uncommitted changes
↑ 1 commit ahead of origin/main

Run 'dotcor sync' to commit and push changes
```

**Legend:**
- ✓ = symlink exists and target exists
- ✗ = symlink broken (target missing or not pointing to repo)
- ! = file exists at source location but not a symlink (conflict)

---

#### 2.5 `dotcor remove <file>` (`cmd/dotcor/remove.go`)

**What it does:**
1. Validate file is managed
2. Prompt: "Remove symlink? [y/N]"
3. If yes: remove symlink
4. Prompt: "Delete from repository? [y/N]"
5. If yes: delete from `~/.dotcor/files/`
6. Remove from config.yaml
7. Git commit: "Remove {source_path}"
8. Optionally restore file to original location (if not deleted from repo)

**Command definition:**

```go
var removeCmd = &cobra.Command{
    Use:   "remove [file]",
    Short: "Stop managing a dotfile",
    Args:  cobra.ExactArgs(1),
    Run:   runRemove,
}

func init() {
    removeCmd.Flags().Bool("keep-file", false, "Keep file at source location after removing symlink")
}
```

**Output example:**

```
$ dotcor remove ~/.zshrc

? Remove symlink ~/.zshrc? [y/N]: y
✓ Symlink removed

? Delete from repository? [y/N]: n
✓ File kept in repository at shell/zshrc
✓ Copied back to ~/.zshrc

✓ Removed from config
✓ Committed to Git
```

**With `--keep-file` flag:**

```
$ dotcor remove ~/.zshrc --keep-file

✓ Symlink removed
✓ Copied back to ~/.zshrc
? Delete from repository? [y/N]: y
✓ Deleted from repository
✓ Committed to Git
```

---

#### 2.6 `dotcor sync` (`cmd/dotcor/sync.go`)

**What it does:**
1. Check for deleted files (symlinks pointing to nowhere)
2. Prompt to remove deleted files from config
3. Commit all changes with message: "Sync dotfiles - {date}"
4. Push to remote (if configured)
5. Summary

**Command definition:**

```go
var syncCmd = &cobra.Command{
    Use:   "sync",
    Short: "Commit all changes and push to remote",
    Run:   runSync,
}

func init() {
    syncCmd.Flags().Bool("no-push", false, "Commit but don't push to remote")
}
```

**Output example:**

```
$ dotcor sync

Checking for changes...
✓ ~/.zshrc (modified)
! ~/.tmux.conf (deleted from system)

? ~/.tmux.conf was deleted. Remove from dotcor? [y/N]: y

✓ Committed 2 changes: "Sync dotfiles - 2025-01-04 15:30"
✓ Pushed to origin/main
```

**Error handling:**
- No remote configured: show helpful message
- Push fails: show Git error
- Merge conflicts: tell user to resolve manually

---

#### 2.7 `dotcor restore <file>` (`cmd/dotcor/restore.go`)

**What it does:**
1. Validate file is managed
2. Restore file from Git history (default: HEAD)
3. Optionally restore from specific commit
4. Success message

**Command definition:**

```go
var restoreCmd = &cobra.Command{
    Use:   "restore [file]",
    Short: "Restore a dotfile from Git history",
    Args:  cobra.ExactArgs(1),
    Run:   runRestore,
}

func init() {
    restoreCmd.Flags().String("to", "HEAD", "Git reference to restore from (e.g., HEAD~5, abc123)")
}
```

**Output example:**

```
$ dotcor restore ~/.zshrc

✓ Restored ~/.zshrc from HEAD

$ dotcor restore ~/.zshrc --to=HEAD~5

✓ Restored ~/.zshrc from HEAD~5 (5 commits ago)
```

---

#### 2.8 `dotcor history <file>` (`cmd/dotcor/history.go`)

**What it does:**
1. Validate file is managed
2. Show Git log for that file
3. Display in readable format

**Command definition:**

```go
var historyCmd = &cobra.Command{
    Use:   "history [file]",
    Short: "Show Git history for a dotfile",
    Args:  cobra.ExactArgs(1),
    Run:   runHistory,
}

func init() {
    historyCmd.Flags().Int("n", 10, "Number of commits to show")
}
```

**Output example:**

```
$ dotcor history ~/.zshrc

History for ~/.zshrc (shell/zshrc):

abc123f - 2025-01-04 15:30 - Update zsh aliases
def456a - 2025-01-03 09:15 - Add new PATH entries
789beef - 2025-01-02 14:22 - Sync dotfiles
...

Use 'dotcor restore ~/.zshrc --to=<commit>' to restore
```

---

### Phase 3: Testing & Polish

#### 3.1 Manual Testing Flow

```bash
# Initialize
dotcor init

# Add some files
dotcor add ~/.zshrc ~/.bashrc
dotcor add ~/.config/nvim
dotcor list

# Verify symlinks created
ls -la ~/.zshrc  # Should show: .zshrc -> /Users/you/.dotcor/files/shell/zshrc

# Edit a file (changes immediately reflected in repo)
echo "alias test='echo test'" >> ~/.zshrc

# Check status
dotcor status  # Should show uncommitted changes

# Sync changes
dotcor sync

# Test restore
dotcor history ~/.zshrc
dotcor restore ~/.zshrc --to=HEAD~1

# Test on "new machine" (simulate)
dotcor remove ~/.zshrc --keep-file
rm ~/.zshrc
dotcor add ~/.zshrc  # Should create symlink
```

#### 3.2 Edge Cases to Test

- Add non-existent file (should error)
- Add already-managed file (should prompt to update)
- Add directory vs single file
- Remove file that's not managed (should error)
- Init when already initialized (should warn)
- Operations without Git installed (should work, skip Git)
- Symlink on Windows without dev mode (should fall back to copy)
- Platform-specific files (only link on correct platform)
- Sync when no changes (should skip)
- Sync when no remote configured (should warn)

#### 3.3 Polish Items

- Consistent error messages
- Help text for all commands
- Color output using a library like fatih/color or charmbracelet/lipgloss
- Progress indicators for long operations
- Validate Git is installed (warn if not)
- Shell completion (Cobra supports this)
- Man page generation
- Version command

---

## Development Order Checklist

**Infrastructure (build first):**
- [ ] `internal/config/config.go` - Config structs and Viper integration
- [ ] `internal/config/paths.go` - Path normalization and repo path generation
- [ ] `internal/fs/symlink.go` - Cross-platform symlink handling
- [ ] `internal/fs/fs.go` - File operations (move, copy, etc.)
- [ ] `internal/core/validator.go` - Validation logic
- [ ] `internal/git/git.go` - Git wrapper

**Commands (build in order):**
- [ ] `cmd/dotcor/main.go` - Cobra setup
- [ ] `cmd/dotcor/init.go` - Initialize DotCor
- [ ] `cmd/dotcor/add.go` - Add files/directories
- [ ] `cmd/dotcor/list.go` - List managed files
- [ ] `cmd/dotcor/status.go` - Show status
- [ ] `cmd/dotcor/remove.go` - Remove files
- [ ] `cmd/dotcor/sync.go` - Sync changes
- [ ] `cmd/dotcor/restore.go` - Restore from history
- [ ] `cmd/dotcor/history.go` - Show Git history

**Testing:**
- [ ] Manual end-to-end test flow
- [ ] Edge case testing
- [ ] Cross-platform testing (macOS, Linux, Windows)
- [ ] Polish and error messages

---

## Future Enhancements (Post-MVP)

### v1.0 - MVP (Current)
- Core symlink-based management
- Git auto-commit
- Cross-platform support
- Basic restore/history

### v2.0 - Enhanced Workflow
- Watch mode: `dotcor watch` to auto-sync on file changes
- Template support: basic variable substitution `{{ .hostname }}`
- Hooks: run commands before/after operations
- Batch operations: `dotcor add ~/.config/*`

### v3.0 - Power Features
- Machine profiles (work, home, server)
- Encrypted secrets integration (age, gpg)
- Package manager integration (export Brewfile, apt list, etc.)
- TUI interface (using charmbracelet/bubbletea)

### v4.0 - Advanced
- Desktop GUI (Tauri or Wails)
- Plugin system
- Cloud sync (beyond Git)
- Migration tools (from chezmoi, stow, etc.)

---

## Resources

- [Cobra Documentation](https://github.com/spf13/cobra)
- [Viper Documentation](https://github.com/spf13/viper)
- [Go YAML v3](https://github.com/go-yaml/yaml)
- [GNU Stow](https://www.gnu.org/software/stow/)
- [Go filepath package](https://pkg.go.dev/path/filepath) - Cross-platform path handling
