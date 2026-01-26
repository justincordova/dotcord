package fs

// CreateSymlink creates a symlink from link to target
// On Windows: falls back to copying if symlink fails (no admin/dev mode)
func CreateSymlink(target, link string) error {
	// TODO: Implement
	// 1. Ensure parent directory exists
	// 2. Try os.Symlink(target, link)
	// 3. On Windows, if fails, fall back to CopyFile with warning
	return nil
}

// RemoveSymlink removes a symlink
func RemoveSymlink(link string) error {
	// TODO: Implement using os.Remove
	return nil
}

// IsSymlink checks if path is a symlink
func IsSymlink(path string) (bool, error) {
	// TODO: Implement using os.Lstat
	return false, nil
}

// ReadSymlink reads the target of a symlink
func ReadSymlink(link string) (string, error) {
	// TODO: Implement using os.Readlink
	return "", nil
}

// IsValidSymlink checks if symlink exists and points to existing target
func IsValidSymlink(link string) (bool, error) {
	// TODO: Implement - check both symlink and target exist
	return false, nil
}

// SupportsSymlinks checks if current platform supports symlinks
// Windows: requires admin rights or developer mode
func SupportsSymlinks() bool {
	// TODO: Implement - try creating a test symlink
	return false
}
