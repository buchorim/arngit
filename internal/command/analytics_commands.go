package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/arfrfrr/arngit/internal/analytics"
	"github.com/arfrfrr/arngit/internal/github"
	"github.com/arfrfrr/arngit/internal/ui"
)

// RegisterAnalyticsCommands registers analytics commands.
func (r *Router) RegisterAnalyticsCommands() {
	// stats command
	r.Register(&Command{
		Name:        "stats",
		Description: "Show repository statistics",
		Usage:       "arngit stats",
		Handler:     r.handleStats,
	})

	// blame command
	r.Register(&Command{
		Name:        "blame",
		Description: "Show file annotations",
		Usage:       "arngit blame <file>",
		Handler:     r.handleBlame,
	})
}

// handleStats shows repository statistics.
func (r *Router) handleStats(ctx *Context) error {
	gitSvc := r.getGitService()
	if !gitSvc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a repo")
	}

	stats, err := analytics.GetRepoStats()
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to get stats: %v", err))
		return err
	}

	ctx.UI.Title("Repository Statistics")

	// Basic stats
	rows := [][]string{
		{"Total Commits", analytics.FormatNumber(stats.TotalCommits)},
		{"Branches", analytics.FormatNumber(stats.TotalBranches)},
		{"Tags", analytics.FormatNumber(stats.TotalTags)},
		{"Contributors", analytics.FormatNumber(stats.Contributors)},
	}

	if !stats.FirstCommitDate.IsZero() {
		rows = append(rows, []string{"First Commit", stats.FirstCommitDate.Format("2006-01-02")})
		rows = append(rows, []string{"Last Commit", stats.LastCommitDate.Format("2006-01-02")})
		rows = append(rows, []string{"Age", fmt.Sprintf("%d days", stats.AgeInDays)})
	}

	for _, row := range rows {
		fmt.Printf("  %-18s %s\n",
			r.ui.Color(ui.Dim, row[0]+":"),
			r.ui.Color(ui.BrightCyan, row[1]))
	}
	fmt.Println()

	// Commit activity
	activity, err := analytics.GetCommitActivity(100)
	if err == nil && len(activity.ByAuthor) > 0 {
		ctx.UI.Title("Top Contributors")
		for author, count := range activity.ByAuthor {
			bar := strings.Repeat("█", count)
			if len(bar) > 30 {
				bar = bar[:30]
			}
			fmt.Printf("  %-20s %s %s\n",
				r.ui.Color(ui.White, author),
				r.ui.Color(ui.BrightCyan, bar),
				r.ui.Color(ui.Dim, fmt.Sprintf("(%d)", count)))
		}
		fmt.Println()
	}

	return nil
}

// handleBlame shows file annotations.
func (r *Router) handleBlame(ctx *Context) error {
	if len(ctx.Args) < 1 {
		ctx.UI.Error("File path required")
		ctx.UI.Hint("Usage: arngit blame <file>")
		return fmt.Errorf("missing file")
	}

	file := ctx.Args[0]
	cmd := exec.Command("git", "blame", "--color-by-age", file)
	out, err := cmd.Output()
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to blame: %v", err))
		return err
	}

	fmt.Println(string(out))
	return nil
}

// handleAccountCheck validates the current account's PAT.
func (r *Router) handleAccountCheck(ctx *Context) error {
	acc := r.engine.Accounts().Current()
	if acc == nil {
		ctx.UI.Error("No account configured")
		ctx.UI.Hint("Add one with: arngit account add <name>")
		return fmt.Errorf("no account")
	}

	pat, err := r.engine.Accounts().GetPAT(acc.Name)
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to get PAT: %v", err))
		return err
	}

	client := github.NewClient(acc.Username, pat)
	info, err := client.ValidatePAT()
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to validate PAT: %v", err))
		return err
	}

	if !info.Valid {
		ctx.UI.Error("PAT is invalid or expired")
		ctx.UI.Hint("Add a new account with: arngit account add <name>")
		return fmt.Errorf("invalid PAT")
	}

	ctx.UI.Success(fmt.Sprintf("PAT is valid for: %s", info.User.Login))

	// Show scopes
	if len(info.Scopes) > 0 {
		fmt.Printf("  %s %s\n", r.ui.Color(ui.Dim, "Scopes:"), r.ui.Color(ui.Cyan, strings.Join(info.Scopes, ", ")))
	}

	// Show rate limit
	if info.RateLimit != nil {
		fmt.Printf("  %s %d/%d (resets %s)\n",
			r.ui.Color(ui.Dim, "Rate Limit:"),
			info.RateLimit.Remaining,
			info.RateLimit.Limit,
			info.RateLimit.Reset.Format("15:04:05"))
	}

	fmt.Println()
	return nil
}
