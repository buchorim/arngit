package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ProtectedRepo represents a protected repository.
type ProtectedRepo struct {
	Path         string    `json:"path"`
	Password     string    `json:"password,omitempty"` // Hashed password
	ProtectedAt  time.Time `json:"protected_at"`
	LastAccessed time.Time `json:"last_accessed,omitempty"`
}

// ProtectedRepoManager manages protected repositories.
type ProtectedRepoManager struct {
	path  string
	repos map[string]*ProtectedRepo
	mu    sync.RWMutex
}

// NewProtectedRepoManager creates a new protected repo manager.
func NewProtectedRepoManager(path string) (*ProtectedRepoManager, error) {
	pm := &ProtectedRepoManager{
		path:  path,
		repos: make(map[string]*ProtectedRepo),
	}

	// Load existing protected repos
	if err := pm.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return pm, nil
}

// load loads protected repos from disk.
func (pm *ProtectedRepoManager) load() error {
	data, err := os.ReadFile(pm.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &pm.repos)
}

// save saves protected repos to disk.
func (pm *ProtectedRepoManager) save() error {
	data, err := json.MarshalIndent(pm.repos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pm.path, data, 0600)
}

// IsProtected checks if a path is in a protected repository.
func (pm *ProtectedRepoManager) IsProtected(repoPath string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Normalize path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return false
	}

	// Check if this path or any parent is protected
	for protectedPath := range pm.repos {
		if absPath == protectedPath || isSubPath(absPath, protectedPath) {
			return true
		}
	}

	return false
}

// GetProtection returns protection details for a repo.
func (pm *ProtectedRepoManager) GetProtection(repoPath string) *ProtectedRepo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil
	}

	// Check exact match first
	if repo, ok := pm.repos[absPath]; ok {
		return repo
	}

	// Check parents
	for protectedPath, repo := range pm.repos {
		if isSubPath(absPath, protectedPath) {
			return repo
		}
	}

	return nil
}

// Protect adds a repository to the protected list.
func (pm *ProtectedRepoManager) Protect(repoPath string, password string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return err
	}

	hashedPassword := ""
	if password != "" {
		hashedPassword = hashPassword(password)
	}

	pm.repos[absPath] = &ProtectedRepo{
		Path:        absPath,
		Password:    hashedPassword,
		ProtectedAt: time.Now(),
	}

	return pm.save()
}

// Unprotect removes a repository from the protected list.
func (pm *ProtectedRepoManager) Unprotect(repoPath string, password string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return err
	}

	repo, ok := pm.repos[absPath]
	if !ok {
		return nil // Not protected
	}

	// Verify password if set
	if repo.Password != "" {
		if !verifyPassword(password, repo.Password) {
			return NewError(ErrGitProtected, nil)
		}
	}

	delete(pm.repos, absPath)
	return pm.save()
}

// VerifyAccess verifies access to a protected repo.
func (pm *ProtectedRepoManager) VerifyAccess(repoPath string, password string) bool {
	repo := pm.GetProtection(repoPath)
	if repo == nil {
		return true // Not protected
	}

	if repo.Password == "" {
		return true // No password set
	}

	return verifyPassword(password, repo.Password)
}

// UpdateLastAccessed updates the last access time.
func (pm *ProtectedRepoManager) UpdateLastAccessed(repoPath string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	absPath, _ := filepath.Abs(repoPath)
	if repo, ok := pm.repos[absPath]; ok {
		repo.LastAccessed = time.Now()
		pm.save()
	}
}

// List returns all protected repositories.
func (pm *ProtectedRepoManager) List() []*ProtectedRepo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	repos := make([]*ProtectedRepo, 0, len(pm.repos))
	for _, repo := range pm.repos {
		repos = append(repos, repo)
	}
	return repos
}

// isSubPath checks if path is inside parentPath.
func isSubPath(path, parentPath string) bool {
	rel, err := filepath.Rel(parentPath, path)
	if err != nil {
		return false
	}
	return rel != ".." && !filepath.IsAbs(rel) && rel != "."
}

// hashPassword creates a simple hash (should use bcrypt in production).
func hashPassword(password string) string {
	// Simple hash for now - in production use bcrypt
	hash := uint64(0)
	for i, c := range password {
		hash = hash*31 + uint64(c) + uint64(i)
	}
	return string(rune(hash))
}

// verifyPassword verifies a password against a hash.
func verifyPassword(password, hash string) bool {
	return hashPassword(password) == hash
}
