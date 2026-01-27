package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	patterns := []string{
		"*.key",
		".env",
		".env.*",
		"id_rsa",
		"*.swp",
		".DS_Store",
	}

	tests := []struct {
		name        string
		path        string
		wantIgnored bool
	}{
		{
			name:        "matches key pattern",
			path:        "/home/user/secret.key",
			wantIgnored: true,
		},
		{
			name:        "matches exact env",
			path:        "/home/user/.env",
			wantIgnored: true,
		},
		{
			name:        "matches env.local",
			path:        "/home/user/.env.local",
			wantIgnored: true,
		},
		{
			name:        "matches id_rsa",
			path:        "/home/user/.ssh/id_rsa",
			wantIgnored: true,
		},
		{
			name:        "matches swp",
			path:        "/home/user/.zshrc.swp",
			wantIgnored: true,
		},
		{
			name:        "matches DS_Store",
			path:        "/home/user/.DS_Store",
			wantIgnored: true,
		},
		{
			name:        "normal file not ignored",
			path:        "/home/user/.zshrc",
			wantIgnored: false,
		},
		{
			name:        "gitconfig not ignored",
			path:        "/home/user/.gitconfig",
			wantIgnored: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ShouldIgnore(tt.path, patterns)
			if got != tt.wantIgnored {
				t.Errorf("ShouldIgnore() = %v, want %v", got, tt.wantIgnored)
			}
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{
			name:    "glob asterisk match",
			path:    "/home/user/secret.key",
			pattern: "*.key",
			want:    true,
		},
		{
			name:    "exact match",
			path:    "/home/user/.env",
			pattern: ".env",
			want:    true,
		},
		{
			name:    "no match",
			path:    "/home/user/.zshrc",
			pattern: "*.env",
			want:    false,
		},
		{
			name:    "question mark match",
			path:    "/home/user/file1.txt",
			pattern: "file?.txt",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesPattern(tt.path, tt.pattern)
			if got != tt.want {
				t.Errorf("MatchesPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSecretFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"id_rsa", true},
		{"id_rsa.pub", true},
		{"id_ed25519", true},
		{".env", true},
		{".env.local", true},
		{"secret.key", true},
		{"credentials.pem", true},
		{".zshrc", false},
		{".gitconfig", false},
		{"config.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := IsSecretFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsSecretFile(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsHistoryFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{".bash_history", true},
		{".zsh_history", true},
		{".mysql_history", true},
		{".node_repl_history", true},
		{".lesshst", true},
		{".zshrc", false},
		{".bashrc", false},
		{"history.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := IsHistoryFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsHistoryFile(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsTemporaryFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{".zshrc.swp", true},
		{"file.swo", true},
		{"backup~", true},
		{"file.tmp", true},
		{"file.bak", true},
		{"file.orig", true},
		{".zshrc", false},
		{"config.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := IsTemporaryFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsTemporaryFile(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsSystemFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{".DS_Store", true},
		{"Thumbs.db", true},
		{"desktop.ini", true},
		{".zshrc", false},
		{"config.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := IsSystemFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsSystemFile(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestGetFileCategory(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"id_rsa", "secret"},
		{".env", "secret"},
		{".bash_history", "history"},
		{".zshrc.swp", "temporary"},
		{".DS_Store", "system"},
		{".zshrc", "normal"},
		{".gitconfig", "normal"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := GetFileCategory(tt.filename)
			if got != tt.want {
				t.Errorf("GetFileCategory(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestFilterByPatterns(t *testing.T) {
	paths := []string{
		"/home/user/.zshrc",
		"/home/user/.env",
		"/home/user/.gitconfig",
		"/home/user/secret.key",
		"/home/user/.DS_Store",
	}

	patterns := []string{".env", "*.key", ".DS_Store"}

	got := FilterByPatterns(paths, patterns)

	// Should filter out .env, secret.key, and .DS_Store
	expected := []string{
		"/home/user/.zshrc",
		"/home/user/.gitconfig",
	}

	if len(got) != len(expected) {
		t.Errorf("FilterByPatterns() returned %d paths, want %d", len(got), len(expected))
	}

	for i, path := range expected {
		if got[i] != path {
			t.Errorf("FilterByPatterns()[%d] = %v, want %v", i, got[i], path)
		}
	}
}

func TestMergePatterns(t *testing.T) {
	list1 := []string{"*.key", ".env", "*.swp"}
	list2 := []string{".env", ".DS_Store", "*.key"} // Has duplicates

	got := MergePatterns(list1, list2)

	// Should have 4 unique patterns
	if len(got) != 4 {
		t.Errorf("MergePatterns() returned %d patterns, want 4", len(got))
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, p := range got {
		if seen[p] {
			t.Errorf("MergePatterns() has duplicate: %s", p)
		}
		seen[p] = true
	}
}

func TestLoadGitignorePatterns(t *testing.T) {
	// Create temp dir
	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test gitignore file
	gitignoreContent := `# Comment line
*.swp
.env

# Another comment
*.key
id_rsa
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("failed to create gitignore file: %v", err)
	}

	patterns, err := LoadGitignorePatterns(gitignorePath)
	if err != nil {
		t.Fatalf("LoadGitignorePatterns() error = %v", err)
	}

	// Should have 4 patterns (comments and blank lines excluded)
	expected := []string{"*.swp", ".env", "*.key", "id_rsa"}

	if len(patterns) != len(expected) {
		t.Errorf("LoadGitignorePatterns() returned %d patterns, want %d", len(patterns), len(expected))
	}

	for i, pattern := range expected {
		if patterns[i] != pattern {
			t.Errorf("LoadGitignorePatterns()[%d] = %v, want %v", i, patterns[i], pattern)
		}
	}
}
