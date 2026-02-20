package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Account represents a GitHub account.
type Account struct {
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	PAT       string    `json:"pat"` // Encrypted
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountManager handles multiple GitHub accounts.
type AccountManager struct {
	dir      string
	accounts map[string]*Account
	current  string
}

// encryptionKey derives a key from machine-specific data.
func encryptionKey() []byte {
	hostname, _ := os.Hostname()
	homeDir, _ := os.UserHomeDir()
	seed := hostname + homeDir + "arngit-secret-sauce"
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}

// encrypt encrypts plaintext using AES-256-GCM.
func encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(encryptionKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts ciphertext using AES-256-GCM.
func decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], string(data[gcm.NonceSize():])
	plaintext, err := gcm.Open(nil, nonce, []byte(ciphertext), nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// NewAccountManager creates a new AccountManager.
func NewAccountManager(dir string) (*AccountManager, error) {
	am := &AccountManager{
		dir:      dir,
		accounts: make(map[string]*Account),
	}

	// Load existing accounts
	if err := am.loadAll(); err != nil {
		return nil, err
	}

	return am, nil
}

// loadAll loads all accounts from disk.
func (am *AccountManager) loadAll() error {
	files, err := filepath.Glob(filepath.Join(am.dir, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var acc Account
		if err := json.Unmarshal(data, &acc); err != nil {
			continue
		}

		am.accounts[acc.Name] = &acc
		if acc.IsDefault {
			am.current = acc.Name
		}
	}

	return nil
}

// Add adds a new account.
func (am *AccountManager) Add(name, username, email, pat string) error {
	if _, exists := am.accounts[name]; exists {
		return fmt.Errorf("account '%s' already exists", name)
	}

	encryptedPAT, err := encrypt(strings.TrimSpace(pat))
	if err != nil {
		return fmt.Errorf("failed to encrypt PAT: %w", err)
	}

	acc := &Account{
		Name:      name,
		Username:  username,
		Email:     email,
		PAT:       encryptedPAT,
		IsDefault: len(am.accounts) == 0, // First account is default
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := am.save(acc); err != nil {
		return err
	}

	am.accounts[name] = acc
	if acc.IsDefault {
		am.current = name
	}

	return nil
}

// Remove removes an account.
func (am *AccountManager) Remove(name string) error {
	if _, exists := am.accounts[name]; !exists {
		return fmt.Errorf("account '%s' not found", name)
	}

	path := filepath.Join(am.dir, name+".json")
	if err := os.Remove(path); err != nil {
		return err
	}

	delete(am.accounts, name)

	// Reset current if needed
	if am.current == name {
		am.current = ""
		for n := range am.accounts {
			am.current = n
			break
		}
	}

	return nil
}

// Switch switches to a different account.
func (am *AccountManager) Switch(name string) error {
	if _, exists := am.accounts[name]; !exists {
		return fmt.Errorf("account '%s' not found", name)
	}

	// Update default flags
	for n, acc := range am.accounts {
		wasDefault := acc.IsDefault
		acc.IsDefault = (n == name)
		if wasDefault != acc.IsDefault {
			am.save(acc)
		}
	}

	am.current = name
	return nil
}

// Current returns the current active account.
func (am *AccountManager) Current() *Account {
	if am.current == "" {
		return nil
	}
	return am.accounts[am.current]
}

// CurrentName returns the name of the current account.
func (am *AccountManager) CurrentName() string {
	return am.current
}

// Get returns an account by name.
func (am *AccountManager) Get(name string) *Account {
	return am.accounts[name]
}

// List returns all account names.
func (am *AccountManager) List() []string {
	names := make([]string, 0, len(am.accounts))
	for name := range am.accounts {
		names = append(names, name)
	}
	return names
}

// GetPAT returns the decrypted PAT for an account.
func (am *AccountManager) GetPAT(name string) (string, error) {
	acc, exists := am.accounts[name]
	if !exists {
		return "", fmt.Errorf("account '%s' not found", name)
	}

	pat, err := decrypt(acc.PAT)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pat), nil
}

// save saves an account to disk.
func (am *AccountManager) save(acc *Account) error {
	data, err := json.MarshalIndent(acc, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(am.dir, acc.Name+".json")
	return os.WriteFile(path, data, 0600)
}
