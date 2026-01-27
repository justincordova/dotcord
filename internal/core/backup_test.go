package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateBackup(t *testing.T) {
	// Create temp dir structure
	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a source file
	sourceContent := []byte("original content")
	sourceFile := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(sourceFile, sourceContent, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Create backup
	backupPath, err := CreateBackup(sourceFile)
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("CreateBackup() backup file not created")
	}

	// Verify backup content matches
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup file: %v", err)
	}
	if string(backupContent) != string(sourceContent) {
		t.Errorf("CreateBackup() content mismatch: got %q, want %q", backupContent, sourceContent)
	}
}

func TestCreateBackupNonexistent(t *testing.T) {
	_, err := CreateBackup("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("CreateBackup() should error for nonexistent file")
	}
}

func TestRestoreBackup(t *testing.T) {
	// Create temp dir
	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a backup file
	backupContent := []byte("backup content")
	backupFile := filepath.Join(tempDir, "backup.txt")
	if err := os.WriteFile(backupFile, backupContent, 0644); err != nil {
		t.Fatalf("failed to create backup file: %v", err)
	}

	// Restore to target
	targetFile := filepath.Join(tempDir, "restored.txt")
	if err := RestoreBackup(backupFile, targetFile); err != nil {
		t.Fatalf("RestoreBackup() error = %v", err)
	}

	// Verify target exists with correct content
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}
	if string(targetContent) != string(backupContent) {
		t.Errorf("RestoreBackup() content mismatch: got %q, want %q", targetContent, backupContent)
	}
}

func TestRestoreBackupCreatesParentDir(t *testing.T) {
	// Create temp dir
	tempDir, err := os.MkdirTemp("", "dotcor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a backup file
	backupFile := filepath.Join(tempDir, "backup.txt")
	if err := os.WriteFile(backupFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create backup file: %v", err)
	}

	// Restore to nested path (parent doesn't exist)
	targetFile := filepath.Join(tempDir, "nested", "dir", "restored.txt")
	if err := RestoreBackup(backupFile, targetFile); err != nil {
		t.Fatalf("RestoreBackup() error = %v", err)
	}

	// Verify target exists
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Error("RestoreBackup() target file not created")
	}
}

func TestBackupExists(t *testing.T) {
	// Without any backups created, BackupExists should return false
	// for a random filename
	exists := BackupExists("random_nonexistent_file_12345.txt")
	if exists {
		t.Error("BackupExists() should return false for nonexistent backups")
	}
}

func TestGetBackupCount(t *testing.T) {
	// Get current count (may be 0 or have existing backups)
	count, err := GetBackupCount()
	if err != nil {
		t.Fatalf("GetBackupCount() error = %v", err)
	}

	// Should be non-negative
	if count < 0 {
		t.Errorf("GetBackupCount() = %d, should be >= 0", count)
	}
}

func TestGetTotalBackupSize(t *testing.T) {
	size, err := GetTotalBackupSize()
	if err != nil {
		t.Fatalf("GetTotalBackupSize() error = %v", err)
	}

	// Should be non-negative
	if size < 0 {
		t.Errorf("GetTotalBackupSize() = %d, should be >= 0", size)
	}
}

func TestTimestampFormat(t *testing.T) {
	// Test that TimestampFormat produces parseable timestamps
	now := time.Now()
	formatted := now.Format(TimestampFormat)

	parsed, err := time.Parse(TimestampFormat, formatted)
	if err != nil {
		t.Errorf("TimestampFormat not parseable: %v", err)
	}

	// Verify year, month, day, hour, minute, second match
	if parsed.Year() != now.Year() ||
		parsed.Month() != now.Month() ||
		parsed.Day() != now.Day() ||
		parsed.Hour() != now.Hour() ||
		parsed.Minute() != now.Minute() ||
		parsed.Second() != now.Second() {
		t.Error("TimestampFormat lost precision")
	}
}

func TestPreviewCleanup(t *testing.T) {
	// Preview cleanup with default settings shouldn't error
	_, _, err := PreviewCleanup(30*24*time.Hour, 5)
	if err != nil {
		t.Errorf("PreviewCleanup() error = %v", err)
	}
}

func TestCleanupCandidate(t *testing.T) {
	// Test that CleanupCandidate struct works correctly
	candidate := CleanupCandidate{
		Path:      "/some/path",
		Timestamp: time.Now(),
		Size:      1024,
	}

	if candidate.Path != "/some/path" {
		t.Error("CleanupCandidate.Path not set correctly")
	}
	if candidate.Size != 1024 {
		t.Error("CleanupCandidate.Size not set correctly")
	}
}
