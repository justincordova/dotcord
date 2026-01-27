# DotCor Code Review

## Critical Issues

### 1. **Race Condition in Lock Acquisition** (internal/core/lock.go:56-76)
**Severity:** High
**Location:** `AcquireLock()` function

The lock implementation has a classic TOCTOU (time-of-check-time-of-use) race condition:
```go
// Check if lock file exists
if fs.FileExists(lockPath) {
    // ... check if stale ...
}
// Later, create lock file
return writeLockFile(lockPath)
```

Between checking if the file exists and creating it, another process can acquire the lock.

**Fix:** Use atomic file creation with `O_EXCL` flag:
```go
f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
if err != nil {
    if os.IsExist(err) {
        // Lock exists, check if stale
    }
    return err
}
defer f.Close()
// Write lock content to f
```

### 2. **Wrong PID Displayed in ClearStaleLock** (internal/core/lock.go:170-171)
**Severity:** Medium
**Location:** `ClearStaleLock()` function

```go
if !stale {
    return fmt.Errorf("lock is not stale (process %d is still running)", os.Getpid())
}
```

Displays the current process PID instead of the lock owner's PID. Should be `info.PID`.

### 3. **Category Flag Misuse** (cmd/dotcor/add.go:190-193)
**Severity:** Medium
**Location:** `processAddFile()` function

The `--category` flag is meant to specify a category like "shell" or "vim", but it's used directly as the full repo path:
```go
customRepoPath := ""
if category != "" {
    customRepoPath = category
}
repoPath, err := config.GenerateRepoPath(sourcePath, customRepoPath)
```

If user passes `--category shell` for `~/.zshrc`, the repo path becomes just "shell" instead of "shell/zshrc".

**Fix:** The category should be combined with the filename, not replace the entire path.

### 4. **Transaction Operations Field Unused** (internal/core/transaction.go:20, 335)
**Severity:** Medium
**Location:** `Transaction` struct and `AddFileTransaction()`

The `operations` field in Transaction is populated by `AddFileTransaction()` but never used:
```go
type Transaction struct {
    operations []Operation  // Never read
    executed   []Operation  // Actually used
    committed  bool
}
```

The `AddFileTransaction` builds operations into `tx.operations` but returns the transaction without executing them. This is confusing API design.

**Fix:** Either execute operations in the function, or rename to make it clear it's building a transaction template.

## High Priority Issues

### 5. **Incomplete Windows Lock Implementation** (internal/core/lock.go:249-266)
**Severity:** High
**Location:** `isProcessAliveWindows()` function

The Windows implementation claims to use a "different approach" but actually uses the same Unix method which doesn't work on Windows:
```go
// Comment says: "We use a different approach"
// But then does:
err = process.Signal(syscall.Signal(0))
```

On Windows, you need to use Windows API calls like `OpenProcess` to check if a process exists.

### 6. **Silent Error Swallowing in Backup Cleanup** (internal/core/backup.go:230-232)
**Severity:** Medium
**Location:** `CleanOldBackups()` function

```go
for _, candidate := range candidates {
    if err := fs.RemoveAll(candidate.Path); err != nil {
        // Continue deleting others even if one fails
        continue
    }
    deleted++
}
```

Deletion errors are silently ignored. Should at least log or collect errors.

### 7. **Ignored Size Calculation Errors** (internal/core/backup.go:301)
**Severity:** Medium
**Location:** `getCleanupCandidates()` function

```go
size, _ := getDirSize(dir.path)  // Error ignored
totalSize += size
```

If size calculation fails, it uses 0 which leads to incorrect reporting to the user.

### 8. **Backup Failure is Non-Fatal** (cmd/dotcor/add.go:211-215)
**Severity:** High
**Location:** `processAddFile()` function

```go
backupPath, err := core.CreateBackup(expanded)
if err != nil {
    // Non-fatal, continue but warn
    fmt.Printf("  ⚠ Backup failed for %s: %v\n", normalized, err)
}
```

If backup fails, the operation continues anyway. This means if something goes wrong later, the user has no safety net to recover their original file.

**Recommendation:** Backup should be mandatory for destructive operations.

### 9. **Return Value Inconsistency** (cmd/dotcor/add.go:207, 246)
**Severity:** Medium
**Location:** `processAddFile()` function

Dry-run returns relative path, actual execution returns full path:
```go
if dryRun {
    return addResultSuccess, repoPath, nil  // Relative
}
// ...
return addResultSuccess, fullRepoPath, nil  // Absolute
```

Callers append these to `gitFiles`, creating a mix of full and relative paths.

### 10. **Error Ignored in Git Operations** (cmd/dotcor/add.go:118, cmd/dotcor/remove.go:131)
**Severity:** Low
**Location:** Multiple git commit locations

```go
repoPath, _ := config.ExpandPath(cfg.RepoPath)  // Error ignored
```

If ExpandPath fails, malformed path is used for git operations.

## Medium Priority Issues

### 11. **Inconsistent Error Handling in Path Expansion** (internal/fs/symlink.go)
**Severity:** Medium
**Location:** Multiple functions (IsSymlink, ReadSymlink, etc.)

Throughout symlink.go, when `config.ExpandPath()` fails, the code falls back to the unexpanded path:
```go
expandedPath, err := config.ExpandPath(path)
if err != nil {
    expandedPath = path  // Fallback
}
```

This could treat `~/.zshrc` as a literal file named `~` in the current directory rather than failing fast.

### 12. **Race Condition in CreateSymlink** (internal/fs/symlink.go:62-66)
**Severity:** Low
**Location:** `CreateSymlink()` function

TOCTOU between checking if file exists and removing it:
```go
if _, err := os.Lstat(expandedLink); err == nil {
    if err := os.Remove(expandedLink); err != nil {
        return fmt.Errorf("removing existing file: %w", err)
    }
}
```

Another process could modify the file between Lstat and Remove.

### 13. **No Transaction Rollback Protection** (cmd/dotcor/remove.go:164-179)
**Severity:** Medium
**Location:** `processRemoveFile()` function

Remove operations don't use transactions. If config update fails after removing the symlink, the symlink is deleted but config still shows it as managed.

### 14. **Silently Ignored IsSymlink Error** (cmd/dotcor/remove.go:165)
**Severity:** Medium
**Location:** `processRemoveFile()` function

```go
isLink, _ := fs.IsSymlink(sourcePath)  // Error ignored
```

If IsSymlink fails (permission denied, etc.), `isLink` will be false and code assumes it's not a symlink, leading to incorrect behavior.

### 15. **Data Loss Risk When keepRepo=true** (cmd/dotcor/remove.go:168-178)
**Severity:** Medium
**Location:** `processRemoveFile()` function

If source path is NOT a symlink (user manually replaced it), and `keepRepo=true`, the code only removes from config without restoring the repo file. This orphans the repo file.

### 16. **LockTimeout Constant Unused** (internal/core/lock.go:26, 140)
**Severity:** Low
**Location:** Lock timeout definitions

`LockTimeout` is defined as 30 seconds but stale check uses 1 hour hardcoded:
```go
const LockTimeout = 30 * time.Second  // Line 26
// ...
if time.Since(info.Timestamp) > time.Hour {  // Line 140
```

### 17. **Insufficient Path Traversal Validation** (internal/core/validator.go:118)
**Severity:** Low
**Location:** `ValidateRepoPath()` function

Only checks for literal `".."`:
```go
if strings.Contains(path, "..") {
    return fmt.Errorf("repo path cannot contain '..': %s", path)
}
```

Doesn't catch URL-encoded or other bypass attempts (though low risk for this use case).

## Low Priority / Code Quality Issues

### 18. **Parameter Order Documentation Mismatch** (internal/fs/symlink.go:27)
**Severity:** Low
**Location:** `CreateSymlink()` function comment

Comment says "from link to target" but signature is `(target, link)`, which is confusing.

### 19. **Silent Cleanup Errors** (internal/fs/symlink.go:189-191)
**Severity:** Low
**Location:** `SupportsSymlinks()` function

Test file cleanup ignores errors, potentially leaving orphaned test files.

### 20. **Potential Command Injection False Alarm** (internal/git/git.go:74)
**Severity:** None
**Location:** `AutoCommit()` function

Initially appeared to have command injection risk, but `exec.Command()` treats arguments separately, so this is safe. No issue.

### 21. **Git Push Without Upstream** (internal/git/git.go:104)
**Severity:** Low
**Location:** `Sync()` function

```go
pushCmd := exec.Command("git", "push")
```

Will fail if branch doesn't have upstream tracking set. Should check or use `-u origin <branch>`.

### 22. **Secret Detection False Positives** (internal/core/validator.go:23-24, 43-45)
**Severity:** Low
**Location:** Secret detection regex patterns

Patterns will match intentional examples in comments/docs:
- `password=validpassword` in code comments
- `postgres://user:${DB_PASS}@localhost` (placeholder)
- `export PASSWORD_POLICY=minimum8chars` (policy setting)

### 23. **Regex Performance** (internal/core/validator.go:193-205)
**Severity:** Low
**Location:** `DetectSecrets()` function

Scans every line against 15+ regexes. For large files (1000 lines), that's 15,000+ regex operations. Could optimize by:
- Limiting file size for secret scanning
- Using faster string matching for simple patterns

### 24. **Transaction Rollback Error Handling** (internal/core/transaction.go:56-68)
**Severity:** Low
**Location:** `Rollback()` function

Only returns the last rollback error. Multiple rollback failures are silently lost:
```go
for i := len(t.executed) - 1; i >= 0; i-- {
    if err := op.Undo(); err != nil {
        lastErr = err  // Overwrites previous errors
    }
}
```

Should collect all errors or use multierror.

### 25. **No Thread Safety** (internal/core/transaction.go)
**Severity:** Low
**Location:** `Transaction` struct

Transaction has no mutex, so concurrent Execute() or Rollback() calls could cause race conditions. Probably not an issue given the usage pattern, but worth noting.

### 26. **Timestamp Collision** (internal/core/backup.go:56-57)
**Severity:** Low
**Location:** `CreateBackup()` function

Two backups created within the same second share a timestamp directory. File collisions are handled, but could be cleaner with millisecond precision in timestamps.

### 27. **Race in AutoCommit** (internal/git/git.go:57-64)
**Severity:** Low
**Location:** `AutoCommit()` function

Between checking HasChanges() and staging files, another process could modify the repo. Low risk in practice.

## Summary

- **Critical Issues:** 4 (race condition in lock, wrong PID, category flag misuse, unused transaction field)
- **High Priority:** 6 (Windows lock, backup errors, size errors, backup non-fatal, return inconsistency, git path errors)
- **Medium Priority:** 11 (path expansion, symlink race, transaction protection, error swallowing, data loss risk, etc.)
- **Low Priority:** 7 (documentation, cleanup errors, performance, false positives, etc.)

Total Issues: **28**

## Recommendations

1. **Fix critical lock race condition** - Use atomic file creation with O_EXCL
2. **Implement proper Windows lock checking** - Use Windows API instead of Unix signals
3. **Make backups mandatory** - Don't continue if backup fails
4. **Add transaction protection to remove operations** - Prevent partial failures
5. **Fix category flag behavior** - Combine category with filename correctly
6. **Improve error handling** - Don't silently ignore errors, especially in critical paths
7. **Consider adding mutex to Transaction** - If concurrent usage is possible
8. **Review all path expansion fallbacks** - Fail fast instead of using unexpanded paths
9. **Collect all rollback errors** - Use multierror or error slice
10. **Add integration tests for race conditions** - Test concurrent operations
