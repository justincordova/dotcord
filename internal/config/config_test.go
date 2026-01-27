package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetDefaultIgnorePatterns(t *testing.T) {
	patterns := GetDefaultIgnorePatterns()

	if len(patterns) == 0 {
		t.Error("GetDefaultIgnorePatterns() returned empty slice")
	}

	// Check for expected patterns
	expected := []string{"*.key", ".env", "id_rsa", "*.swp", ".DS_Store"}
	for _, exp := range expected {
		found := false
		for _, p := range patterns {
			if p == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetDefaultIgnorePatterns() missing expected pattern: %s", exp)
		}
	}
}

func TestNewDefaultConfig(t *testing.T) {
	cfg, err := NewDefaultConfig()
	if err != nil {
		t.Fatalf("NewDefaultConfig() error = %v", err)
	}

	if cfg.Version != CurrentConfigVersion {
		t.Errorf("Version = %v, want %v", cfg.Version, CurrentConfigVersion)
	}

	if !cfg.GitEnabled {
		t.Error("GitEnabled should be true by default")
	}

	if len(cfg.IgnorePatterns) == 0 {
		t.Error("IgnorePatterns should not be empty")
	}

	if len(cfg.ManagedFiles) != 0 {
		t.Error("ManagedFiles should be empty initially")
	}
}

func TestShouldApplyOnPlatform(t *testing.T) {
	tests := []struct {
		name            string
		platforms       []string
		currentPlatform string
		want            bool
	}{
		{
			name:            "empty platforms means all",
			platforms:       []string{},
			currentPlatform: "darwin",
			want:            true,
		},
		{
			name:            "nil platforms means all",
			platforms:       nil,
			currentPlatform: "linux",
			want:            true,
		},
		{
			name:            "matching platform",
			platforms:       []string{"darwin", "linux"},
			currentPlatform: "darwin",
			want:            true,
		},
		{
			name:            "non-matching platform",
			platforms:       []string{"darwin"},
			currentPlatform: "linux",
			want:            false,
		},
		{
			name:            "wsl platform",
			platforms:       []string{"wsl", "linux"},
			currentPlatform: "wsl",
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldApplyOnPlatform(tt.platforms, tt.currentPlatform)
			if got != tt.want {
				t.Errorf("ShouldApplyOnPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigManagedFiles(t *testing.T) {
	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config
	cfg := &Config{
		Version:        CurrentConfigVersion,
		RepoPath:       filepath.Join(tempDir, "files"),
		GitEnabled:     false,
		IgnorePatterns: []string{},
		ManagedFiles:   []ManagedFile{},
	}

	// Test IsManaged on empty config
	if cfg.IsManaged("~/.zshrc") {
		t.Error("IsManaged() should return false for unmanaged file")
	}

	// Add a managed file manually (without saving)
	mf := ManagedFile{
		SourcePath: "~/.zshrc",
		RepoPath:   "shell/zshrc",
		AddedAt:    time.Now(),
		Platforms:  []string{},
	}
	cfg.ManagedFiles = append(cfg.ManagedFiles, mf)

	// Test IsManaged
	if !cfg.IsManaged("~/.zshrc") {
		t.Error("IsManaged() should return true for managed file")
	}

	// Test GetManagedFile
	got, err := cfg.GetManagedFile("~/.zshrc")
	if err != nil {
		t.Errorf("GetManagedFile() error = %v", err)
	}
	if got.RepoPath != "shell/zshrc" {
		t.Errorf("GetManagedFile().RepoPath = %v, want shell/zshrc", got.RepoPath)
	}

	// Test GetManagedFile for non-existent file
	_, err = cfg.GetManagedFile("~/.nonexistent")
	if err == nil {
		t.Error("GetManagedFile() should return error for non-existent file")
	}
}

func TestGetManagedFilesForPlatform(t *testing.T) {
	cfg := &Config{
		Version:    CurrentConfigVersion,
		RepoPath:   "~/.dotcor/files",
		GitEnabled: false,
		ManagedFiles: []ManagedFile{
			{
				SourcePath: "~/.zshrc",
				RepoPath:   "shell/zshrc",
				Platforms:  []string{}, // All platforms
			},
			{
				SourcePath: "~/.bashrc",
				RepoPath:   "shell/bashrc",
				Platforms:  []string{"linux", "darwin"},
			},
			{
				SourcePath: "~/.wslconfig",
				RepoPath:   "wsl/wslconfig",
				Platforms:  []string{"wsl"},
			},
		},
	}

	// Get files for current platform
	files := cfg.GetManagedFilesForPlatform()

	// Should at least include the universal file
	found := false
	for _, f := range files {
		if f.SourcePath == "~/.zshrc" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetManagedFilesForPlatform() should include universal files")
	}
}

func TestGetUncommittedFiles(t *testing.T) {
	cfg := &Config{
		Version:    CurrentConfigVersion,
		RepoPath:   "~/.dotcor/files",
		GitEnabled: true,
		ManagedFiles: []ManagedFile{
			{
				SourcePath:     "~/.zshrc",
				RepoPath:       "shell/zshrc",
				HasUncommitted: false,
			},
			{
				SourcePath:     "~/.bashrc",
				RepoPath:       "shell/bashrc",
				HasUncommitted: true,
			},
			{
				SourcePath:     "~/.vimrc",
				RepoPath:       "vim/vimrc",
				HasUncommitted: true,
			},
		},
	}

	uncommitted := cfg.GetUncommittedFiles()

	if len(uncommitted) != 2 {
		t.Errorf("GetUncommittedFiles() returned %d files, want 2", len(uncommitted))
	}

	// Verify the correct files are returned
	for _, f := range uncommitted {
		if !f.HasUncommitted {
			t.Errorf("GetUncommittedFiles() returned file without uncommitted flag: %s", f.SourcePath)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "foo", false},
		{"Microsoft", "Microsoft", true},
		{"Linux version WSL", "WSL", true},
		{"", "foo", false},
		{"foo", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
