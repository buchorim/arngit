// Package core provides the core engine and infrastructure for ArnGit.
package core

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// Engine is the main application engine that manages all services.
type Engine struct {
	config         *Config
	storage        *Storage
	accounts       *AccountManager
	protectedRepos *ProtectedRepoManager
	updateManager  *UpdateManager
	logger         *Logger

	// Version info
	version   string
	buildTime string
	gitCommit string

	// State
	startTime time.Time
	mu        sync.RWMutex
}

// NewEngine creates and initializes a new Engine instance.
func NewEngine() (*Engine, error) {
	startTime := time.Now()

	// Initialize storage first
	storage, err := NewStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Load or create config
	configPath := filepath.Join(storage.ConfigDir(), "config.yaml")
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize account manager
	accounts, err := NewAccountManager(storage.AccountsDir())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize accounts: %w", err)
	}

	// Initialize protected repos manager
	protectedPath := filepath.Join(storage.ConfigDir(), "protected.json")
	protectedRepos, err := NewProtectedRepoManager(protectedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize protected repos: %w", err)
	}

	// Initialize logger
	logPath := filepath.Join(storage.LogsDir(), "arngit.log")
	logger, err := NewLogger(logPath, 1000) // Keep last 1000 entries
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize update manager
	updateManager := NewUpdateManager(config, storage)

	engine := &Engine{
		config:         config,
		storage:        storage,
		accounts:       accounts,
		protectedRepos: protectedRepos,
		updateManager:  updateManager,
		logger:         logger,
		startTime:      startTime,
	}

	// Log startup
	logger.Info("ArnGit engine started")

	// Check for updates in background if configured
	if config.UpdateInterval > 0 {
		go engine.checkUpdatesBackground()
	}

	return engine, nil
}

// Close releases all resources held by the engine.
func (e *Engine) Close() error {
	e.logger.Info("ArnGit engine shutting down")
	return e.logger.Close()
}

// Uptime returns how long the engine has been running.
func (e *Engine) Uptime() time.Duration {
	return time.Since(e.startTime)
}

// SetVersion sets the version information.
func (e *Engine) SetVersion(version, buildTime, gitCommit string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.version = version
	e.buildTime = buildTime
	e.gitCommit = gitCommit
}

// Version returns the current version string.
func (e *Engine) Version() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.version
}

// BuildTime returns the build timestamp.
func (e *Engine) BuildTime() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.buildTime
}

// GitCommit returns the git commit hash.
func (e *Engine) GitCommit() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.gitCommit
}

// Config returns the configuration.
func (e *Engine) Config() *Config {
	return e.config
}

// Storage returns the storage manager.
func (e *Engine) Storage() *Storage {
	return e.storage
}

// Accounts returns the account manager.
func (e *Engine) Accounts() *AccountManager {
	return e.accounts
}

// ProtectedRepos returns the protected repository manager.
func (e *Engine) ProtectedRepos() *ProtectedRepoManager {
	return e.protectedRepos
}

// UpdateManager returns the update manager.
func (e *Engine) UpdateManager() *UpdateManager {
	return e.updateManager
}

// Logger returns the logger.
func (e *Engine) Logger() *Logger {
	return e.logger
}

// checkUpdatesBackground periodically checks for updates.
func (e *Engine) checkUpdatesBackground() {
	// Initial delay
	time.Sleep(5 * time.Second)

	for {
		if update, err := e.updateManager.CheckForUpdate(e.version); err == nil && update != nil {
			e.logger.Info(fmt.Sprintf("Update available: %s", update.Version))
		}

		// Sleep for configured interval
		interval := time.Duration(e.config.UpdateInterval) * time.Hour
		if interval < time.Hour {
			interval = time.Hour
		}
		time.Sleep(interval)
	}
}
