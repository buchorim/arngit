// Package analytics provides git analytics and statistics.
package analytics

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// RepoStats contains repository statistics.
type RepoStats struct {
	TotalCommits    int
	TotalBranches   int
	TotalTags       int
	Contributors    int
	FirstCommitDate time.Time
	LastCommitDate  time.Time
	AgeInDays       int
}

// GetRepoStats returns repository statistics.
func GetRepoStats() (*RepoStats, error) {
	stats := &RepoStats{}

	// Total commits
	if out, err := exec.Command("git", "rev-list", "--count", "HEAD").Output(); err == nil {
		stats.TotalCommits, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	// Total branches
	if out, err := exec.Command("git", "branch", "-a").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if lines[0] != "" {
			stats.TotalBranches = len(lines)
		}
	}

	// Total tags
	if out, err := exec.Command("git", "tag").Output(); err == nil {
		tags := strings.TrimSpace(string(out))
		if tags != "" {
			stats.TotalTags = len(strings.Split(tags, "\n"))
		}
	}

	// Contributors
	if out, err := exec.Command("git", "shortlog", "-sn", "HEAD").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if lines[0] != "" {
			stats.Contributors = len(lines)
		}
	}

	// First commit date
	if out, err := exec.Command("git", "log", "--reverse", "--format=%aI", "-1").Output(); err == nil {
		if date, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out))); err == nil {
			stats.FirstCommitDate = date
		}
	}

	// Last commit date
	if out, err := exec.Command("git", "log", "--format=%aI", "-1").Output(); err == nil {
		if date, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out))); err == nil {
			stats.LastCommitDate = date
		}
	}

	// Age in days
	if !stats.FirstCommitDate.IsZero() {
		stats.AgeInDays = int(time.Since(stats.FirstCommitDate).Hours() / 24)
	}

	return stats, nil
}

// CommitActivity contains commit activity stats.
type CommitActivity struct {
	ByHour    [24]int
	ByWeekday [7]int
	ByAuthor  map[string]int
}

// GetCommitActivity analyzes commit patterns.
func GetCommitActivity(limit int) (*CommitActivity, error) {
	activity := &CommitActivity{
		ByAuthor: make(map[string]int),
	}

	cmd := exec.Command("git", "log", "--format=%aI|%an", fmt.Sprintf("-%d", limit))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 2 {
			continue
		}

		date, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			continue
		}

		activity.ByHour[date.Hour()]++
		activity.ByWeekday[date.Weekday()]++
		activity.ByAuthor[parts[1]]++
	}

	return activity, nil
}

// WeekdayName returns the name of a weekday.
func WeekdayName(day int) string {
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if day >= 0 && day < 7 {
		return days[day]
	}
	return ""
}

// FormatNumber formats a number with commas.
func FormatNumber(n int) string {
	str := strconv.Itoa(n)
	if len(str) <= 3 {
		return str
	}

	var result []byte
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
