package ui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arfrfrr/arngit/internal/core"
	"github.com/arfrfrr/arngit/internal/git"
)

// Greetings for the dashboard
var greetings = []string{
	"Wanna Grab Some Coffee?",
	"Ready to Ship Some Code?",
	"Let's Build Something Great!",
	"Time to Push Some Commits!",
	"Ready to Code?",
	"Let's Get Things Done!",
	"What Are We Building Today?",
	"Ready for Action!",
	"Let's Make Magic Happen!",
	"Your Repos Await!",
}

// ShowDashboard displays the main dashboard when arngit is run without args.
func ShowDashboard(engine *core.Engine, ui *Renderer) error {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Logo
	ui.Logo()

	// Greeting with current user
	acc := engine.Accounts().Current()
	username := "Guest"
	if acc != nil {
		username = "@" + acc.Username
	}

	rand.Seed(time.Now().UnixNano())
	greeting := greetings[rand.Intn(len(greetings))]

	fmt.Printf("\n  %s %s\n", ui.color(BrightMagenta+Bold, username), ui.color(Dim, greeting))

	// Current time
	now := time.Now()
	timeStr := now.Format("Monday, 02 Jan 2006 15:04:05")
	fmt.Printf("  %s %s\n\n", ui.color(Dim, "🕐"), ui.color(Dim, timeStr))

	// Account status box
	if acc != nil {
		ui.Box("Active Account", fmt.Sprintf("%s <%s>", acc.Username, acc.Email), Cyan)
	} else {
		ui.Box("No Account", "Run 'arngit account add <name>' to add an account", Yellow)
	}

	fmt.Println()

	// Repository status (if in a git repo)
	gitSvc := git.NewService(engine)
	if gitSvc.IsRepo() {
		info := getDashboardRepoInfo(gitSvc)
		if info != nil {
			// Main repo info
			statusLines := []string{
				fmt.Sprintf("Branch:  %s", info.branch),
				fmt.Sprintf("Remote:  %s", info.remote),
				fmt.Sprintf("Status:  %s", info.status),
			}
			ui.Box("Repository", strings.Join(statusLines, "\n"), Green)
			fmt.Println()

			// Recent commits (compact)
			if len(info.recentCommits) > 0 {
				ui.Title("Recent Commits")
				for _, commit := range info.recentCommits {
					fmt.Printf("  %s %s %s\n",
						ui.color(Yellow, commit.hash),
						ui.color(White, commit.message),
						ui.color(Dim, commit.time))
				}
				fmt.Println()
			}
		}
	}

	// Quick commands
	ui.Title("Quick Commands")
	commands := []struct {
		cmd, desc string
	}{
		{"push", "Push commits to remote"},
		{"pull", "Pull latest changes"},
		{"commit -m \"msg\"", "Create a new commit"},
		{"status", "Show repository status"},
		{"account list", "List GitHub accounts"},
		{"help", "Show all commands"},
	}

	for _, c := range commands {
		fmt.Printf("  %s %-22s %s\n",
			ui.color(BrightCyan, "→"),
			ui.color(Cyan, c.cmd),
			ui.color(Dim, c.desc))
	}

	fmt.Println()

	// Version info
	version := engine.Version()
	if version == "" || version == "dev" {
		version = "development"
	}
	fmt.Printf("  %s\n\n", ui.color(Dim, "ArnGit "+version))

	return nil
}

// commitInfo holds compact commit info.
type commitInfo struct {
	hash    string
	message string
	time    string
}

// dashboardRepoInfo holds basic git repository information.
type dashboardRepoInfo struct {
	branch        string
	remote        string
	status        string
	recentCommits []commitInfo
}

// getDashboardRepoInfo returns info about the current git repo using git.Service.
func getDashboardRepoInfo(svc *git.Service) *dashboardRepoInfo {
	info := &dashboardRepoInfo{
		branch: "unknown",
		remote: "none",
		status: "clean",
	}

	// Get current branch
	if branch, err := svc.CurrentBranch(); err == nil {
		info.branch = branch
	}

	// Get remote URL
	if url, err := svc.GetRemoteURL("origin"); err == nil {
		remote := url
		// Shorten URL
		if strings.Contains(remote, "github.com") {
			parts := strings.Split(remote, "github.com")
			if len(parts) > 1 {
				remote = "github.com" + strings.TrimSuffix(parts[1], ".git")
			}
		}
		info.remote = remote
	}

	// Get status
	if repoStatus, err := svc.Status(); err == nil {
		if !repoStatus.HasChanges() {
			info.status = "clean"
		} else {
			modified := len(repoStatus.Modified) + len(repoStatus.Staged)
			untracked := len(repoStatus.Untracked)
			parts := []string{}
			if modified > 0 {
				parts = append(parts, fmt.Sprintf("%d modified", modified))
			}
			if untracked > 0 {
				parts = append(parts, fmt.Sprintf("%d untracked", untracked))
			}
			info.status = strings.Join(parts, ", ")
		}
	}

	// Get recent commits (last 3)
	if commits, err := svc.Log(3, false); err == nil {
		for _, c := range commits {
			relTime := time.Since(c.Date)
			timeStr := formatRelativeTime(relTime)
			info.recentCommits = append(info.recentCommits, commitInfo{
				hash:    c.Hash[:7],
				message: truncate(c.Message, 40),
				time:    timeStr,
			})
		}
	}

	return info
}

// formatRelativeTime formats a duration as a human-readable relative time.
func formatRelativeTime(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// truncate truncates a string to maxLen with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// findGitDir finds the .git directory.
func findGitDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		gitPath := filepath.Join(cwd, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return gitPath
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			return ""
		}
		cwd = parent
	}
}
