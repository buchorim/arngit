package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// UpdateInfo contains information about an available update.
type UpdateInfo struct {
	Version     string    `json:"version"`
	ReleaseURL  string    `json:"release_url"`
	DownloadURL string    `json:"download_url"`
	ReleaseDate time.Time `json:"release_date"`
	Changelog   string    `json:"changelog"`
	Size        int64     `json:"size"`
}

// UpdateManager handles application updates.
type UpdateManager struct {
	config  *Config
	storage *Storage

	// GitHub release info
	owner string
	repo  string

	// State
	lastCheck     time.Time
	latestUpdate  *UpdateInfo
	updatePending bool
}

// NewUpdateManager creates a new update manager.
func NewUpdateManager(config *Config, storage *Storage) *UpdateManager {
	return &UpdateManager{
		config:  config,
		storage: storage,
		owner:   "arfrfrr",
		repo:    "arngit-releases",
	}
}

// CheckForUpdate checks GitHub releases for updates.
func (um *UpdateManager) CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	um.lastCheck = time.Now()

	// Fetch latest release from GitHub
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", um.owner, um.repo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil // No releases yet
		}
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var release struct {
		TagName     string `json:"tag_name"`
		HTMLURL     string `json:"html_url"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	// Compare versions
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	if !isNewerVersion(latestVersion, currentVersion) {
		return nil, nil // No update available
	}

	// Find the right asset for this platform
	assetName := um.getAssetName()
	var downloadURL string
	var size int64

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			size = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no asset found for platform: %s", assetName)
	}

	publishedAt, _ := time.Parse(time.RFC3339, release.PublishedAt)

	update := &UpdateInfo{
		Version:     latestVersion,
		ReleaseURL:  release.HTMLURL,
		DownloadURL: downloadURL,
		ReleaseDate: publishedAt,
		Changelog:   release.Body,
		Size:        size,
	}

	um.latestUpdate = update
	um.updatePending = true

	return update, nil
}

// DownloadUpdate downloads the update to a temp location.
func (um *UpdateManager) DownloadUpdate(update *UpdateInfo, progressFn func(downloaded, total int64)) (string, error) {
	// Create temp file
	tempPath := filepath.Join(um.storage.CacheDir(), "arngit-update.exe.tmp")

	resp, err := http.Get(update.DownloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Download with progress
	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			file.Write(buf[:n])
			downloaded += int64(n)
			if progressFn != nil {
				progressFn(downloaded, update.Size)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tempPath)
			return "", err
		}
	}

	return tempPath, nil
}

// ApplyUpdate applies a downloaded update.
func (um *UpdateManager) ApplyUpdate(downloadPath string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Create backup
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Move new binary
	if err := os.Rename(downloadPath, execPath); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to apply update: %w", err)
	}

	um.updatePending = false
	return nil
}

// Rollback reverts to the previous version.
func (um *UpdateManager) Rollback() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	backupPath := execPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup available for rollback")
	}

	// Remove current and restore backup
	if err := os.Remove(execPath); err != nil {
		return err
	}

	return os.Rename(backupPath, execPath)
}

// HasPendingUpdate returns true if an update is available.
func (um *UpdateManager) HasPendingUpdate() bool {
	return um.updatePending
}

// LatestUpdate returns the latest update info.
func (um *UpdateManager) LatestUpdate() *UpdateInfo {
	return um.latestUpdate
}

// LastCheckTime returns when updates were last checked.
func (um *UpdateManager) LastCheckTime() time.Time {
	return um.lastCheck
}

// getAssetName returns the expected asset name for this platform.
func (um *UpdateManager) getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	ext := ""
	if os == "windows" {
		ext = ".exe"
	}

	return fmt.Sprintf("arngit-%s-%s%s", os, arch, ext)
}

// isNewerVersion compares semantic versions.
func isNewerVersion(latest, current string) bool {
	// Simple comparison - should use semver in production
	if current == "dev" || current == "" {
		return true
	}

	latestParts := strings.Split(strings.TrimPrefix(latest, "v"), ".")
	currentParts := strings.Split(strings.TrimPrefix(current, "v"), ".")

	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		var l, c int
		fmt.Sscanf(latestParts[i], "%d", &l)
		fmt.Sscanf(currentParts[i], "%d", &c)

		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}

	return len(latestParts) > len(currentParts)
}
