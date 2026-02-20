package core

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	path string `yaml:"-"`

	// General settings
	DefaultAccount string `yaml:"default_account,omitempty"`
	Theme          string `yaml:"theme,omitempty"`

	// Update settings
	UpdateChannel  string `yaml:"update_channel,omitempty"`  // stable, beta, nightly
	UpdateInterval int    `yaml:"update_interval,omitempty"` // hours

	// UI settings
	CompactMode bool `yaml:"compact_mode,omitempty"`
	ColorOutput bool `yaml:"color_output,omitempty"`

	// Git settings
	DefaultBranch    string `yaml:"default_branch,omitempty"`
	AutoStage        bool   `yaml:"auto_stage,omitempty"`
	SignCommits      bool   `yaml:"sign_commits,omitempty"`
	GPGKeyID         string `yaml:"gpg_key_id,omitempty"`
	CommitTemplate   string `yaml:"commit_template,omitempty"`
	PushAfterCommit  bool   `yaml:"push_after_commit,omitempty"`
	ProtectedRepoDir string `yaml:"protected_repo_dir,omitempty"`
}

// DefaultConfig returns a new Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Theme:           "default",
		UpdateChannel:   "stable",
		UpdateInterval:  24,
		CompactMode:     false,
		ColorOutput:     true,
		DefaultBranch:   "main",
		AutoStage:       false,
		SignCommits:     false,
		PushAfterCommit: false,
	}
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()
	config.path = path

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config file
		if err := config.Save(); err != nil {
			return nil, err
		}
		return config, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// Save writes the configuration to the YAML file.
func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0600)
}

// Get returns a config value by key.
func (c *Config) Get(key string) interface{} {
	switch key {
	case "default_account":
		return c.DefaultAccount
	case "theme":
		return c.Theme
	case "update_channel":
		return c.UpdateChannel
	case "update_interval":
		return c.UpdateInterval
	case "compact_mode":
		return c.CompactMode
	case "color_output":
		return c.ColorOutput
	case "default_branch":
		return c.DefaultBranch
	case "auto_stage":
		return c.AutoStage
	case "sign_commits":
		return c.SignCommits
	case "gpg_key_id":
		return c.GPGKeyID
	case "commit_template":
		return c.CommitTemplate
	case "push_after_commit":
		return c.PushAfterCommit
	default:
		return nil
	}
}

// Set sets a config value by key.
func (c *Config) Set(key string, value interface{}) bool {
	switch key {
	case "default_account":
		c.DefaultAccount = value.(string)
	case "theme":
		c.Theme = value.(string)
	case "update_channel":
		c.UpdateChannel = value.(string)
	case "update_interval":
		c.UpdateInterval = value.(int)
	case "compact_mode":
		c.CompactMode = value.(bool)
	case "color_output":
		c.ColorOutput = value.(bool)
	case "default_branch":
		c.DefaultBranch = value.(string)
	case "auto_stage":
		c.AutoStage = value.(bool)
	case "sign_commits":
		c.SignCommits = value.(bool)
	case "gpg_key_id":
		c.GPGKeyID = value.(string)
	case "commit_template":
		c.CommitTemplate = value.(string)
	case "push_after_commit":
		c.PushAfterCommit = value.(bool)
	default:
		return false
	}
	return true
}
