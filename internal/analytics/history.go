// Package analytics - Commit history visualization
package analytics

import (
	"fmt"
	"os/exec"
	"strings"
)

// HistoryEntry represents a commit in history view.
type HistoryEntry struct {
	Hash      string
	ShortHash string
	Author    string
	Date      string
	Message   string
	Branches  []string
	Tags      []string
	IsHead    bool
}

// GetHistory returns formatted commit history.
func GetHistory(limit int) ([]HistoryEntry, error) {
	format := "%H|%h|%an|%ar|%s|%D"
	cmd := exec.Command("git", "log", "--format="+format, fmt.Sprintf("-%d", limit))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var entries []HistoryEntry
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 5 {
			continue
		}

		entry := HistoryEntry{
			Hash:      parts[0],
			ShortHash: parts[1],
			Author:    parts[2],
			Date:      parts[3],
			Message:   parts[4],
		}

		// Parse refs (branches, tags)
		if len(parts) >= 6 && parts[5] != "" {
			refs := strings.Split(parts[5], ", ")
			for _, ref := range refs {
				ref = strings.TrimSpace(ref)
				if strings.HasPrefix(ref, "HEAD") {
					entry.IsHead = true
				}
				if strings.HasPrefix(ref, "tag: ") {
					entry.Tags = append(entry.Tags, strings.TrimPrefix(ref, "tag: "))
				} else if !strings.Contains(ref, "HEAD") {
					entry.Branches = append(entry.Branches, ref)
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// FormatHistoryTree formats history as ASCII tree.
func FormatHistoryTree(entries []HistoryEntry) string {
	var sb strings.Builder

	for i, e := range entries {
		// Draw tree line
		if i == 0 {
			sb.WriteString("  ◉")
		} else {
			sb.WriteString("  │\n  ○")
		}

		// Commit info
		sb.WriteString(fmt.Sprintf(" %s ", e.ShortHash))

		// Tags/branches
		if len(e.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("[%s] ", strings.Join(e.Tags, ", ")))
		}
		if e.IsHead {
			sb.WriteString("(HEAD) ")
		}

		// Message (truncate if too long)
		msg := e.Message
		if len(msg) > 50 {
			msg = msg[:47] + "..."
		}
		sb.WriteString(msg)
		sb.WriteString("\n")

		// Author and date (dimmed)
		sb.WriteString(fmt.Sprintf("  │   %s, %s\n", e.Author, e.Date))
	}

	return sb.String()
}

// GetCommitGraph returns a simple branch graph.
func GetCommitGraph(limit int) (string, error) {
	cmd := exec.Command("git", "log", "--graph", "--oneline", "--all", fmt.Sprintf("-%d", limit))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// BlameEntry represents a line in blame output.
type BlameEntry struct {
	Hash    string
	Author  string
	Date    string
	Line    int
	Content string
}

// GetBlame returns blame information for a file.
func GetBlame(filePath string, startLine, endLine int) ([]BlameEntry, error) {
	args := []string{"blame", "--line-porcelain"}
	if startLine > 0 && endLine > 0 {
		args = append(args, fmt.Sprintf("-L%d,%d", startLine, endLine))
	}
	args = append(args, filePath)

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var entries []BlameEntry
	lines := strings.Split(string(out), "\n")

	var current BlameEntry
	lineNum := startLine
	if lineNum == 0 {
		lineNum = 1
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "\t") {
			// Content line
			current.Content = strings.TrimPrefix(line, "\t")
			current.Line = lineNum
			entries = append(entries, current)
			current = BlameEntry{}
			lineNum++
		} else if len(line) >= 40 && !strings.Contains(line[:40], " ") {
			// Hash line
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				current.Hash = parts[0][:7]
			}
		} else if strings.HasPrefix(line, "author ") {
			current.Author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-time ") {
			// Could parse timestamp here
		} else if strings.HasPrefix(line, "summary ") {
			// Commit message
		}
	}

	return entries, nil
}

// FormatBlame formats blame output.
func FormatBlame(entries []BlameEntry) string {
	var sb strings.Builder

	for _, e := range entries {
		author := e.Author
		if len(author) > 12 {
			author = author[:12]
		}

		sb.WriteString(fmt.Sprintf("  %s │ %-12s │ %4d │ %s\n",
			e.Hash, author, e.Line, e.Content))
	}

	return sb.String()
}
