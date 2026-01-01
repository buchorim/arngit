// Package automation provides automation utilities for git operations.
package automation

import (
	"fmt"
	"os/exec"
	"strings"
)

// SyncResult contains the result of a sync operation.
type SyncResult struct {
	FetchOutput  string
	RebaseOutput string
	Ahead        int
	Behind       int
	Updated      bool
	Conflicts    bool
}

// Sync performs fetch + rebase from remote.
func Sync(remote, branch string) (*SyncResult, error) {
	result := &SyncResult{}

	// Fetch
	fetchCmd := exec.Command("git", "fetch", remote)
	fetchOut, err := fetchCmd.CombinedOutput()
	result.FetchOutput = string(fetchOut)
	if err != nil {
		return result, fmt.Errorf("fetch failed: %w", err)
	}

	// Check if behind
	behindCmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("HEAD..%s/%s", remote, branch))
	behindOut, _ := behindCmd.Output()
	fmt.Sscanf(strings.TrimSpace(string(behindOut)), "%d", &result.Behind)

	// Check if ahead
	aheadCmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s/%s..HEAD", remote, branch))
	aheadOut, _ := aheadCmd.Output()
	fmt.Sscanf(strings.TrimSpace(string(aheadOut)), "%d", &result.Ahead)

	// If behind, rebase
	if result.Behind > 0 {
		rebaseCmd := exec.Command("git", "rebase", fmt.Sprintf("%s/%s", remote, branch))
		rebaseOut, err := rebaseCmd.CombinedOutput()
		result.RebaseOutput = string(rebaseOut)
		if err != nil {
			// Check for conflicts
			if strings.Contains(result.RebaseOutput, "CONFLICT") {
				result.Conflicts = true
				return result, fmt.Errorf("rebase conflicts detected")
			}
			return result, fmt.Errorf("rebase failed: %w", err)
		}
		result.Updated = true
	}

	return result, nil
}

// QuickSync does a simple pull --rebase.
func QuickSync(remote, branch string) error {
	cmd := exec.Command("git", "pull", "--rebase", remote, branch)
	return cmd.Run()
}

// FetchAll fetches from all remotes.
func FetchAll() error {
	cmd := exec.Command("git", "fetch", "--all")
	return cmd.Run()
}

// GetSyncStatus returns current sync status with remote.
func GetSyncStatus(remote, branch string) (ahead, behind int, err error) {
	// Fetch first
	exec.Command("git", "fetch", remote).Run()

	// Ahead
	aheadCmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s/%s..HEAD", remote, branch))
	aheadOut, _ := aheadCmd.Output()
	fmt.Sscanf(strings.TrimSpace(string(aheadOut)), "%d", &ahead)

	// Behind
	behindCmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("HEAD..%s/%s", remote, branch))
	behindOut, _ := behindCmd.Output()
	fmt.Sscanf(strings.TrimSpace(string(behindOut)), "%d", &behind)

	return ahead, behind, nil
}
