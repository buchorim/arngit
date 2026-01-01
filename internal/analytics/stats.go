// Package analytics provides git analytics and statistics.
package analytics

import (
	"fmt"
	"os/exec"
	"regexp"
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
	LinesAdded      int
	LinesDeleted    int
	FilesChanged    int
}

// GetRepoStats returns repository statistics.
func GetRepoStats() (*RepoStats, error) {
	stats := &RepoStats{}

	// Total commits
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	out, _ := cmd.Output()
	stats.TotalCommits, _ = strconv.Atoi(strings.TrimSpace(string(out)))

	// Total branches
	cmd = exec.Command("git", "branch", "-a")
	out, _ = cmd.Output()
	stats.TotalBranches = len(strings.Split(strings.TrimSpace(string(out)), "\n"))

	// Total tags
	cmd = exec.Command("git", "tag")
	out, _ = cmd.Output()
	tags := strings.TrimSpace(string(out))
	if tags != "" {
		stats.TotalTags = len(strings.Split(tags, "\n"))
	}

	// Contributors
	cmd = exec.Command("git", "shortlog", "-sn", "HEAD")
	out, _ = cmd.Output()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if lines[0] != "" {
		stats.Contributors = len(lines)
	}

	// First commit date
	cmd = exec.Command("git", "log", "--reverse", "--format=%aI", "-1")
	out, _ = cmd.Output()
	if date, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out))); err == nil {
		stats.FirstCommitDate = date
	}

	// Last commit date
	cmd = exec.Command("git", "log", "--format=%aI", "-1")
	out, _ = cmd.Output()
	if date, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out))); err == nil {
		stats.LastCommitDate = date
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
	ByMonth   map[string]int
	ByAuthor  map[string]int
}

// GetCommitActivity analyzes commit patterns.
func GetCommitActivity(limit int) (*CommitActivity, error) {
	activity := &CommitActivity{
		ByMonth:  make(map[string]int),
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

		author := parts[1]

		// By hour
		activity.ByHour[date.Hour()]++

		// By weekday
		activity.ByWeekday[date.Weekday()]++

		// By month
		monthKey := date.Format("2006-01")
		activity.ByMonth[monthKey]++

		// By author
		activity.ByAuthor[author]++
	}

	return activity, nil
}

// FileStats contains file change statistics.
type FileStats struct {
	Path    string
	Changes int
	Added   int
	Deleted int
}

// GetMostChangedFiles returns files with most changes.
func GetMostChangedFiles(limit int) ([]FileStats, error) {
	cmd := exec.Command("git", "log", "--format=", "--numstat", fmt.Sprintf("-%d", limit))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	fileChanges := make(map[string]*FileStats)

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		added, _ := strconv.Atoi(parts[0])
		deleted, _ := strconv.Atoi(parts[1])
		path := parts[2]

		if _, ok := fileChanges[path]; !ok {
			fileChanges[path] = &FileStats{Path: path}
		}

		fileChanges[path].Changes++
		fileChanges[path].Added += added
		fileChanges[path].Deleted += deleted
	}

	// Convert to slice and sort
	var result []FileStats
	for _, fs := range fileChanges {
		result = append(result, *fs)
	}

	// Sort by changes descending
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Changes > result[i].Changes {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if len(result) > 10 {
		result = result[:10]
	}

	return result, nil
}

// BranchInfo contains branch information.
type BranchInfo struct {
	Name       string
	IsCurrent  bool
	LastCommit string
	Ahead      int
	Behind     int
}

// GetBranchInfo returns information about all branches.
func GetBranchInfo() ([]BranchInfo, error) {
	cmd := exec.Command("git", "branch", "-vv")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []BranchInfo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		info := BranchInfo{}

		// Check if current branch
		if strings.HasPrefix(line, "*") {
			info.IsCurrent = true
			line = strings.TrimPrefix(line, "* ")
		} else {
			line = strings.TrimSpace(line)
		}

		// Parse branch name and info
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			info.Name = parts[0]
		}
		if len(parts) >= 2 {
			info.LastCommit = parts[1]
		}

		// Parse ahead/behind
		re := regexp.MustCompile(`\[.*?(?:ahead (\d+))?(?:, )?(?:behind (\d+))?\]`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 && matches[1] != "" {
			info.Ahead, _ = strconv.Atoi(matches[1])
		}
		if len(matches) >= 3 && matches[2] != "" {
			info.Behind, _ = strconv.Atoi(matches[2])
		}

		branches = append(branches, info)
	}

	return branches, nil
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
