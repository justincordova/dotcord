package fs

// MoveFile moves a file from src to dst
func MoveFile(src, dst string) error {
	// TODO: Implement using os.Rename or copy+delete for cross-device moves
	return nil
}

// CopyFile copies file with permissions preserved (Windows fallback)
func CopyFile(src, dst string) error {
	// TODO: Implement using io.Copy and preserve permissions
	return nil
}

// FileExists checks if file exists
func FileExists(path string) bool {
	// TODO: Implement using os.Stat
	return false
}

// EnsureDir creates directory if it doesn't exist
func EnsureDir(path string) error {
	// TODO: Implement using os.MkdirAll
	return nil
}

// IsDirectory checks if path is a directory
func IsDirectory(path string) (bool, error) {
	// TODO: Implement using os.Stat
	return false, nil
}
