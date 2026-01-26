package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/justincordova/dotcor/internal/config"
	"github.com/justincordova/dotcor/internal/fs"
)

// LockInfo contains information about the current lock
type LockInfo struct {
	PID       int
	Timestamp time.Time
	Hostname  string
}

// Lock timeout duration
const LockTimeout = 30 * time.Second

// ErrLockHeld is returned when lock is already held by another process
var ErrLockHeld = errors.New("lock is held by another process")

// ErrStaleLock is returned when lock appears to be stale
var ErrStaleLock = errors.New("stale lock detected")

// getLockPath returns the path to the lock file
func getLockPath() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".lock"), nil
}

// AcquireLock acquires file-based lock for dotcor operations
// Returns error if lock is already held
func AcquireLock() error {
	lockPath, err := getLockPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := fs.EnsureDir(filepath.Dir(lockPath)); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Check if lock file exists
	if fs.FileExists(lockPath) {
		// Check if lock is stale
		stale, err := IsStale(lockPath)
		if err != nil {
			return fmt.Errorf("checking stale lock: %w", err)
		}

		if stale {
			// Offer suggestion to clear stale lock
			info, _ := ReadLockInfo(lockPath)
			return fmt.Errorf("%w: PID %d (process appears dead). Run 'dotcor doctor --fix' to clear", ErrStaleLock, info.PID)
		}

		// Lock is held by active process
		info, _ := ReadLockInfo(lockPath)
		return fmt.Errorf("%w: PID %d on %s. If this is incorrect, run 'dotcor doctor --fix'", ErrLockHeld, info.PID, info.Hostname)
	}

	// Create lock file
	return writeLockFile(lockPath)
}

// ReleaseLock releases the file lock
func ReleaseLock() error {
	lockPath, err := getLockPath()
	if err != nil {
		return err
	}

	// Check if we own the lock
	if !fs.FileExists(lockPath) {
		return nil // No lock to release
	}

	info, err := ReadLockInfo(lockPath)
	if err != nil {
		// Can't read lock, try to remove anyway
		return os.Remove(lockPath)
	}

	// Only remove if we own it
	if info.PID != os.Getpid() {
		return fmt.Errorf("cannot release lock owned by PID %d", info.PID)
	}

	return os.Remove(lockPath)
}

// WithLock executes a function while holding the lock
// Automatically releases lock on completion or panic
func WithLock(fn func() error) error {
	if err := AcquireLock(); err != nil {
		return err
	}

	defer func() {
		ReleaseLock()
		if r := recover(); r != nil {
			panic(r) // Re-panic after releasing lock
		}
	}()

	return fn()
}

// IsLocked checks if lock is currently held
func IsLocked() (bool, error) {
	lockPath, err := getLockPath()
	if err != nil {
		return false, err
	}

	return fs.FileExists(lockPath), nil
}

// IsStale checks if lock file is stale (process dead)
func IsStale(lockPath string) (bool, error) {
	info, err := ReadLockInfo(lockPath)
	if err != nil {
		return true, nil // Malformed lock file is considered stale
	}

	// Check if lock is very old (more than 1 hour)
	if time.Since(info.Timestamp) > time.Hour {
		return true, nil
	}

	// Check if process is alive
	alive, err := isProcessAlive(info.PID)
	if err != nil {
		return true, nil // Can't check, assume stale
	}

	return !alive, nil
}

// ClearStaleLock removes stale lock file
func ClearStaleLock() error {
	lockPath, err := getLockPath()
	if err != nil {
		return err
	}

	if !fs.FileExists(lockPath) {
		return nil // No lock to clear
	}

	// Verify it's actually stale
	stale, err := IsStale(lockPath)
	if err != nil {
		return fmt.Errorf("checking if stale: %w", err)
	}

	if !stale {
		return fmt.Errorf("lock is not stale (process %d is still running)", os.Getpid())
	}

	return os.Remove(lockPath)
}

// ReadLockInfo reads lock information from lock file
func ReadLockInfo(lockPath string) (LockInfo, error) {
	content, err := os.ReadFile(lockPath)
	if err != nil {
		return LockInfo{}, fmt.Errorf("reading lock file: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 3 {
		return LockInfo{}, fmt.Errorf("malformed lock file: expected 3 lines, got %d", len(lines))
	}

	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return LockInfo{}, fmt.Errorf("invalid PID in lock file: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(lines[1]))
	if err != nil {
		return LockInfo{}, fmt.Errorf("invalid timestamp in lock file: %w", err)
	}

	hostname := strings.TrimSpace(lines[2])

	return LockInfo{
		PID:       pid,
		Timestamp: timestamp,
		Hostname:  hostname,
	}, nil
}

// writeLockFile writes lock information to file
func writeLockFile(lockPath string) error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	content := fmt.Sprintf("%d\n%s\n%s\n",
		os.Getpid(),
		time.Now().Format(time.RFC3339),
		hostname,
	)

	return os.WriteFile(lockPath, []byte(content), 0644)
}

// isProcessAlive checks if a process with given PID is still running
func isProcessAlive(pid int) (bool, error) {
	if runtime.GOOS == "windows" {
		return isProcessAliveWindows(pid)
	}
	return isProcessAliveUnix(pid)
}

// isProcessAliveUnix checks if process is alive on Unix systems
func isProcessAliveUnix(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil // Process doesn't exist
	}

	// On Unix, signal 0 checks if process exists without killing it
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// Process doesn't exist or we don't have permission
		return false, nil
	}

	return true, nil
}

// isProcessAliveWindows checks if process is alive on Windows
func isProcessAliveWindows(pid int) (bool, error) {
	// On Windows, FindProcess always succeeds
	// We need to try to open the process to check if it exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}

	// Try to signal - on Windows this will fail if process doesn't exist
	// We use a different approach: check if we can find process info
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false, nil
	}

	return true, nil
}

// GetLockInfo returns information about the current lock, if any
func GetLockInfo() (*LockInfo, error) {
	lockPath, err := getLockPath()
	if err != nil {
		return nil, err
	}

	if !fs.FileExists(lockPath) {
		return nil, nil // No lock
	}

	info, err := ReadLockInfo(lockPath)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// ForceReleaseLock forcibly removes the lock file regardless of owner
// Use with caution - only when you're sure the lock is stale
func ForceReleaseLock() error {
	lockPath, err := getLockPath()
	if err != nil {
		return err
	}

	if !fs.FileExists(lockPath) {
		return nil
	}

	return os.Remove(lockPath)
}

// IsOwnLock checks if current process owns the lock
func IsOwnLock() (bool, error) {
	info, err := GetLockInfo()
	if err != nil {
		return false, err
	}

	if info == nil {
		return false, nil // No lock exists
	}

	return info.PID == os.Getpid(), nil
}
