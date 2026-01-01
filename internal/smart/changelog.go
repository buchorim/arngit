// Package smart provides smart features for git operations.
package smart

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Commit represents a parsed git commit.
type Commit struct {
	Hash      string
	ShortHash string
	Author    string
	Date      time.Time
	Message   string
	Type      string // feat, fix, chore, etc.
	Scope     string
	Breaking  bool
}

// ParseCommitType parses conventional commit format.
func ParseCommitType(message string) (commitType, scope, description string, breaking bool) {
	// Pattern: type(scope)!: description
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
func GetCommitsSince(since string) ([]Commit, error) {
	format := "%H|%h|%an|%aI|%s"
	args := []string{"log", "--format=" + format}

	if since != "" {
		args = append(args, since+"..HEAD")
	} else {
		args = append(args, "-50") // Last 50 commits if no reference
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []Commit
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

		commits = append(commits, Commit{
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

// ChangelogEntry represents a changelog entry.
type ChangelogEntry struct {
	Version  string
	Date     string
	Features []string
	Fixes    []string
	Breaking []string
	Other    []string
}

// GenerateChangelog generates a changelog from commits.
func GenerateChangelog(commits []Commit, version string) *ChangelogEntry {
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
		sb.WriteString("### ⚠ BREAKING CHANGES\n\n")
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

// GetAllTags returns all tags sorted by version.
func GetAllTags() ([]string, error) {
	cmd := exec.Command("git", "tag", "-l", "--sort=-v:refname")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	tags := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, t := range tags {
		if t != "" {
			result = append(result, t)
		}
	}

	return result, nil
}

// SuggestNextVersion suggests next version based on commits.
func SuggestNextVersion(commits []Commit, currentVersion string) string {
	// Parse current version (v1.2.3 or 1.2.3)
	version := strings.TrimPrefix(currentVersion, "v")
	parts := strings.Split(version, ".")

	major, minor, patch := 0, 0, 0
	if len(parts) >= 1 {
		fmt.Sscanf(parts[0], "%d", &major)
	}
	if len(parts) >= 2 {
		fmt.Sscanf(parts[1], "%d", &minor)
	}
	if len(parts) >= 3 {
		fmt.Sscanf(parts[2], "%d", &patch)
	}

	// Determine bump type based on commits
	hasBreaking := false
	hasFeature := false

	for _, c := range commits {
		if c.Breaking {
			hasBreaking = true
		}
		if c.Type == "feat" || c.Type == "feature" {
			hasFeature = true
		}
	}

	if hasBreaking {
		major++
		minor = 0
		patch = 0
	} else if hasFeature {
		minor++
		patch = 0
	} else {
		patch++
	}

	prefix := ""
	if strings.HasPrefix(currentVersion, "v") {
		prefix = "v"
	}

	return fmt.Sprintf("%s%d.%d.%d", prefix, major, minor, patch)
}

// ContributorStats contains contributor statistics.
type ContributorStats struct {
	Name    string
	Email   string
	Commits int
}

// GetContributorStats returns contributor statistics.
func GetContributorStats() ([]ContributorStats, error) {
	cmd := exec.Command("git", "shortlog", "-sne", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var stats []ContributorStats
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	re := regexp.MustCompile(`^\s*(\d+)\s+(.+?)\s+<(.+)>$`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 4 {
			var count int
			fmt.Sscanf(matches[1], "%d", &count)
			stats = append(stats, ContributorStats{
				Commits: count,
				Name:    matches[2],
				Email:   matches[3],
			})
		}
	}

	// Sort by commits descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Commits > stats[j].Commits
	})

	return stats, nil
}
