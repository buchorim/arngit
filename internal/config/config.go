// Package config handles global configuration for arngit.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ThresholdType defines the type of auto-push threshold.
type ThresholdType string

const (
	ThresholdCommits ThresholdType = "commits"
	ThresholdTime    ThresholdType = "time"
	ThresholdSize    ThresholdType = "size"
)

// Threshold defines auto-push trigger conditions.
type Threshold struct {
	Type  ThresholdType `json:"type"`
	Value string        `json:"value"` // "5" for commits, "30m" for time, "100KB" for size
}

// Config represents the global arngit configuration.
type Config struct {
	DefaultThreshold Threshold `json:"default_threshold"`
	WatcherInterval  string    `json:"watcher_interval"`
	AutoInitRemote   bool      `json:"auto_init_remote"`
	DefaultBranch    string    `json:"default_branch"`
	LogLevel         string    `json:"log_level"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		DefaultThreshold: Threshold{
			Type:  ThresholdCommits,
			Value: "5",
		},
		WatcherInterval: "10s",
		AutoInitRemote:  true,
		DefaultBranch:   "main",
		LogLevel:        "info",
	}
}

// GetConfigDir returns the arngit config directory path.
func GetConfigDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(appData, "arngit"), nil
}

// GetConfigPath returns the path to config.json.
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// Load reads the config from disk, creating default if not exists.
func Load() (*Config, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, err
	}

	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := cfg.Save(); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Set updates a config value by key.
func (c *Config) Set(key, value string) error {
	switch key {
	case "threshold.type":
		c.DefaultThreshold.Type = ThresholdType(value)
	case "threshold.value":
		c.DefaultThreshold.Value = value
	case "watcher_interval":
		c.WatcherInterval = value
	case "auto_init_remote":
		c.AutoInitRemote = value == "true"
	case "default_branch":
		c.DefaultBranch = value
	case "log_level":
		c.LogLevel = value
	default:
		return &ConfigError{Key: key, Message: "unknown config key"}
	}
	return c.Save()
}

// Get retrieves a config value by key.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "threshold.type":
		return string(c.DefaultThreshold.Type), nil
	case "threshold.value":
		return c.DefaultThreshold.Value, nil
	case "watcher_interval":
		return c.WatcherInterval, nil
	case "auto_init_remote":
		if c.AutoInitRemote {
			return "true", nil
		}
		return "false", nil
	case "default_branch":
		return c.DefaultBranch, nil
	case "log_level":
		return c.LogLevel, nil
	default:
		return "", &ConfigError{Key: key, Message: "unknown config key"}
	}
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Key     string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error [" + e.Key + "]: " + e.Message
}
