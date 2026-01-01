// Package account manages multiple git accounts.
package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arinara/arngit/internal/config"
	"github.com/arinara/arngit/internal/credential"
)

const credentialPrefix = "arngit:"

// Account represents a git account configuration.
type Account struct {
	Name          string    `json:"name"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	CredentialKey string    `json:"credential_key"`
	CreatedAt     time.Time `json:"created_at"`
}

// AccountStore manages the accounts.json file.
type AccountStore struct {
	Active   string              `json:"active"`
	Accounts map[string]*Account `json:"accounts"`
}

// getAccountsPath returns the path to accounts.json.
func getAccountsPath() (string, error) {
	dir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "accounts.json"), nil
}

// LoadStore loads the account store from disk.
func LoadStore() (*AccountStore, error) {
	if err := config.EnsureConfigDir(); err != nil {
		return nil, err
	}

	path, err := getAccountsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		store := &AccountStore{
			Accounts: make(map[string]*Account),
		}
		return store, nil
	}
	if err != nil {
		return nil, err
	}

	var store AccountStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	if store.Accounts == nil {
		store.Accounts = make(map[string]*Account)
	}

	return &store, nil
}

// Save writes the account store to disk.
func (s *AccountStore) Save() error {
	path, err := getAccountsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Add creates a new account with the given details.
func (s *AccountStore) Add(name, username, email, pat string) error {
	if _, exists := s.Accounts[name]; exists {
		return fmt.Errorf("account '%s' already exists", name)
	}

	credKey := credentialPrefix + name

	// Store PAT in Windows Credential Manager
	if err := credential.Store(credKey, username, pat); err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}

	s.Accounts[name] = &Account{
		Name:          name,
		Username:      username,
		Email:         email,
		CredentialKey: credKey,
		CreatedAt:     time.Now(),
	}

	// Set as active if first account
	if s.Active == "" {
		s.Active = name
	}

	return s.Save()
}

// Remove deletes an account.
func (s *AccountStore) Remove(name string) error {
	acc, exists := s.Accounts[name]
	if !exists {
		return fmt.Errorf("account '%s' not found", name)
	}

	// Remove credential from Windows Credential Manager
	_ = credential.Delete(acc.CredentialKey)

	delete(s.Accounts, name)

	// Update active if needed
	if s.Active == name {
		s.Active = ""
		for n := range s.Accounts {
			s.Active = n
			break
		}
	}

	return s.Save()
}

// Switch changes the active account.
func (s *AccountStore) Switch(name string) error {
	if _, exists := s.Accounts[name]; !exists {
		return fmt.Errorf("account '%s' not found", name)
	}

	s.Active = name
	return s.Save()
}

// GetActive returns the currently active account.
func (s *AccountStore) GetActive() (*Account, error) {
	if s.Active == "" {
		return nil, errors.New("no active account, use 'arngit account add' to add one")
	}

	acc, exists := s.Accounts[s.Active]
	if !exists {
		return nil, fmt.Errorf("active account '%s' not found", s.Active)
	}

	return acc, nil
}

// GetCredentials retrieves the PAT for an account.
func (s *AccountStore) GetCredentials(name string) (username, pat string, err error) {
	acc, exists := s.Accounts[name]
	if !exists {
		return "", "", fmt.Errorf("account '%s' not found", name)
	}

	return credential.Retrieve(acc.CredentialKey)
}

// GetActiveCredentials retrieves credentials for the active account.
func (s *AccountStore) GetActiveCredentials() (username, pat string, err error) {
	if s.Active == "" {
		return "", "", errors.New("no active account")
	}
	return s.GetCredentials(s.Active)
}

// List returns all account names.
func (s *AccountStore) List() []string {
	names := make([]string, 0, len(s.Accounts))
	for name := range s.Accounts {
		names = append(names, name)
	}
	return names
}

// Get returns an account by name.
func (s *AccountStore) Get(name string) (*Account, error) {
	acc, exists := s.Accounts[name]
	if !exists {
		return nil, fmt.Errorf("account '%s' not found", name)
	}
	return acc, nil
}

// Update modifies an existing account.
func (s *AccountStore) Update(name, username, email, pat string) error {
	acc, exists := s.Accounts[name]
	if !exists {
		return fmt.Errorf("account '%s' not found", name)
	}

	if username != "" {
		acc.Username = username
	}
	if email != "" {
		acc.Email = email
	}
	if pat != "" {
		if err := credential.Store(acc.CredentialKey, acc.Username, pat); err != nil {
			return fmt.Errorf("failed to update credential: %w", err)
		}
	}

	return s.Save()
}
