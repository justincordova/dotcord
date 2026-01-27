package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestIsGitInstalled(t *testing.T) {
	// Git should be installed on dev machines
	if !IsGitInstalled() {
		t.Skip("git not installed, skipping tests")
	}
}

func TestInitRepo(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	// Verify .git directory exists
	gitDir := filepath.Join(tempDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error("InitRepo() did not create .git directory")
	}
}

func TestIsRepo(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Should not be a repo initially
	if IsRepo(tempDir) {
		t.Error("IsRepo() should return false for non-repo directory")
	}

	// Initialize and check again
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	if !IsRepo(tempDir) {
		t.Error("IsRepo() should return true after InitRepo()")
	}
}

func TestHasChanges(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	// No changes initially
	hasChanges, err := HasChanges(tempDir)
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if hasChanges {
		t.Error("HasChanges() should return false for clean repo")
	}

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should have changes now
	hasChanges, err = HasChanges(tempDir)
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if !hasChanges {
		t.Error("HasChanges() should return true after adding file")
	}
}

func TestAutoCommit(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	// Configure git user (required for commits)
	configureGitUser(t, tempDir)

	// AutoCommit with no changes should succeed silently
	if err := AutoCommit(tempDir, "test commit"); err != nil {
		t.Fatalf("AutoCommit() with no changes error = %v", err)
	}

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// AutoCommit should commit the file
	if err := AutoCommit(tempDir, "add test file"); err != nil {
		t.Fatalf("AutoCommit() error = %v", err)
	}

	// Verify no more changes
	hasChanges, err := HasChanges(tempDir)
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if hasChanges {
		t.Error("AutoCommit() should have committed all changes")
	}
}

func TestGetStatus(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	configureGitUser(t, tempDir)

	status, err := GetStatus(tempDir)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	// No remote configured
	if status.RemoteExists {
		t.Error("GetStatus().RemoteExists should be false without remote")
	}

	// No changes
	if status.HasUncommitted {
		t.Error("GetStatus().HasUncommitted should be false for clean repo")
	}

	// Add uncommitted file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	status, err = GetStatus(tempDir)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if !status.HasUncommitted {
		t.Error("GetStatus().HasUncommitted should be true with changes")
	}
}

func TestGetRemoteURL(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	// No remote initially
	url, err := GetRemoteURL(tempDir)
	if err != nil {
		t.Fatalf("GetRemoteURL() error = %v", err)
	}
	if url != "" {
		t.Errorf("GetRemoteURL() = %q, want empty string", url)
	}
}

func TestSetRemote(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	// Set remote
	testURL := "https://github.com/test/repo.git"
	if err := SetRemote(tempDir, "origin", testURL); err != nil {
		t.Fatalf("SetRemote() error = %v", err)
	}

	// Verify remote was set
	url, err := GetRemoteURL(tempDir)
	if err != nil {
		t.Fatalf("GetRemoteURL() error = %v", err)
	}
	if url != testURL {
		t.Errorf("GetRemoteURL() = %q, want %q", url, testURL)
	}

	// Update remote
	newURL := "https://github.com/test/new-repo.git"
	if err := SetRemote(tempDir, "origin", newURL); err != nil {
		t.Fatalf("SetRemote() update error = %v", err)
	}

	url, err = GetRemoteURL(tempDir)
	if err != nil {
		t.Fatalf("GetRemoteURL() error = %v", err)
	}
	if url != newURL {
		t.Errorf("GetRemoteURL() after update = %q, want %q", url, newURL)
	}
}

func TestGetFileHistory(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	configureGitUser(t, tempDir)

	// Create and commit a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("v1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := AutoCommit(tempDir, "initial commit"); err != nil {
		t.Fatalf("AutoCommit() error = %v", err)
	}

	// Update and commit again
	if err := os.WriteFile(testFile, []byte("v2"), 0644); err != nil {
		t.Fatalf("failed to update test file: %v", err)
	}

	if err := AutoCommit(tempDir, "second commit"); err != nil {
		t.Fatalf("AutoCommit() error = %v", err)
	}

	// Get history
	history, err := GetFileHistory(tempDir, "test.txt", 10)
	if err != nil {
		t.Fatalf("GetFileHistory() error = %v", err)
	}

	if len(history) != 2 {
		t.Errorf("GetFileHistory() returned %d commits, want 2", len(history))
	}

	// Most recent first
	if len(history) > 0 && history[0].Message != "second commit" {
		t.Errorf("GetFileHistory()[0].Message = %q, want %q", history[0].Message, "second commit")
	}
}

func TestGetCurrentCommit(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	configureGitUser(t, tempDir)

	// Create and commit a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := AutoCommit(tempDir, "test commit"); err != nil {
		t.Fatalf("AutoCommit() error = %v", err)
	}

	// Get current commit
	commit, err := GetCurrentCommit(tempDir)
	if err != nil {
		t.Fatalf("GetCurrentCommit() error = %v", err)
	}

	// Should be a 40-character hex string
	if len(commit) != 40 {
		t.Errorf("GetCurrentCommit() returned %q (len=%d), want 40 chars", commit, len(commit))
	}
}

func TestGetChangedFiles(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	// No changes initially
	files, err := GetChangedFiles(tempDir)
	if err != nil {
		t.Fatalf("GetChangedFiles() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("GetChangedFiles() returned %d files, want 0", len(files))
	}

	// Create files
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		testFile := filepath.Join(tempDir, name)
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Should have 3 changed files
	files, err = GetChangedFiles(tempDir)
	if err != nil {
		t.Fatalf("GetChangedFiles() error = %v", err)
	}
	if len(files) != 3 {
		t.Errorf("GetChangedFiles() returned %d files, want 3", len(files))
	}
}

func TestGetDiff(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	configureGitUser(t, tempDir)

	// Create and commit a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := AutoCommit(tempDir, "initial commit"); err != nil {
		t.Fatalf("AutoCommit() error = %v", err)
	}

	// No diff after commit
	diff, err := GetDiff(tempDir)
	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}
	if diff != "" {
		t.Error("GetDiff() should return empty string for clean repo")
	}

	// Modify file
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Should have diff now
	diff, err = GetDiff(tempDir)
	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}
	if diff == "" {
		t.Error("GetDiff() should return diff for modified file")
	}
}

func TestStageAndUnstageFile(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}

	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repo
	if err := InitRepo(tempDir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	configureGitUser(t, tempDir)

	// Create and commit initial file to have HEAD
	initialFile := filepath.Join(tempDir, "initial.txt")
	if err := os.WriteFile(initialFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("failed to create initial file: %v", err)
	}
	if err := AutoCommit(tempDir, "initial commit"); err != nil {
		t.Fatalf("AutoCommit() error = %v", err)
	}

	// Create a new file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Stage the file
	if err := StageFile(tempDir, "test.txt"); err != nil {
		t.Fatalf("StageFile() error = %v", err)
	}

	// Unstage the file
	if err := UnstageFile(tempDir, "test.txt"); err != nil {
		t.Fatalf("UnstageFile() error = %v", err)
	}
}

func TestStatusInfo(t *testing.T) {
	// Test StatusInfo struct fields
	info := StatusInfo{
		HasUncommitted: true,
		AheadBy:        2,
		BehindBy:       1,
		Branch:         "main",
		RemoteExists:   true,
	}

	if !info.HasUncommitted {
		t.Error("StatusInfo.HasUncommitted not set")
	}
	if info.AheadBy != 2 {
		t.Errorf("StatusInfo.AheadBy = %d, want 2", info.AheadBy)
	}
	if info.BehindBy != 1 {
		t.Errorf("StatusInfo.BehindBy = %d, want 1", info.BehindBy)
	}
	if info.Branch != "main" {
		t.Errorf("StatusInfo.Branch = %q, want %q", info.Branch, "main")
	}
	if !info.RemoteExists {
		t.Error("StatusInfo.RemoteExists not set")
	}
}

func TestCommitInfo(t *testing.T) {
	// Test CommitInfo struct fields
	now := time.Now()
	info := CommitInfo{
		Hash:    "abc123",
		Author:  "Test User",
		Date:    now,
		Message: "test commit",
	}

	if info.Hash != "abc123" {
		t.Errorf("CommitInfo.Hash = %q, want %q", info.Hash, "abc123")
	}
	if info.Author != "Test User" {
		t.Errorf("CommitInfo.Author = %q, want %q", info.Author, "Test User")
	}
	if info.Message != "test commit" {
		t.Errorf("CommitInfo.Message = %q, want %q", info.Message, "test commit")
	}
}

// Helper function to configure git user in test repos
func configureGitUser(t *testing.T, repoPath string) {
	t.Helper()

	// Set user.email
	cmd := exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to configure git user.email: %v", err)
	}

	// Set user.name
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to configure git user.name: %v", err)
	}
}
