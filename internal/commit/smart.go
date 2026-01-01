// Package commit provides smart commit message generation.
package commit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arinara/arngit/internal/git"
)

// ChangeType represents the type of file change.
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
	ChangeRenamed  ChangeType = "renamed"
)

// FileChange represents a changed file.
type FileChange struct {
	Path       string
	ChangeType ChangeType
	OldPath    string // For renames
}

// GetChanges returns the list of staged changes.
func GetChanges(g *git.Git) ([]FileChange, error) {
	status, err := g.Status()
	if err != nil {
		return nil, err
	}

	var changes []FileChange
	lines := strings.Split(status, "\n")

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		// Parse git status --short format
		// XY PATH
		// X = staged status, Y = unstaged status
		staged := line[0]
		path := strings.TrimSpace(line[3:])

		var changeType ChangeType
		switch staged {
		case 'A':
			changeType = ChangeAdded
		case 'M':
			changeType = ChangeModified
		case 'D':
			changeType = ChangeDeleted
		case 'R':
			changeType = ChangeRenamed
		case '?':
			continue // Untracked
		default:
			if staged == ' ' {
				continue // Not staged
			}
			changeType = ChangeModified
		}

		change := FileChange{
			Path:       path,
			ChangeType: changeType,
		}

		// Handle renamed files (format: "R  old -> new")
		if changeType == ChangeRenamed && strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			if len(parts) == 2 {
				change.OldPath = parts[0]
				change.Path = parts[1]
			}
		}

		changes = append(changes, change)
	}

	return changes, nil
}

// GenerateMessage creates a commit message from changes.
func GenerateMessage(changes []FileChange) string {
	if len(changes) == 0 {
		return "chore: update files"
	}

	// Analyze changes
	var added, modified, deleted []string
	for _, c := range changes {
		switch c.ChangeType {
		case ChangeAdded:
			added = append(added, filepath.Base(c.Path))
		case ChangeModified:
			modified = append(modified, filepath.Base(c.Path))
		case ChangeDeleted:
			deleted = append(deleted, filepath.Base(c.Path))
		case ChangeRenamed:
			modified = append(modified, filepath.Base(c.Path))
		}
	}

	// Generate title based on primary action
	var title string
	var commitType string

	// Determine commit type from file patterns
	hasGo := false
	hasTest := false
	hasConfig := false
	hasDocs := false

	for _, c := range changes {
		path := strings.ToLower(c.Path)
		ext := strings.ToLower(filepath.Ext(c.Path))

		if ext == ".go" {
			hasGo = true
		}
		if strings.Contains(path, "test") || strings.Contains(path, "_test") {
			hasTest = true
		}
		if ext == ".json" || ext == ".yaml" || ext == ".yml" || ext == ".toml" || ext == ".ini" {
			hasConfig = true
		}
		if ext == ".md" || strings.Contains(path, "doc") || strings.Contains(path, "readme") {
			hasDocs = true
		}
	}

	// Determine commit type
	if hasTest && !hasGo {
		commitType = "test"
	} else if hasDocs && len(changes) == 1 {
		commitType = "docs"
	} else if hasConfig && len(changes) == 1 {
		commitType = "chore"
	} else if len(deleted) > 0 && len(added) == 0 && len(modified) == 0 {
		commitType = "chore"
	} else if len(added) > 0 {
		commitType = "feat"
	} else {
		commitType = "fix"
	}

	// Generate title
	if len(changes) == 1 {
		c := changes[0]
		base := filepath.Base(c.Path)
		switch c.ChangeType {
		case ChangeAdded:
			title = fmt.Sprintf("%s: add %s", commitType, base)
		case ChangeModified:
			title = fmt.Sprintf("%s: update %s", commitType, base)
		case ChangeDeleted:
			title = fmt.Sprintf("%s: remove %s", commitType, base)
		case ChangeRenamed:
			title = fmt.Sprintf("%s: rename %s to %s", commitType, filepath.Base(c.OldPath), base)
		}
	} else {
		// Multiple files
		if len(added) > 0 && len(modified) == 0 && len(deleted) == 0 {
			title = fmt.Sprintf("%s: add %d new files", commitType, len(added))
		} else if len(modified) > 0 && len(added) == 0 && len(deleted) == 0 {
			title = fmt.Sprintf("%s: update %d files", commitType, len(modified))
		} else if len(deleted) > 0 && len(added) == 0 && len(modified) == 0 {
			title = fmt.Sprintf("%s: remove %d files", commitType, len(deleted))
		} else {
			title = fmt.Sprintf("%s: update %d files", commitType, len(changes))
		}
	}

	// Build body
	var body strings.Builder
	body.WriteString("\nChanges:\n")

	if len(added) > 0 {
		for _, f := range added {
			body.WriteString(fmt.Sprintf("  + %s\n", f))
		}
	}
	if len(modified) > 0 {
		for _, f := range modified {
			body.WriteString(fmt.Sprintf("  ~ %s\n", f))
		}
	}
	if len(deleted) > 0 {
		for _, f := range deleted {
			body.WriteString(fmt.Sprintf("  - %s\n", f))
		}
	}

	return title + body.String()
}

// GetShortMessage returns just the title line.
func GetShortMessage(changes []FileChange) string {
	full := GenerateMessage(changes)
	lines := strings.Split(full, "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return "chore: update"
}
