package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/justincordova/dotcor/internal/config"
	"github.com/justincordova/dotcor/internal/fs"
)

// BackupInfo represents information about a backup
type BackupInfo struct {
	Timestamp  time.Time
	SourcePath string // Original file path (normalized)
	BackupPath string // Full path to backup file
	Size       int64
}

// TimestampFormat is the format used for backup directory names
// Format: YYYY-MM-DD_HH-MM-SS (sortable, filesystem-safe)
const TimestampFormat = "2006-01-02_15-04-05"

// GetBackupDir returns the backup directory path (~/.dotcor/backups)
func GetBackupDir() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "backups"), nil
}

// CreateBackup creates a timestamped backup of a file before destructive operations
// Returns backup path and error
func CreateBackup(sourcePath string) (string, error) {
	// Expand source path
	expanded, err := config.ExpandPath(sourcePath)
	if err != nil {
		return "", fmt.Errorf("expanding source path: %w", err)
	}

	// Check if source exists
	if !fs.FileExists(expanded) {
		return "", fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	// Get backup directory
	backupDir, err := GetBackupDir()
	if err != nil {
		return "", err
	}

	// Create timestamped subdirectory
	timestamp := time.Now().Format(TimestampFormat)
	timestampDir := filepath.Join(backupDir, timestamp)

	if err := fs.EnsureDir(timestampDir); err != nil {
		return "", fmt.Errorf("creating backup directory: %w", err)
	}

	// Generate backup filename (strip leading dot, use original name)
	filename := filepath.Base(expanded)
	backupPath := filepath.Join(timestampDir, filename)

	// Handle name collisions by appending counter
	counter := 1
	for fs.FileExists(backupPath) {
		ext := filepath.Ext(filename)
		name := filename[:len(filename)-len(ext)]
		backupPath = filepath.Join(timestampDir, fmt.Sprintf("%s_%d%s", name, counter, ext))
		counter++
	}

	// Copy file to backup location
	if err := fs.CopyWithPermissions(expanded, backupPath); err != nil {
		return "", fmt.Errorf("copying to backup: %w", err)
	}

	return backupPath, nil
}

// RestoreBackup restores a file from backup to target path
func RestoreBackup(backupPath string, targetPath string) error {
	// Expand paths
	expandedBackup, err := config.ExpandPath(backupPath)
	if err != nil {
		return fmt.Errorf("expanding backup path: %w", err)
	}

	expandedTarget, err := config.ExpandPath(targetPath)
	if err != nil {
		return fmt.Errorf("expanding target path: %w", err)
	}

	// Check if backup exists
	if !fs.FileExists(expandedBackup) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Ensure target directory exists
	if err := fs.EnsureDir(filepath.Dir(expandedTarget)); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}

	// Copy backup to target
	if err := fs.CopyWithPermissions(expandedBackup, expandedTarget); err != nil {
		return fmt.Errorf("restoring from backup: %w", err)
	}

	return nil
}

// ListBackups returns list of all backups with timestamps
func ListBackups() ([]BackupInfo, error) {
	backupDir, err := GetBackupDir()
	if err != nil {
		return nil, err
	}

	// Check if backup directory exists
	if !fs.PathExists(backupDir) {
		return []BackupInfo{}, nil
	}

	var backups []BackupInfo

	// Walk through backup directory
	err = filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the timestamp directory name
		relPath, err := filepath.Rel(backupDir, path)
		if err != nil {
			return nil // Skip files we can't process
		}

		// Extract timestamp from directory name
		parts := filepath.SplitList(relPath)
		if len(parts) == 0 {
			parts = []string{filepath.Dir(relPath)}
		}

		timestampStr := filepath.Dir(relPath)
		if timestampStr == "." {
			return nil // Skip files directly in backup dir
		}

		// Parse timestamp
		timestamp, err := time.Parse(TimestampFormat, filepath.Base(timestampStr))
		if err != nil {
			// Try to get from directory name directly
			dirName := filepath.Base(filepath.Dir(path))
			timestamp, err = time.Parse(TimestampFormat, dirName)
			if err != nil {
				return nil // Skip if we can't parse timestamp
			}
		}

		backups = append(backups, BackupInfo{
			Timestamp:  timestamp,
			SourcePath: info.Name(), // Just filename, original path unknown
			BackupPath: path,
			Size:       info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking backup directory: %w", err)
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// CleanOldBackups removes backups older than specified duration, keeping at least keepLast
func CleanOldBackups(olderThan time.Duration, keepLast int) (int, int64, error) {
	backupDir, err := GetBackupDir()
	if err != nil {
		return 0, 0, err
	}

	// Check if backup directory exists
	if !fs.PathExists(backupDir) {
		return 0, 0, nil
	}

	// Get list of timestamp directories
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return 0, 0, fmt.Errorf("reading backup directory: %w", err)
	}

	// Parse and sort directories by timestamp
	type timestampDir struct {
		name      string
		timestamp time.Time
		path      string
	}

	var dirs []timestampDir
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		timestamp, err := time.Parse(TimestampFormat, entry.Name())
		if err != nil {
			continue // Skip directories that don't match timestamp format
		}

		dirs = append(dirs, timestampDir{
			name:      entry.Name(),
			timestamp: timestamp,
			path:      filepath.Join(backupDir, entry.Name()),
		})
	}

	// Sort by timestamp (newest first)
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].timestamp.After(dirs[j].timestamp)
	})

	// Determine which directories to delete
	cutoff := time.Now().Add(-olderThan)
	var toDelete []string
	var totalSize int64

	for i, dir := range dirs {
		// Keep at least keepLast backups
		if i < keepLast {
			continue
		}

		// Check if older than cutoff
		if dir.timestamp.Before(cutoff) {
			// Calculate size
			size, _ := getDirSize(dir.path)
			totalSize += size
			toDelete = append(toDelete, dir.path)
		}
	}

	// Delete old directories
	deleted := 0
	for _, path := range toDelete {
		if err := fs.RemoveAll(path); err != nil {
			// Continue deleting others even if one fails
			continue
		}
		deleted++
	}

	return deleted, totalSize, nil
}

// getDirSize calculates the total size of a directory
func getDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// GetBackupsForFile returns backups for a specific file (by filename)
func GetBackupsForFile(filename string) ([]BackupInfo, error) {
	allBackups, err := ListBackups()
	if err != nil {
		return nil, err
	}

	var fileBackups []BackupInfo
	for _, backup := range allBackups {
		if backup.SourcePath == filename || filepath.Base(backup.BackupPath) == filename {
			fileBackups = append(fileBackups, backup)
		}
	}

	return fileBackups, nil
}

// GetLatestBackup returns the most recent backup for a file
func GetLatestBackup(filename string) (*BackupInfo, error) {
	backups, err := GetBackupsForFile(filename)
	if err != nil {
		return nil, err
	}

	if len(backups) == 0 {
		return nil, fmt.Errorf("no backups found for: %s", filename)
	}

	// Already sorted newest first
	return &backups[0], nil
}

// BackupExists checks if any backup exists for a file
func BackupExists(filename string) bool {
	backups, err := GetBackupsForFile(filename)
	if err != nil {
		return false
	}
	return len(backups) > 0
}

// GetBackupCount returns the total number of backups
func GetBackupCount() (int, error) {
	backups, err := ListBackups()
	if err != nil {
		return 0, err
	}
	return len(backups), nil
}

// GetTotalBackupSize returns the total size of all backups
func GetTotalBackupSize() (int64, error) {
	backupDir, err := GetBackupDir()
	if err != nil {
		return 0, err
	}

	if !fs.PathExists(backupDir) {
		return 0, nil
	}

	return getDirSize(backupDir)
}
