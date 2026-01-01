// Package updater provides self-update functionality.
package updater

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

const (
	repoOwner = "buchorim"
	repoName  = "arngit"
)

// Release represents a GitHub release.
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateInfo contains update information.
type UpdateInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	ReleaseNotes    string
	DownloadURL     string
	AssetName       string
	AssetSize       int64
}

// CheckForUpdate checks if a new version is available.
func CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "arngit-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to check for updates: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		ReleaseNotes:   release.Body,
	}

	// Compare versions
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(release.TagName, "v")
	info.UpdateAvailable = latest != current && latest > current

	// Find matching asset
	assetName := getAssetName()
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			info.DownloadURL = asset.BrowserDownloadURL
			info.AssetName = asset.Name
			info.AssetSize = asset.Size
			break
		}
	}

	return info, nil
}

// getAssetName returns the expected asset name for current platform.
func getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "windows":
		return "arngit.exe"
	case "darwin":
		if arch == "arm64" {
			return "arngit-darwin-arm64"
		}
		return "arngit-darwin-amd64"
	case "linux":
		if arch == "arm64" {
			return "arngit-linux-arm64"
		}
		return "arngit-linux-amd64"
	}

	return "arngit"
}

// DownloadUpdate downloads the update to a temporary file.
func DownloadUpdate(downloadURL string, progressFn func(downloaded, total int64)) (string, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "arngit-update-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Download with progress
	total := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			tmpFile.Write(buf[:n])
			downloaded += int64(n)
			if progressFn != nil {
				progressFn(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return tmpFile.Name(), nil
}

// ApplyUpdate replaces the current executable with the new one.
func ApplyUpdate(downloadPath string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}

	// Backup current executable
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %w", err)
	}

	// Move new executable to place
	if err := copyFile(downloadPath, execPath); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Make executable
	if runtime.GOOS != "windows" {
		os.Chmod(execPath, 0755)
	}

	// Clean up
	os.Remove(backupPath)
	os.Remove(downloadPath)

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// GetCurrentExecutablePath returns the path to the current executable.
func GetCurrentExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(execPath)
}
