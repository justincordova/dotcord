# DotMan Implementation Plan

A detailed, step-by-step implementation plan for building DotMan CLI dotfile manager.

---

## Overview

DotMan is a CLI-first dotfile manager built in Go. It tracks dotfiles in a Git repository, backs them up safely, and applies them across machines.

**Core Design Decisions:**
- **Storage:** Copy-based (repo stores copies, not symlinks)
- **Git:** Automatic initialization and commits
- **Conflicts:** Backup + show diff + user prompt
- **Config:** YAML format with Viper
- **Paths:** Normalize to `~` for portability

---

## Project Structure

```
dotman/
├── cmd/
│   └── dotman/
│       ├── main.go           # Entry point + Cobra root command
│       ├── init.go           # dotman init
│       ├── track.go          # dotman track <file>
│       ├── untrack.go        # dotman untrack <file>
│       ├── list.go           # dotman list
│       ├── status.go         # dotman status
│       ├── apply.go          # dotman apply
│       ├── push.go           # dotman push
│       └── pull.go           # dotman pull
│
├── internal/
│   ├── config/
│   │   ├── config.go         # Config struct, Load/Save operations
│   │   └── paths.go          # Path normalization utilities
│   │
│   ├── core/
│   │   ├── tracker.go        # Track/untrack business logic
│   │   ├── applier.go        # Apply logic with conflict resolution
│   │   └── differ.go         # File comparison and diff display
│   │
│   ├── fs/
│   │   ├── fs.go             # File operations (copy, backup)
│   │   └── validate.go       # File validation
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

**Location:** `~/.dotman/config.yaml`

```yaml
repo_path: /Users/you/.dotman
backup_path: /Users/you/.dotman/backups
git_enabled: true
git_remote: ""  # Optional remote URL

tracked_files:
  - source_path: ~/.zshrc
    repo_path: zsh/zshrc
    tracked_at: 2025-01-04T10:30:00Z

  - source_path: ~/.config/nvim/init.vim
    repo_path: nvim/init.vim
    tracked_at: 2025-01-04T10:31:00Z
```

### Directory Layout

**Location:** `~/.dotman/`

```
~/.dotman/
├── config.yaml              # Metadata and tracked files
├── files/                   # Git repository storing dotfiles
│   ├── .git/
│   ├── zsh/
│   │   └── zshrc
│   └── nvim/
│       └── init.vim
└── backups/                 # Timestamped backups
    └── 2025-01-04_103000_zshrc
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
    RepoPath     string        `yaml:"repo_path"`
    BackupPath   string        `yaml:"backup_path"`
    GitEnabled   bool          `yaml:"git_enabled"`
    GitRemote    string        `yaml:"git_remote"`
    TrackedFiles []TrackedFile `yaml:"tracked_files"`
}

type TrackedFile struct {
    SourcePath string    `yaml:"source_path"`  // ~/.zshrc (normalized)
    RepoPath   string    `yaml:"repo_path"`    // zsh/zshrc (relative to files/)
    TrackedAt  time.Time `yaml:"tracked_at"`
}
```

**Required functions:**

```go
// LoadConfig loads config from ~/.dotman/config.yaml
func LoadConfig() (*Config, error)

// SaveConfig writes config to ~/.dotman/config.yaml
func (c *Config) SaveConfig() error

// AddTrackedFile adds a new tracked file
func (c *Config) AddTrackedFile(tf TrackedFile) error

// RemoveTrackedFile removes a tracked file by source path
func (c *Config) RemoveTrackedFile(sourcePath string) error

// GetTrackedFile retrieves tracked file by source path
func (c *Config) GetTrackedFile(sourcePath string) (*TrackedFile, error)

// IsTracked checks if a file is already tracked
func (c *Config) IsTracked(sourcePath string) bool
```

**Implementation notes:**
- Use Viper for YAML parsing
- Default paths: `~/.dotman`, `~/.dotman/backups`
- Handle missing config gracefully

---

#### 1.2 Path Utilities (`internal/config/paths.go`)

**Required functions:**

```go
// NormalizePath converts absolute path to ~ notation
// Example: /Users/you/.zshrc -> ~/.zshrc
func NormalizePath(path string) (string, error)

// ExpandPath converts ~ notation to absolute path
// Example: ~/.zshrc -> /Users/you/.zshrc
func ExpandPath(path string) (string, error)

// GetRepoFilePath returns full path to file in repo
// Example: zsh/zshrc -> ~/.dotman/files/zsh/zshrc
func GetRepoFilePath(repoPath string) (string, error)

// GenerateRepoPath creates repo path from source path
// Example: ~/.config/nvim/init.vim -> nvim/init.vim
func GenerateRepoPath(sourcePath string) string

// GetBackupFileName generates timestamped backup filename
// Example: ~/.zshrc -> 2025-01-04_103000_zshrc
func GetBackupFileName(sourcePath string) string
```

**Implementation notes:**
- Use `os.UserHomeDir()` for home directory
- Handle edge cases: no home dir, already absolute, etc.
- Repo path generation: strip leading dot, flatten structure intelligently

---

#### 1.3 File Operations (`internal/fs/fs.go`)

**Required functions:**

```go
// CopyFile copies file with permissions preserved
func CopyFile(src, dst string) error

// BackupFile creates timestamped backup of file
func BackupFile(sourcePath, backupDir string) (string, error)

// FileExists checks if file exists
func FileExists(path string) bool

// FilesIdentical compares two files byte-by-byte
func FilesIdentical(path1, path2 string) (bool, error)

// EnsureDir creates directory if it doesn't exist
func EnsureDir(path string) error

// GetFileDiff returns unified diff between two files
func GetFileDiff(path1, path2, label1, label2 string) (string, error)
```

**Implementation notes:**
- Preserve file permissions when copying
- Create parent directories as needed
- Use `io.Copy` for efficient file copying
- For diff: consider using `os/exec` with system `diff` or a Go diff library

---

#### 1.4 File Validation (`internal/fs/validate.go`)

**Required functions:**

```go
// ValidateSourceFile checks if source file is valid for tracking
func ValidateSourceFile(path string) error

// ValidateRepoPath checks if repo path is valid
func ValidateRepoPath(path string) error
```

**Validation rules:**
- File must exist
- File must be readable
- File must not be a directory (for MVP)
- Path must be absolute or start with ~

---

#### 1.5 Git Wrapper (`internal/git/git.go`)

**Required functions:**

```go
// InitRepo initializes git repository in directory
func InitRepo(repoPath string) error

// Commit stages all changes and commits with message
func Commit(repoPath, message string) error

// Push pushes to remote (if configured)
func Push(repoPath string) error

// Pull pulls from remote (if configured)
func Pull(repoPath string) error

// HasChanges checks if working tree has uncommitted changes
func HasChanges(repoPath string) (bool, error)

// SetRemote configures git remote
func SetRemote(repoPath, remoteName, remoteURL string) error
```

**Implementation notes:**
- Use `os/exec.Command("git", ...)`
- Run git commands in repo directory
- Capture stderr for error messages
- Check if git is installed

---

### Phase 2: Commands (Implement in Order)

#### 2.1 `dotman init` (`cmd/dotman/init.go`)

**What it does:**
1. Check if `~/.dotman` already exists (prevent re-init)
2. Create directory structure:
   - `~/.dotman/`
   - `~/.dotman/files/`
   - `~/.dotman/backups/`
3. Initialize Git repository in `~/.dotman/files/`
4. Create default `config.yaml`
5. Success message

**Command definition:**

```go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize DotMan repository",
    Long:  `Creates ~/.dotman directory structure and initializes Git repository.`,
    Run:   runInit,
}

func runInit(cmd *cobra.Command, args []string) {
    // Implementation
}
```

**Output example:**

```
✓ Created ~/.dotman/
✓ Created ~/.dotman/files/
✓ Created ~/.dotman/backups/
✓ Initialized Git repository
✓ Created config.yaml

DotMan is ready! Next steps:
  dotman track ~/.zshrc
  dotman list
```

**Error handling:**
- Already initialized: friendly message, exit
- Permission denied: clear error message
- Git not installed: warning (can work without Git)

---

#### 2.2 `dotman track <file>` (`cmd/dotman/track.go`)

**What it does:**
1. Validate file exists and is readable
2. Normalize source path (convert to ~ notation)
3. Check if already tracked
4. Generate repo path (e.g., `zsh/zshrc`)
5. Copy file to `~/.dotman/files/{repo_path}`
6. Add to config.yaml
7. Git commit (if enabled): "Track {source_path}"
8. Success message

**Command definition:**

```go
var trackCmd = &cobra.Command{
    Use:   "track [file]",
    Short: "Track a dotfile",
    Args:  cobra.ExactArgs(1),
    Run:   runTrack,
}
```

**Output example:**

```
✓ Tracking ~/.zshrc
  Stored as: zsh/zshrc
  Location: ~/.dotman/files/zsh/zshrc
✓ Committed to Git
```

**Error handling:**
- File doesn't exist
- File not readable
- Already tracked (offer to update)
- Git commit fails (warn but don't fail)

---

#### 2.3 `dotman list` (`cmd/dotman/list.go`)

**What it does:**
1. Load config
2. Display tracked files in table format
3. Show count

**Command definition:**

```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all tracked dotfiles",
    Aliases: []string{"ls"},
    Run:   runList,
}
```

**Output example:**

```
Tracked dotfiles (3):

SOURCE PATH                     REPO PATH              TRACKED AT
~/.zshrc                        zsh/zshrc              2025-01-04 10:30
~/.config/nvim/init.vim         nvim/init.vim          2025-01-04 10:31
~/.gitconfig                    git/gitconfig          2025-01-04 10:32
```

**Error handling:**
- No tracked files: friendly message
- Config not found: suggest running `dotman init`

---

#### 2.4 `dotman status` (`cmd/dotman/status.go`)

**What it does:**
1. Load config
2. For each tracked file:
   - Compare repo version vs system version
   - Detect: identical, modified, missing
3. Display status for each file

**Command definition:**

```go
var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show status of tracked dotfiles",
    Run:   runStatus,
}
```

**Output example:**

```
Status:

✓ ~/.zshrc                      Up to date
✗ ~/.config/nvim/init.vim       Modified (system differs from repo)
! ~/.gitconfig                  Missing from system
```

**Legend:**
- ✓ = up to date
- ✗ = modified
- ! = missing

---

#### 2.5 `dotman apply` (`cmd/dotman/apply.go`)

**What it does:**
1. Load config
2. For each tracked file:
   - If system file doesn't exist: copy from repo
   - If system file identical: skip
   - If system file differs:
     - Show diff
     - Prompt: "Overwrite? [y/N/d=diff]"
     - If yes: backup system file, copy from repo
     - If no: skip
3. Git commit (if any changes): "Apply dotfiles"
4. Summary

**Command definition:**

```go
var applyCmd = &cobra.Command{
    Use:   "apply",
    Short: "Apply dotfiles from repository to system",
    Run:   runApply,
}
```

**Output example:**

```
Applying dotfiles...

✓ ~/.zshrc                      Already up to date
? ~/.config/nvim/init.vim       System file differs

--- System
+++ Repository
@@ -1,3 +1,4 @@
 set number
+set relativenumber
 syntax on

Overwrite ~/.config/nvim/init.vim? [y/N/d]: y
✓ Backed up to: ~/.dotman/backups/2025-01-04_103500_init.vim
✓ Applied from repository

! ~/.gitconfig                  Missing from system
✓ Copied from repository

Summary: 2 applied, 1 skipped, 0 errors
```

**Flags:**
- `--force` - Apply all without prompting
- `--dry-run` - Show what would be applied

---

#### 2.6 `dotman untrack <file>` (`cmd/dotman/untrack.go`)

**What it does:**
1. Validate file is tracked
2. Remove from config.yaml
3. Prompt: "Delete from repository? [y/N]"
4. If yes: delete from `~/.dotman/files/`
5. Git commit: "Untrack {source_path}"

**Command definition:**

```go
var untrackCmd = &cobra.Command{
    Use:   "untrack [file]",
    Short: "Stop tracking a dotfile",
    Args:  cobra.ExactArgs(1),
    Run:   runUntrack,
}
```

**Output example:**

```
✓ Untracked ~/.zshrc
? Delete from repository? [y/N]: y
✓ Deleted from repository
✓ Committed to Git
```

---

#### 2.7 `dotman push` (`cmd/dotman/push.go`)

**What it does:**
1. Check if Git remote is configured
2. Check for uncommitted changes, auto-commit if needed
3. Git push to remote
4. Success message

**Command definition:**

```go
var pushCmd = &cobra.Command{
    Use:   "push",
    Short: "Push dotfiles to Git remote",
    Run:   runPush,
}
```

**Output example:**

```
✓ Committed local changes
✓ Pushed to origin
```

**Error handling:**
- No remote configured: show instructions
- Push fails: show Git error

---

#### 2.8 `dotman pull` (`cmd/dotman/pull.go`)

**What it does:**
1. Check if Git remote is configured
2. Git pull from remote
3. Prompt: "Apply pulled changes? [y/N]"
4. If yes: run `dotman apply`

**Command definition:**

```go
var pullCmd = &cobra.Command{
    Use:   "pull",
    Short: "Pull dotfiles from Git remote",
    Run:   runPull,
}
```

**Output example:**

```
✓ Pulled from origin
? Apply changes to system? [y/N]: y

[... runs dotman apply ...]
```

---

### Phase 3: Testing & Polish

#### 3.1 Manual Testing Flow

```bash
# Initialize
dotman init

# Track some files
dotman track ~/.zshrc
dotman track ~/.gitconfig
dotman list

# Modify a system file
echo "# test" >> ~/.zshrc
dotman status

# Apply from repo (should show diff and prompt)
dotman apply

# Test on "new machine" (simulate by moving files)
mv ~/.zshrc ~/.zshrc.bak
dotman apply
```

#### 3.2 Edge Cases to Test

- Track non-existent file (should error)
- Track already-tracked file (should prompt to update)
- Apply when system file is newer (should prompt)
- Apply when no tracked files (friendly message)
- Untrack non-tracked file (should error)
- Init when already initialized (should warn)
- Operations without Git installed (should work, skip Git)

#### 3.3 Polish Items

- Consistent error messages
- Help text for all commands
- Color output (consider using a color library)
- Progress indicators for long operations
- Validate Git is installed (warn if not)

---

## Development Order Checklist

**Infrastructure (build first):**
- [ ] `internal/config/config.go` - Config structs and Viper integration
- [ ] `internal/config/paths.go` - Path normalization
- [ ] `internal/fs/fs.go` - File operations
- [ ] `internal/fs/validate.go` - Validation
- [ ] `internal/git/git.go` - Git wrapper

**Commands (build in order):**
- [ ] `cmd/dotman/main.go` - Cobra setup
- [ ] `cmd/dotman/init.go` - Initialize DotMan
- [ ] `cmd/dotman/track.go` - Track files
- [ ] `cmd/dotman/list.go` - List tracked files
- [ ] `cmd/dotman/status.go` - Show status
- [ ] `cmd/dotman/apply.go` - Apply dotfiles
- [ ] `cmd/dotman/untrack.go` - Untrack files
- [ ] `cmd/dotman/push.go` - Git push
- [ ] `cmd/dotman/pull.go` - Git pull

**Testing:**
- [ ] Manual end-to-end test flow
- [ ] Edge case testing
- [ ] Polish and error messages

---

## Future Enhancements (Post-MVP)

- Unit tests for core packages
- Integration tests
- CI/CD with GitHub Actions
- Homebrew formula
- Binary releases with `goreleaser`
- Symlink mode (alternative to copy)
- Profile support (work, home, server)
- Template variables
- Encrypted secrets support

---

## Resources

- [Cobra Documentation](https://github.com/spf13/cobra)
- [Viper Documentation](https://github.com/spf13/viper)
- [Go YAML v3](https://github.com/go-yaml/yaml)
