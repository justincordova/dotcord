# DotMan

A simple, safe, and powerful dotfile manager built in Go.

DotMan helps you track, backup, and sync your dotfiles across machines with Git integration and a focus on safety.

---

## Features

- **Simple CLI** - Easy-to-use commands for everyday dotfile management
- **Git Integration** - Automatic commits and optional remote sync
- **Safety First** - Always backs up before overwriting files
- **Cross-Platform** - Works on macOS, Linux, and Windows
- **Portable** - Path normalization works across different users and machines
- **No Surprises** - Shows diffs and prompts before making changes

---

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/justincordova/dotman.git
cd dotman

# Build and install
go build -o dotman cmd/dotman/main.go
sudo mv dotman /usr/local/bin/

# Or just run directly
go run cmd/dotman/main.go
```

### Basic Usage

```bash
# Initialize DotMan
dotman init

# Track your dotfiles
dotman track ~/.zshrc
dotman track ~/.gitconfig
dotman track ~/.config/nvim/init.vim

# List tracked files
dotman list

# Check status (what's different between repo and system)
dotman status

# Apply dotfiles from repository to system
dotman apply

# Push to remote Git repository
dotman push

# Pull from remote and apply
dotman pull
```

---

## How It Works

### Storage Model

DotMan uses a **copy-based** approach:

1. **Track** - Copies your dotfile to `~/.dotman/files/`
2. **Repository** - Stores dotfiles in a Git repository
3. **Apply** - Copies dotfiles from repository back to system locations

```
~/.zshrc  ──track──>  ~/.dotman/files/zsh/zshrc  ──apply──>  ~/.zshrc
                             (Git repo)
```

### Directory Structure

```
~/.dotman/
├── config.yaml          # Metadata: which files are tracked
├── files/               # Git repository storing your dotfiles
│   ├── .git/
│   ├── zsh/
│   │   └── zshrc
│   └── nvim/
│       └── init.vim
└── backups/             # Timestamped backups before overwrites
    └── 2025-01-04_103000_zshrc
```

### Safety Features

- **Always backs up** before overwriting system files
- **Shows diffs** before applying changes
- **Requires confirmation** for destructive operations
- **Never silently modifies** your files
- **Timestamped backups** stored in `~/.dotman/backups/`

---

## Commands

### `dotman init`

Initialize DotMan repository.

```bash
dotman init
```

Creates:
- `~/.dotman/` directory structure
- Git repository in `~/.dotman/files/`
- Default configuration file

---

### `dotman track <file>`

Track a dotfile.

```bash
dotman track ~/.zshrc
dotman track ~/.config/nvim/init.vim
```

What it does:
1. Validates file exists
2. Copies to `~/.dotman/files/`
3. Records in `config.yaml`
4. Git commits automatically

---

### `dotman list`

List all tracked dotfiles.

```bash
dotman list
```

Output:
```
Tracked dotfiles (3):

SOURCE PATH                     REPO PATH              TRACKED AT
~/.zshrc                        zsh/zshrc              2025-01-04 10:30
~/.config/nvim/init.vim         nvim/init.vim          2025-01-04 10:31
~/.gitconfig                    git/gitconfig          2025-01-04 10:32
```

---

### `dotman status`

Show status of tracked dotfiles.

```bash
dotman status
```

Compares repository version with system version:

```
Status:

✓ ~/.zshrc                      Up to date
✗ ~/.config/nvim/init.vim       Modified (system differs)
! ~/.gitconfig                  Missing from system
```

---

### `dotman apply`

Apply dotfiles from repository to system.

```bash
dotman apply
```

What it does:
1. Compares each tracked file
2. Shows diff if different
3. Prompts for confirmation
4. Backs up existing file
5. Copies from repository

**Flags:**
- `--force` - Apply without prompting
- `--dry-run` - Show what would be applied

Example output:
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

Summary: 1 applied, 1 skipped
```

---

### `dotman untrack <file>`

Stop tracking a dotfile.

```bash
dotman untrack ~/.zshrc
```

What it does:
1. Removes from `config.yaml`
2. Prompts to delete from repository
3. Git commits automatically

---

### `dotman push`

Push dotfiles to Git remote.

```bash
dotman push
```

Requires Git remote to be configured:
```bash
cd ~/.dotman/files
git remote add origin git@github.com:yourusername/dotfiles.git
```

---

### `dotman pull`

Pull dotfiles from Git remote.

```bash
dotman pull
```

What it does:
1. Pulls from remote repository
2. Prompts to apply changes
3. Runs `dotman apply` if confirmed

---

## Configuration

Configuration is stored in `~/.dotman/config.yaml`:

```yaml
repo_path: /Users/you/.dotman
backup_path: /Users/you/.dotman/backups
git_enabled: true
git_remote: ""

tracked_files:
  - source_path: ~/.zshrc
    repo_path: zsh/zshrc
    tracked_at: 2025-01-04T10:30:00Z
```

You can manually edit this file if needed.

---

## Use Cases

### New Machine Setup

On your main machine:
```bash
dotman init
dotman track ~/.zshrc
dotman track ~/.gitconfig
cd ~/.dotman/files
git remote add origin git@github.com:you/dotfiles.git
dotman push
```

On a new machine:
```bash
# Install DotMan
dotman init
cd ~/.dotman/files
git remote add origin git@github.com:you/dotfiles.git
dotman pull
# Prompts to apply → your dotfiles are restored!
```

---

### Daily Workflow

Edit your dotfiles as usual. When you want to update your repository:

```bash
dotman status          # See what changed
dotman track ~/.zshrc  # Update tracked file in repo
dotman push            # Sync to remote
```

Apply changes on another machine:

```bash
dotman pull            # Pull and apply
```

---

### Recover from Mistakes

DotMan backs up every file before overwriting:

```bash
# Oops, applied the wrong version
ls ~/.dotman/backups/
# 2025-01-04_103500_zshrc

# Restore from backup
cp ~/.dotman/backups/2025-01-04_103500_zshrc ~/.zshrc
```

---

## Why DotMan?

### vs. Manual Git Repository

- **DotMan:** Automatic backups, conflict detection, path normalization
- **Manual:** You handle everything yourself

### vs. GNU Stow

- **DotMan:** Copy-based, Git integration, works everywhere
- **Stow:** Symlink-based, requires specific directory structure

### vs. Chezmoi

- **DotMan:** Simple, minimal, focused on core features
- **Chezmoi:** Feature-rich, templates, more complexity

DotMan is for users who want:
- Simple dotfile management without learning a complex tool
- Git integration without manual commits
- Safety features (backups, diffs, prompts)
- A tool that stays out of your way

---

## Development

### Project Structure

```
dotman/
├── cmd/dotman/          # CLI commands (Cobra)
├── internal/
│   ├── config/          # Configuration management
│   ├── core/            # Business logic
│   ├── fs/              # File operations
│   └── git/             # Git integration
├── PLAN.md              # Implementation plan
└── README.md            # This file
```

### Building

```bash
go build -o dotman cmd/dotman/main.go
```

### Running

```bash
go run cmd/dotman/main.go [command]
```

### Contributing

See `PLAN.md` for implementation details and development roadmap.

---

## Roadmap

### v1.0 (Current - MVP)
- Core CLI commands
- Git integration
- Safety features (backups, diffs, prompts)
- Path normalization

### v2.0 (Future)
- Profile system (work, home, server)
- OS-specific overrides
- Template variables
- Local HTTP API

### v3.0 (Future)
- Secrets manager with encryption
- System package export (Homebrew, npm, etc.)
- Advanced templating

### v4.0 (Future)
- Desktop GUI (Flutter)
- Plugin system
- Background sync daemon

---

## License

MIT License - see LICENSE file for details

---

## Author

Built by Justin Cordova

---

## Support

- **Issues:** https://github.com/justincordova/dotman/issues
- **Discussions:** https://github.com/justincordova/dotman/discussions
