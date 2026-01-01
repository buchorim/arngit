// Package status provides repository status dashboard.
package status

import (
	"fmt"
	"strings"
	"time"

	"github.com/arinara/arngit/internal/git"
)

// RepoStatus contains repository status information.
type RepoStatus struct {
	RepoPath string
	RepoName string
	Branch   string
	Remote   string
	Account  string

	// Changes
	Modified  int
	Added     int
	Deleted   int
	Untracked int

	// Commits
	Ahead  int
	Behind int

	// State
	HasRemote bool
	IsClean   bool
	LastFetch time.Time
}

// GetStatus returns the current repository status.
func GetStatus(g *git.Git) (*RepoStatus, error) {
	if !g.IsRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	status := &RepoStatus{}

	// Get repo info
	status.RepoPath, _ = g.GetRepoRoot()
	status.RepoName, _ = g.GetRepoName()
	status.Branch, _ = g.GetCurrentBranch()
	status.Remote = "origin"

	// Check remote
	_, err := g.GetRemoteURL("origin")
	status.HasRemote = err == nil

	// Parse git status
	statusOutput, _ := g.Status()
	lines := strings.Split(statusOutput, "\n")

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		staged := line[0]
		unstaged := line[1]

		// Count changes
		switch {
		case staged == '?' && unstaged == '?':
			status.Untracked++
		case staged == 'A' || unstaged == 'A':
			status.Added++
		case staged == 'D' || unstaged == 'D':
			status.Deleted++
		case staged == 'M' || unstaged == 'M':
			status.Modified++
		}
	}

	status.IsClean = status.Modified == 0 && status.Added == 0 &&
		status.Deleted == 0 && status.Untracked == 0

	// Get ahead/behind counts
	if status.HasRemote {
		status.Ahead, _ = g.GetPendingCommitCount(status.Remote, status.Branch)
		// For behind, we would need to fetch first
	}

	return status, nil
}

// FormatStatus returns a formatted status string.
func FormatStatus(s *RepoStatus, account string) string {
	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Repository: %s\n", s.RepoName))
	b.WriteString(fmt.Sprintf("  Branch:     %s\n", s.Branch))
	if account != "" {
		b.WriteString(fmt.Sprintf("  Account:    %s\n", account))
	}
	b.WriteString("\n")

	// Changes
	b.WriteString("  Changes:\n")
	if s.IsClean {
		b.WriteString("    ✓ Working tree clean\n")
	} else {
		if s.Modified > 0 {
			b.WriteString(fmt.Sprintf("    ~ Modified:  %d file(s)\n", s.Modified))
		}
		if s.Added > 0 {
			b.WriteString(fmt.Sprintf("    + Added:     %d file(s)\n", s.Added))
		}
		if s.Deleted > 0 {
			b.WriteString(fmt.Sprintf("    - Deleted:   %d file(s)\n", s.Deleted))
		}
		if s.Untracked > 0 {
			b.WriteString(fmt.Sprintf("    ? Untracked: %d file(s)\n", s.Untracked))
		}
	}
	b.WriteString("\n")

	// Commits
	b.WriteString("  Commits:\n")
	if !s.HasRemote {
		b.WriteString("    ⚠ No remote configured\n")
	} else if s.Ahead == 0 && s.Behind == 0 {
		b.WriteString("    ✓ Up to date with origin\n")
	} else {
		if s.Ahead > 0 {
			b.WriteString(fmt.Sprintf("    ↑ %d commit(s) ahead of origin/%s\n", s.Ahead, s.Branch))
		}
		if s.Behind > 0 {
			b.WriteString(fmt.Sprintf("    ↓ %d commit(s) behind origin/%s\n", s.Behind, s.Branch))
		}
	}
	b.WriteString("\n")

	return b.String()
}

// FormatCompact returns a one-line status summary.
func FormatCompact(s *RepoStatus) string {
	var parts []string

	if !s.IsClean {
		changes := s.Modified + s.Added + s.Deleted
		parts = append(parts, fmt.Sprintf("%d changes", changes))
	}

	if s.Ahead > 0 {
		parts = append(parts, fmt.Sprintf("↑%d", s.Ahead))
	}

	if s.Behind > 0 {
		parts = append(parts, fmt.Sprintf("↓%d", s.Behind))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("%s (%s) - clean", s.RepoName, s.Branch)
	}

	return fmt.Sprintf("%s (%s) - %s", s.RepoName, s.Branch, strings.Join(parts, ", "))
}
