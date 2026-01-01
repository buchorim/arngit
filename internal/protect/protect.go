// Package protect manages repository protection.
package protect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/arinara/arngit/internal/config"
)

// ProtectedRepo represents a protected repository entry.
type ProtectedRepo struct {
	Reason      string    `json:"reason"`
	ProtectedAt time.Time `json:"protected_at"`
}

// ProtectionStore manages the protected.json file.
type ProtectionStore struct {
	Repos map[string]*ProtectedRepo `json:"repos"`
}

// getProtectedPath returns the path to protected.json.
func getProtectedPath() (string, error) {
	dir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "protected.json"), nil
}

// LoadStore loads the protection store from disk.
func LoadStore() (*ProtectionStore, error) {
	if err := config.EnsureConfigDir(); err != nil {
		return nil, err
	}

	path, err := getProtectedPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		store := &ProtectionStore{
			Repos: make(map[string]*ProtectedRepo),
		}
		return store, nil
	}
	if err != nil {
		return nil, err
	}

	var store ProtectionStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	if store.Repos == nil {
		store.Repos = make(map[string]*ProtectedRepo)
	}

	return &store, nil
}

// Save writes the protection store to disk.
func (s *ProtectionStore) Save() error {
	path, err := getProtectedPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Protect marks a repository as protected.
func (s *ProtectionStore) Protect(repoPath, reason string) error {
	// Normalize path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return err
	}

	s.Repos[absPath] = &ProtectedRepo{
		Reason:      reason,
		ProtectedAt: time.Now(),
	}

	return s.Save()
}

// Unprotect removes protection from a repository.
func (s *ProtectionStore) Unprotect(repoPath string) error {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return err
	}

	delete(s.Repos, absPath)
	return s.Save()
}

// IsProtected checks if a repository is protected.
func (s *ProtectionStore) IsProtected(repoPath string) bool {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return false
	}

	_, exists := s.Repos[absPath]
	return exists
}

// GetProtection returns the protection info for a repository.
func (s *ProtectionStore) GetProtection(repoPath string) (*ProtectedRepo, bool) {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, false
	}

	repo, exists := s.Repos[absPath]
	return repo, exists
}

// List returns all protected repositories.
func (s *ProtectionStore) List() map[string]*ProtectedRepo {
	return s.Repos
}
