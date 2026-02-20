package git

import (
	"strconv"
	"strings"
	"time"
)

// RepoStatus represents the repository status.
type RepoStatus struct {
	Branch    string
	Ahead     int
	Behind    int
	Staged    []FileStatus
	Modified  []FileStatus
	Untracked []FileStatus
	Conflicts []FileStatus
}

// FileStatus represents a file's status.
type FileStatus struct {
	Path   string
	Status string
}

// HasChanges returns true if there are uncommitted changes.
func (s *RepoStatus) HasChanges() bool {
	return len(s.Staged) > 0 || len(s.Modified) > 0 || len(s.Untracked) > 0
}

// Branch represents a git branch.
type Branch struct {
	Name    string
	Current bool
	Remote  bool
	Commit  string
	Message string
}

// Remote represents a git remote.
type Remote struct {
	Name     string
	FetchURL string
	PushURL  string
}

// Commit represents a git commit.
type Commit struct {
	Hash    string
	Author  string
	Email   string
	Date    time.Time
	Message string
}

// Stash represents a stash entry.
type Stash struct {
	Index   int
	Branch  string
	Message string
}

// parseStatus parses git status --porcelain -b output.
func parseStatus(output string) *RepoStatus {
	status := &RepoStatus{
		Staged:    []FileStatus{},
		Modified:  []FileStatus{},
		Untracked: []FileStatus{},
		Conflicts: []FileStatus{},
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		// Branch line
		if strings.HasPrefix(line, "##") {
			branchInfo := strings.TrimPrefix(line, "## ")
			parts := strings.Split(branchInfo, "...")
			status.Branch = parts[0]

			if len(parts) > 1 {
				// Parse ahead/behind
				if strings.Contains(parts[1], "ahead") {
					idx := strings.Index(parts[1], "ahead ")
					if idx >= 0 {
						numStr := strings.Fields(parts[1][idx+6:])[0]
						numStr = strings.TrimSuffix(numStr, ",")
						numStr = strings.TrimSuffix(numStr, "]")
						status.Ahead, _ = strconv.Atoi(numStr)
					}
				}
				if strings.Contains(parts[1], "behind") {
					idx := strings.Index(parts[1], "behind ")
					if idx >= 0 {
						numStr := strings.Fields(parts[1][idx+7:])[0]
						numStr = strings.TrimSuffix(numStr, "]")
						status.Behind, _ = strconv.Atoi(numStr)
					}
				}
			}
			continue
		}

		// File status
		x := line[0]
		y := line[1]
		path := strings.TrimSpace(line[3:])

		// Handle renames
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			path = parts[1]
		}

		fs := FileStatus{Path: path}

		// Conflict
		if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
			fs.Status = "conflict"
			status.Conflicts = append(status.Conflicts, fs)
			continue
		}

		// Staged changes
		if x != ' ' && x != '?' {
			fs.Status = string(x)
			status.Staged = append(status.Staged, fs)
		}

		// Unstaged changes
		if y == 'M' || y == 'D' {
			fs.Status = string(y)
			status.Modified = append(status.Modified, fs)
		}

		// Untracked
		if x == '?' && y == '?' {
			fs.Status = "?"
			status.Untracked = append(status.Untracked, fs)
		}
	}

	return status
}

// parseBranches parses git branch -v output.
func parseBranches(output string) []Branch {
	var branches []Branch

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		current := line[0] == '*'
		remote := strings.HasPrefix(strings.TrimSpace(line), "remotes/")

		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "  ")
		line = strings.TrimPrefix(line, "remotes/")

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		branch := Branch{
			Name:    parts[0],
			Current: current,
			Remote:  remote,
			Commit:  parts[1],
		}

		if len(parts) > 2 {
			branch.Message = strings.Join(parts[2:], " ")
		}

		branches = append(branches, branch)
	}

	return branches
}

// parseRemotes parses git remote -v output.
func parseRemotes(output string) []Remote {
	remoteMap := make(map[string]*Remote)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		name := parts[0]
		url := parts[1]
		kind := ""
		if len(parts) > 2 {
			kind = parts[2]
		}

		if _, ok := remoteMap[name]; !ok {
			remoteMap[name] = &Remote{Name: name}
		}

		if strings.Contains(kind, "fetch") {
			remoteMap[name].FetchURL = url
		} else if strings.Contains(kind, "push") {
			remoteMap[name].PushURL = url
		} else {
			remoteMap[name].FetchURL = url
			remoteMap[name].PushURL = url
		}
	}

	var remotes []Remote
	for _, r := range remoteMap {
		remotes = append(remotes, *r)
	}
	return remotes
}

// parseLog parses git log output.
func parseLog(output string) []Commit {
	var commits []Commit

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		timestamp, _ := strconv.ParseInt(parts[3], 10, 64)

		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    time.Unix(timestamp, 0),
			Message: parts[4],
		})
	}

	return commits
}

// parseStashList parses git stash list output.
func parseStashList(output string) []Stash {
	var stashes []Stash

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		// stash@{0}: WIP on main: abc123 message
		// stash@{0}: On main: custom message
		parts := strings.SplitN(line, ": ", 3)
		if len(parts) < 2 {
			continue
		}

		stash := Stash{Index: i}

		// Parse branch
		branchPart := parts[1]
		if strings.HasPrefix(branchPart, "WIP on ") {
			stash.Branch = strings.TrimPrefix(branchPart, "WIP on ")
		} else if strings.HasPrefix(branchPart, "On ") {
			stash.Branch = strings.TrimPrefix(branchPart, "On ")
		}

		// Parse message
		if len(parts) > 2 {
			stash.Message = parts[2]
		}

		stashes = append(stashes, stash)
	}

	return stashes
}
