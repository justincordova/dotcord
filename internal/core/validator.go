package core

import "github.com/justincordova/dotcor/internal/config"

// ValidateSourceFile checks if source file is valid for adding
func ValidateSourceFile(path string) error {
	// TODO: Implement
	// - File must exist
	// - File must be readable
	// - Can be file or directory
	// - Path must be absolute or start with ~ or contain env variables
	// - Must not already be a symlink pointing to our repo
	return nil
}

// ValidateRepoPath checks if repo path is valid
func ValidateRepoPath(path string) error {
	// TODO: Implement
	return nil
}

// ValidateNotAlreadyManaged checks if file is not already managed
func ValidateNotAlreadyManaged(cfg *config.Config, sourcePath string) error {
	// TODO: Implement
	return nil
}
