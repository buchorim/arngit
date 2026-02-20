// Package automation - Changelog generation from conventional commits.
package automation

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ChangelogCommit represents a parsed git commit.
type ChangelogCommit struct {
	Hash      string
	ShortHash string
	Author    string
	Date      time.Time
	Message   string
	Type      string // feat, fix, chore, etc.
	Scope     string
	Breaking  bool
}

// ChangelogEntry represents a changelog entry.
type ChangelogEntry struct {
	Version  string
	Date     string
	Features []string
	Fixes    []string
	Breaking []string
	Other    []string
}

// ParseCommitType parses conventional commit format.
func ParseCommitType(message string) (commitType, scope, description string, breaking bool) {
	re := regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?(!)?\s*:\s*(.+)`)
	matches := re.FindStringSubmatch(message)

	if len(matches) >= 5 {
		commitType = matches[1]
		scope = matches[2]
		breaking = matches[3] == "!"
		description = matches[4]
	} else {
		commitType = "other"
		description = message
	}

	return
}

// GetCommitsSince returns commits since a tag or reference.
func GetCommitsSince(since string) ([]ChangelogCommit, error) {
	format := "%H|%h|%an|%aI|%s"
	args := []string{"log", "--format=" + format}

	if since != "" {
		args = append(args, since+"..HEAD")
	} else {
		args = append(args, "-50")
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []ChangelogCommit
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		date, _ := time.Parse(time.RFC3339, parts[3])
		commitType, scope, _, breaking := ParseCommitType(parts[4])

		commits = append(commits, ChangelogCommit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Author:    parts[2],
			Date:      date,
			Message:   parts[4],
			Type:      commitType,
			Scope:     scope,
			Breaking:  breaking,
		})
	}

	return commits, nil
}

// GenerateChangelog generates a changelog from commits.
func GenerateChangelog(commits []ChangelogCommit, version string) *ChangelogEntry {
	entry := &ChangelogEntry{
		Version: version,
		Date:    time.Now().Format("2006-01-02"),
	}

	for _, c := range commits {
		text := c.Message
		if c.Scope != "" {
			text = fmt.Sprintf("**%s**: %s", c.Scope, c.Message)
		}

		if c.Breaking {
			entry.Breaking = append(entry.Breaking, text)
		}

		switch c.Type {
		case "feat", "feature":
			entry.Features = append(entry.Features, text)
		case "fix", "bugfix":
			entry.Fixes = append(entry.Fixes, text)
		default:
			entry.Other = append(entry.Other, text)
		}
	}

	return entry
}

// FormatChangelog formats a changelog entry to markdown.
func FormatChangelog(entry *ChangelogEntry) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## [%s] - %s\n\n", entry.Version, entry.Date))

	if len(entry.Breaking) > 0 {
		sb.WriteString("### BREAKING CHANGES\n\n")
		for _, item := range entry.Breaking {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(entry.Features) > 0 {
		sb.WriteString("### Features\n\n")
		for _, item := range entry.Features {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(entry.Fixes) > 0 {
		sb.WriteString("### Bug Fixes\n\n")
		for _, item := range entry.Fixes {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(entry.Other) > 0 {
		sb.WriteString("### Other Changes\n\n")
		for _, item := range entry.Other {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GetLatestTag returns the latest git tag.
func GetLatestTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
