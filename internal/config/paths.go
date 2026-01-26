package config

// NormalizePath converts absolute path to ~ notation
// Example: /Users/you/.zshrc -> ~/.zshrc
func NormalizePath(path string) (string, error) {
	// TODO: Implement using os.UserHomeDir()
	return "", nil
}

// ExpandPath converts ~ notation to absolute path
// Example: ~/.zshrc -> /Users/you/.zshrc
// Also handles environment variables: $XDG_CONFIG_HOME, %APPDATA%, etc.
func ExpandPath(path string) (string, error) {
	// TODO: Implement using os.UserHomeDir() and os.ExpandEnv()
	return "", nil
}

// GetRepoFilePath returns full path to file in repo
// Example: shell/zshrc -> ~/.dotcor/files/shell/zshrc
func GetRepoFilePath(config *Config, repoPath string) (string, error) {
	// TODO: Implement
	return "", nil
}

// GenerateRepoPath creates repo path from source path
// Example: ~/.config/nvim/init.vim -> nvim/init.vim
// Example: ~/.zshrc -> shell/zshrc
func GenerateRepoPath(sourcePath string) (string, error) {
	// TODO: Implement category mapping and path stripping
	return "", nil
}

// GetCurrentPlatform returns current OS identifier
// Returns: "darwin", "linux", or "windows"
func GetCurrentPlatform() string {
	// TODO: Implement using runtime.GOOS
	return ""
}

// ShouldApplyOnPlatform checks if file should be linked on current platform
func ShouldApplyOnPlatform(platforms []string) bool {
	// TODO: Implement - return true if platforms is empty or contains current platform
	return false
}
