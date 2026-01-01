// Package git provides wrapper functions for git commands.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Git wraps git command execution.
type Git struct {
	WorkDir string
}

// New creates a new Git instance for the given directory.
func New(workDir string) *Git {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	return &Git{WorkDir: workDir}
}

// run executes a git command and returns the output.
func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.WorkDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// IsRepo checks if current directory is a git repository.
func (g *Git) IsRepo() bool {
	_, err := g.run("rev-parse", "--git-dir")
	return err == nil
}

// Init initializes a new git repository.
func (g *Git) Init() error {
	_, err := g.run("init")
	return err
}

// GetRepoRoot returns the root directory of the repository.
func (g *Git) GetRepoRoot() (string, error) {
	return g.run("rev-parse", "--show-toplevel")
}

// GetCurrentBranch returns the current branch name.
func (g *Git) GetCurrentBranch() (string, error) {
	return g.run("rev-parse", "--abbrev-ref", "HEAD")
}

// GetRemoteURL returns the URL of the remote.
func (g *Git) GetRemoteURL(remote string) (string, error) {
	return g.run("remote", "get-url", remote)
}

// SetRemoteURL sets the URL for a remote.
func (g *Git) SetRemoteURL(remote, url string) error {
	// Check if remote exists
	_, err := g.run("remote", "get-url", remote)
	if err != nil {
		// Remote doesn't exist, add it
		_, err = g.run("remote", "add", remote, url)
	} else {
		// Remote exists, update it
		_, err = g.run("remote", "set-url", remote, url)
	}
	return err
}

// SetConfig sets a git config value.
func (g *Git) SetConfig(key, value string) error {
	_, err := g.run("config", key, value)
	return err
}

// GetConfig gets a git config value.
func (g *Git) GetConfig(key string) (string, error) {
	return g.run("config", "--get", key)
}

// Add stages files for commit.
func (g *Git) Add(paths ...string) error {
	args := append([]string{"add"}, paths...)
	_, err := g.run(args...)
	return err
}

// Commit creates a new commit.
func (g *Git) Commit(message string) error {
	_, err := g.run("commit", "-m", message)
	return err
}

// Push pushes commits to remote.
func (g *Git) Push(remote, branch string) error {
	_, err := g.run("push", remote, branch)
	return err
}

// PushWithCredentials pushes using PAT authentication.
func (g *Git) PushWithCredentials(remote, branch, username, pat string) error {
	// Get remote URL
	url, err := g.GetRemoteURL(remote)
	if err != nil {
		return err
	}

	// Modify URL to include credentials for HTTPS
	if strings.HasPrefix(url, "https://") {
		// Extract host and path
		urlPart := strings.TrimPrefix(url, "https://")
		// Remove any existing credentials
		if idx := strings.Index(urlPart, "@"); idx != -1 {
			urlPart = urlPart[idx+1:]
		}
		// Create URL with credentials
		authURL := fmt.Sprintf("https://%s:%s@%s", username, pat, urlPart)
		_, err = g.run("push", authURL, branch)
		return err
	}

	// For SSH, just push normally
	_, err = g.run("push", remote, branch)
	return err
}

// Pull fetches and merges from remote.
func (g *Git) Pull(remote, branch string) error {
	_, err := g.run("pull", remote, branch)
	return err
}

// Status returns the repository status.
func (g *Git) Status() (string, error) {
	return g.run("status", "--short")
}

// GetPendingCommitCount returns the number of commits ahead of remote.
func (g *Git) GetPendingCommitCount(remote, branch string) (int, error) {
	refSpec := fmt.Sprintf("%s/%s..HEAD", remote, branch)
	output, err := g.run("rev-list", "--count", refSpec)
	if err != nil {
		// If remote branch doesn't exist, count all commits
		output, err = g.run("rev-list", "--count", "HEAD")
		if err != nil {
			return 0, err
		}
	}
	return strconv.Atoi(output)
}

// GetUnpushedChangesSize returns approximate size of unpushed changes in bytes.
func (g *Git) GetUnpushedChangesSize(remote, branch string) (int64, error) {
	// Get list of changed files
	refSpec := fmt.Sprintf("%s/%s..HEAD", remote, branch)
	output, err := g.run("diff", "--stat", refSpec)
	if err != nil {
		return 0, err
	}

	// Parse the last line for total changes
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return 0, nil
	}

	// Simple estimation: count characters changed
	var total int64
	for _, line := range lines {
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				// Count + and - characters
				changes := strings.TrimSpace(parts[1])
				total += int64(len(changes))
			}
		}
	}

	return total * 100, nil // Rough estimate in bytes
}

// GetRepoName extracts repository name from remote URL.
func (g *Git) GetRepoName() (string, error) {
	url, err := g.GetRemoteURL("origin")
	if err != nil {
		// Fallback to directory name
		root, err := g.GetRepoRoot()
		if err != nil {
			return "", err
		}
		return filepath.Base(root), nil
	}

	// Extract name from URL
	// https://github.com/user/repo.git -> repo
	// git@github.com:user/repo.git -> repo
	url = strings.TrimSuffix(url, ".git")
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("cannot determine repo name")
}

// HasUncommittedChanges checks if there are uncommitted changes.
func (g *Git) HasUncommittedChanges() (bool, error) {
	status, err := g.Status()
	if err != nil {
		return false, err
	}
	return status != "", nil
}

// Fetch fetches from remote.
func (g *Git) Fetch(remote string) error {
	_, err := g.run("fetch", remote)
	return err
}
