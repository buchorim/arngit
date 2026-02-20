package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// Storage manages the file system storage for ArnGit.
type Storage struct {
	baseDir string
}

// NewStorage creates a new Storage instance with the default base directory.
func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".arngit")
	s := &Storage{baseDir: baseDir}

	// Create directory structure
	dirs := []string{
		s.baseDir,
		s.ConfigDir(),
		s.AccountsDir(),
		s.PluginsDir(),
		s.CacheDir(),
		s.LogsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return s, nil
}

// BaseDir returns the base storage directory.
func (s *Storage) BaseDir() string {
	return s.baseDir
}

// ConfigDir returns the configuration directory.
func (s *Storage) ConfigDir() string {
	return filepath.Join(s.baseDir, "config")
}

// AccountsDir returns the accounts directory.
func (s *Storage) AccountsDir() string {
	return filepath.Join(s.baseDir, "accounts")
}

// PluginsDir returns the plugins directory.
func (s *Storage) PluginsDir() string {
	return filepath.Join(s.baseDir, "plugins")
}

// CacheDir returns the cache directory.
func (s *Storage) CacheDir() string {
	return filepath.Join(s.baseDir, "cache")
}

// LogsDir returns the logs directory.
func (s *Storage) LogsDir() string {
	return filepath.Join(s.baseDir, "logs")
}

// GetStorageStats returns storage usage statistics.
func (s *Storage) GetStorageStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	dirs := map[string]string{
		"config":   s.ConfigDir(),
		"accounts": s.AccountsDir(),
		"plugins":  s.PluginsDir(),
		"cache":    s.CacheDir(),
		"logs":     s.LogsDir(),
	}

	for name, dir := range dirs {
		size, err := s.dirSize(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate size for %s: %w", name, err)
		}
		stats[name] = size
	}

	return stats, nil
}

// dirSize calculates the total size of a directory.
func (s *Storage) dirSize(path string) (int64, error) {
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
