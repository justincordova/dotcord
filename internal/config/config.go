package config

import "time"

// Config represents the DotCor configuration
type Config struct {
	RepoPath     string        `yaml:"repo_path"`
	GitEnabled   bool          `yaml:"git_enabled"`
	GitRemote    string        `yaml:"git_remote"`
	ManagedFiles []ManagedFile `yaml:"managed_files"`
}

// ManagedFile represents a single managed dotfile
type ManagedFile struct {
	SourcePath string    `yaml:"source_path"` // ~/.zshrc (normalized, with ~)
	RepoPath   string    `yaml:"repo_path"`   // shell/zshrc (relative to files/)
	AddedAt    time.Time `yaml:"added_at"`
	Platforms  []string  `yaml:"platforms"` // ["darwin", "linux"] or empty for all
}

// LoadConfig loads config from ~/.dotcor/config.yaml
func LoadConfig() (*Config, error) {
	// TODO: Implement using Viper
	return nil, nil
}

// SaveConfig writes config to ~/.dotcor/config.yaml
func (c *Config) SaveConfig() error {
	// TODO: Implement using Viper
	return nil
}

// AddManagedFile adds a new managed file
func (c *Config) AddManagedFile(mf ManagedFile) error {
	// TODO: Implement
	return nil
}

// RemoveManagedFile removes a managed file by source path
func (c *Config) RemoveManagedFile(sourcePath string) error {
	// TODO: Implement
	return nil
}

// GetManagedFile retrieves managed file by source path
func (c *Config) GetManagedFile(sourcePath string) (*ManagedFile, error) {
	// TODO: Implement
	return nil, nil
}

// IsManaged checks if a file is already managed
func (c *Config) IsManaged(sourcePath string) bool {
	// TODO: Implement
	return false
}

// GetManagedFilesForPlatform returns files that should be linked on current platform
func (c *Config) GetManagedFilesForPlatform() []ManagedFile {
	// TODO: Implement - filter by runtime.GOOS
	return nil
}
