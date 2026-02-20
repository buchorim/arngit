// Package git provides git command execution functionality.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arfrfrr/arngit/internal/core"
)

// Service provides git operations.
type Service struct {
	engine *core.Engine
}

// NewService creates a new git service.
func NewService(engine *core.Engine) *Service {
	return &Service{engine: engine}
}

// IsInstalled checks if git is installed.
func (s *Service) IsInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// Version returns the git version.
func (s *Service) Version() (string, error) {
	out, err := s.run("version")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(out), "git version "), nil
}

// IsRepo checks if the current directory is a git repository.
func (s *Service) IsRepo() bool {
	_, err := s.run("rev-parse", "--git-dir")
	return err == nil
}

// Init initializes a new git repository.
func (s *Service) Init(bare bool) error {
	args := []string{"init"}
	if bare {
		args = append(args, "--bare")
	}
	_, err := s.run(args...)
	return err
}

// Clone clones a repository.
func (s *Service) Clone(url, dest string, depth int) error {
	args := []string{"clone", url}
	if dest != "" {
		args = append(args, dest)
	}
	if depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", depth))
	}
	_, err := s.run(args...)
	return err
}

// Status returns the repository status.
func (s *Service) Status() (*RepoStatus, error) {
	out, err := s.run("status", "--porcelain", "-b")
	if err != nil {
		return nil, err
	}
	return parseStatus(out), nil
}

// Add stages files.
func (s *Service) Add(files ...string) error {
	if len(files) == 0 {
		files = []string{"-A"}
	}
	args := append([]string{"add"}, files...)
	_, err := s.run(args...)
	return err
}

// Commit creates a commit.
func (s *Service) Commit(message string, allowEmpty bool) error {
	args := []string{"commit", "-m", message}
	if allowEmpty {
		args = append(args, "--allow-empty")
	}
	_, err := s.run(args...)
	return err
}

// Push pushes commits to remote.
func (s *Service) Push(remote, branch string, force bool, setUpstream bool) error {
	args := []string{"push"}
	if force {
		args = append(args, "--force")
	}
	if setUpstream {
		args = append(args, "-u")
	}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}
	_, err := s.runWithAuth(args...)
	return err
}

// Pull pulls changes from remote.
func (s *Service) Pull(remote, branch string, rebase bool) error {
	args := []string{"pull"}
	if rebase {
		args = append(args, "--rebase")
	}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}
	_, err := s.runWithAuth(args...)
	return err
}

// Fetch fetches from remote.
func (s *Service) Fetch(remote string, prune bool) error {
	args := []string{"fetch"}
	if prune {
		args = append(args, "--prune")
	}
	if remote != "" {
		args = append(args, remote)
	}
	_, err := s.runWithAuth(args...)
	return err
}

// CurrentBranch returns the current branch name.
func (s *Service) CurrentBranch() (string, error) {
	out, err := s.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Branches returns all branches.
func (s *Service) Branches(all bool) ([]Branch, error) {
	args := []string{"branch", "-v"}
	if all {
		args = append(args, "-a")
	}
	out, err := s.run(args...)
	if err != nil {
		return nil, err
	}
	return parseBranches(out), nil
}

// CreateBranch creates a new branch.
func (s *Service) CreateBranch(name string, checkout bool) error {
	if checkout {
		_, err := s.run("checkout", "-b", name)
		return err
	}
	_, err := s.run("branch", name)
	return err
}

// SwitchBranch switches to a branch.
func (s *Service) SwitchBranch(name string) error {
	_, err := s.run("checkout", name)
	return err
}

// DeleteBranch deletes a branch.
func (s *Service) DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := s.run("branch", flag, name)
	return err
}

// Remotes returns all remotes.
func (s *Service) Remotes() ([]Remote, error) {
	out, err := s.run("remote", "-v")
	if err != nil {
		return nil, err
	}
	return parseRemotes(out), nil
}

// AddRemote adds a remote.
func (s *Service) AddRemote(name, url string) error {
	_, err := s.run("remote", "add", name, url)
	return err
}

// RemoveRemote removes a remote.
func (s *Service) RemoveRemote(name string) error {
	_, err := s.run("remote", "remove", name)
	return err
}

// GetRemoteURL returns the URL of a remote.
func (s *Service) GetRemoteURL(name string) (string, error) {
	out, err := s.run("remote", "get-url", name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Diff returns the diff.
func (s *Service) Diff(staged bool, file string) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}
	if file != "" {
		args = append(args, "--", file)
	}
	return s.run(args...)
}

// Log returns commit history.
func (s *Service) Log(n int, oneline bool) ([]Commit, error) {
	args := []string{"log", fmt.Sprintf("-n%d", n), "--pretty=format:%H|%an|%ae|%at|%s"}
	out, err := s.run(args...)
	if err != nil {
		return nil, err
	}
	return parseLog(out), nil
}

// Stash stashes changes.
func (s *Service) Stash(message string) error {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	_, err := s.run(args...)
	return err
}

// StashList lists stashes.
func (s *Service) StashList() ([]Stash, error) {
	out, err := s.run("stash", "list")
	if err != nil {
		return nil, err
	}
	return parseStashList(out), nil
}

// StashPop pops a stash.
func (s *Service) StashPop(index int) error {
	args := []string{"stash", "pop"}
	if index >= 0 {
		args = append(args, fmt.Sprintf("stash@{%d}", index))
	}
	_, err := s.run(args...)
	return err
}

// Tags returns all tags.
func (s *Service) Tags() ([]string, error) {
	out, err := s.run("tag", "-l")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

// CreateTag creates a tag.
func (s *Service) CreateTag(name, message string) error {
	args := []string{"tag"}
	if message != "" {
		args = append(args, "-a", name, "-m", message)
	} else {
		args = append(args, name)
	}
	_, err := s.run(args...)
	return err
}

// DeleteTag deletes a tag.
func (s *Service) DeleteTag(name string) error {
	_, err := s.run("tag", "-d", name)
	return err
}

// LogGraph returns commit history with ASCII graph.
func (s *Service) LogGraph(n int) (string, error) {
	out, err := s.run("log", fmt.Sprintf("-n%d", n),
		"--graph", "--oneline", "--decorate", "--all")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// PendingCommitCount returns the number of commits not yet pushed.
func (s *Service) PendingCommitCount(remote, branch string) (int, error) {
	ref := fmt.Sprintf("%s/%s..HEAD", remote, branch)
	out, err := s.run("rev-list", "--count", ref)
	if err != nil {
		return 0, err
	}
	count := 0
	fmt.Sscanf(strings.TrimSpace(out), "%d", &count)
	return count, nil
}

// UnpushedChangesSize returns the total size of unpushed changes in bytes.
func (s *Service) UnpushedChangesSize(remote, branch string) (int64, error) {
	ref := fmt.Sprintf("%s/%s..HEAD", remote, branch)
	out, err := s.run("diff", "--stat", ref)
	if err != nil {
		return 0, err
	}
	// Count bytes in the diff output as approximation
	return int64(len(out)), nil
}

// RepoName returns the repository name from the remote URL.
func (s *Service) RepoName() (string, error) {
	url, err := s.GetRemoteURL("origin")
	if err != nil {
		return "", err
	}
	// Extract repo name from URL
	url = strings.TrimSuffix(url, ".git")
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}
	return "", fmt.Errorf("could not parse repo name")
}

// run executes a git command.
func (s *Service) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("%s", strings.TrimSpace(errMsg))
	}

	return stdout.String(), nil
}

// runWithAuth executes a git command with authentication.
func (s *Service) runWithAuth(args ...string) (string, error) {
	// Get current account PAT
	acc := s.engine.Accounts().Current()
	if acc == nil {
		return s.run(args...)
	}

	pat, err := s.engine.Accounts().GetPAT(acc.Name)
	if err != nil {
		return s.run(args...)
	}

	// Set GIT_ASKPASS to inject credentials
	cmd := exec.Command("git", args...)

	// Create a helper script for credential injection
	helperContent := fmt.Sprintf("@echo off\necho %s", pat)
	helperPath := filepath.Join(os.TempDir(), "arngit-helper.bat")
	os.WriteFile(helperPath, []byte(helperContent), 0600)
	defer os.Remove(helperPath)

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", helperPath),
		"GIT_TERMINAL_PROMPT=0",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("%s", strings.TrimSpace(errMsg))
	}

	return stdout.String(), nil
}
