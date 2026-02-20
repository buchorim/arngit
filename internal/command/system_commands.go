package command

import (
	"fmt"
	"os"

	"github.com/arfrfrr/arngit/internal/ui"
)

// RegisterSystemCommands registers system/engine commands.
func (r *Router) RegisterSystemCommands() {
	// Protect command
	r.Register(&Command{
		Name:        "protect",
		Description: "Protect the current repository",
		Usage:       "arngit protect [--password]",
		Handler:     r.handleProtect,
	})

	// Unprotect command
	r.Register(&Command{
		Name:        "unprotect",
		Description: "Remove protection from repository",
		Usage:       "arngit unprotect",
		Handler:     r.handleUnprotect,
	})

	// Update command
	r.Register(&Command{
		Name:        "update",
		Description: "Check for and apply updates",
		Usage:       "arngit update [check|apply|rollback]",
		Handler:     r.handleUpdate,
	})

	// Doctor command
	r.Register(&Command{
		Name:        "doctor",
		Description: "Check system health",
		Usage:       "arngit doctor",
		Handler:     r.handleDoctor,
	})

	// Storage command
	r.Register(&Command{
		Name:        "storage",
		Description: "Show storage usage",
		Usage:       "arngit storage",
		Handler:     r.handleStorage,
	})

	// Logs command
	r.Register(&Command{
		Name:        "logs",
		Description: "View application logs",
		Usage:       "arngit logs [-n N]",
		Handler:     r.handleLogs,
	})
}

// handleProtect protects the current repository.
func (r *Router) handleProtect(ctx *Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		ctx.UI.Error("Failed to get current directory")
		return err
	}

	// Check if already protected
	if ctx.Engine.ProtectedRepos().IsProtected(cwd) {
		ctx.UI.Warning("This repository is already protected")
		return nil
	}

	// Ask for password (optional)
	password := ""
	for _, arg := range ctx.Args {
		if arg == "--password" || arg == "-p" {
			password = ctx.UI.PromptSecret("Set protection password (optional)")
			break
		}
	}

	if err := ctx.Engine.ProtectedRepos().Protect(cwd, password); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success("Repository protected")
	if password != "" {
		ctx.UI.Info("Password protection enabled")
	}
	ctx.UI.Hint("All push operations will require confirmation")

	return nil
}

// handleUnprotect removes protection from repository.
func (r *Router) handleUnprotect(ctx *Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		ctx.UI.Error("Failed to get current directory")
		return err
	}

	// Check if protected
	protection := ctx.Engine.ProtectedRepos().GetProtection(cwd)
	if protection == nil {
		ctx.UI.Info("This repository is not protected")
		return nil
	}

	// Verify password if set
	password := ""
	if protection.Password != "" {
		password = ctx.UI.PromptSecret("Enter protection password")
	}

	if err := ctx.Engine.ProtectedRepos().Unprotect(cwd, password); err != nil {
		ctx.UI.Error("Failed to unprotect: invalid password")
		return err
	}

	ctx.UI.Success("Protection removed")
	return nil
}

// handleUpdate handles update commands.
func (r *Router) handleUpdate(ctx *Context) error {
	action := "check"
	if len(ctx.Args) > 0 {
		action = ctx.Args[0]
	}

	um := ctx.Engine.UpdateManager()

	switch action {
	case "check":
		ctx.UI.Info("Checking for updates...")

		update, err := um.CheckForUpdate(ctx.Engine.Version())
		if err != nil {
			ctx.UI.Error(err.Error())
			return err
		}

		if update == nil {
			ctx.UI.Success("You're running the latest version")
			return nil
		}

		ctx.UI.Box("Update Available", fmt.Sprintf(
			"Version: %s\nDate: %s\nSize: %.2f MB",
			update.Version,
			update.ReleaseDate.Format("2006-01-02"),
			float64(update.Size)/(1024*1024),
		), ui.BrightCyan)

		if update.Changelog != "" {
			fmt.Println()
			ctx.UI.Title("Changelog")
			ctx.UI.Info(update.Changelog)
		}

		ctx.UI.Hint("Run 'arngit update apply' to install")

	case "apply":
		if !um.HasPendingUpdate() {
			// Check first
			update, err := um.CheckForUpdate(ctx.Engine.Version())
			if err != nil {
				ctx.UI.Error(err.Error())
				return err
			}
			if update == nil {
				ctx.UI.Success("Already up to date")
				return nil
			}
		}

		update := um.LatestUpdate()
		if update == nil {
			ctx.UI.Error("No update available")
			return nil
		}

		if !ctx.UI.Confirm(fmt.Sprintf("Install version %s?", update.Version)) {
			ctx.UI.Info("Update cancelled")
			return nil
		}

		ctx.UI.Info("Downloading update...")

		downloadPath, err := um.DownloadUpdate(update, func(downloaded, total int64) {
			ctx.UI.ProgressBar(int(downloaded), int(total), "Downloading")
		})
		if err != nil {
			ctx.UI.Error("Download failed: " + err.Error())
			return err
		}

		ctx.UI.Info("Applying update...")

		if err := um.ApplyUpdate(downloadPath); err != nil {
			ctx.UI.Error("Update failed: " + err.Error())
			ctx.UI.Hint("Run 'arngit update rollback' to restore previous version")
			return err
		}

		ctx.UI.Success("Update complete! Please restart arngit.")

	case "rollback":
		if !ctx.UI.Confirm("Rollback to previous version?") {
			ctx.UI.Info("Rollback cancelled")
			return nil
		}

		if err := um.Rollback(); err != nil {
			ctx.UI.Error(err.Error())
			return err
		}

		ctx.UI.Success("Rolled back to previous version")

	default:
		return fmt.Errorf("unknown action: %s (use check, apply, or rollback)", action)
	}

	return nil
}

// handleDoctor checks system health.
func (r *Router) handleDoctor(ctx *Context) error {
	ctx.UI.Title("System Health Check")

	checks := []struct {
		name    string
		checkFn func() (bool, string)
	}{
		{"Git installed", func() (bool, string) {
			svc := r.getGitService()
			if !svc.IsInstalled() {
				return false, "Git not found in PATH"
			}
			ver, _ := svc.Version()
			return true, "v" + ver
		}},
		{"Config loaded", func() (bool, string) {
			if ctx.Engine.Config() == nil {
				return false, "Failed to load"
			}
			return true, "OK"
		}},
		{"Storage accessible", func() (bool, string) {
			stats, err := ctx.Engine.Storage().GetStorageStats()
			if err != nil {
				return false, err.Error()
			}
			var total int64
			for _, size := range stats {
				total += size
			}
			return true, fmt.Sprintf("%.2f MB used", float64(total)/(1024*1024))
		}},
		{"Account configured", func() (bool, string) {
			acc := ctx.Engine.Accounts().Current()
			if acc == nil {
				return false, "No account"
			}
			return true, acc.Username
		}},
	}

	allPassed := true
	for _, check := range checks {
		passed, info := check.checkFn()

		status := r.ui.Color(ui.BrightGreen, "✓")
		if !passed {
			status = r.ui.Color(ui.BrightRed, "✗")
			allPassed = false
		}

		fmt.Printf("  %s %-20s %s\n", status, check.name, r.ui.Color(ui.Dim, info))
	}

	fmt.Println()
	if allPassed {
		ctx.UI.Success("All checks passed")
	} else {
		ctx.UI.Warning("Some checks failed")
	}

	return nil
}

// handleStorage shows storage usage.
func (r *Router) handleStorage(ctx *Context) error {
	stats, err := ctx.Engine.Storage().GetStorageStats()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Title("Storage Usage")

	var total int64
	for name, size := range stats {
		total += size
		bar := ""
		if size > 0 {
			barLen := int(float64(size) / float64(1024*1024) * 10) // 10 chars per MB
			if barLen > 30 {
				barLen = 30
			}
			if barLen < 1 && size > 0 {
				barLen = 1
			}
			bar = r.ui.Color(ui.BrightCyan, "█") + r.ui.Color(ui.Cyan, string(make([]byte, barLen)))
		}

		fmt.Printf("  %-12s %8.2f KB  %s\n", name, float64(size)/1024, bar)
	}

	fmt.Println()
	fmt.Printf("  %-12s %8.2f KB\n", r.ui.Color(ui.Bold, "Total"), float64(total)/1024)

	return nil
}

// handleLogs shows recent logs.
func (r *Router) handleLogs(ctx *Context) error {
	n := 20
	for i, arg := range ctx.Args {
		if arg == "-n" && i+1 < len(ctx.Args) {
			fmt.Sscanf(ctx.Args[i+1], "%d", &n)
		}
	}

	entries := ctx.Engine.Logger().Last(n)

	if len(entries) == 0 {
		ctx.UI.Info("No log entries")
		return nil
	}

	ctx.UI.Title("Recent Logs")

	for _, entry := range entries {
		levelColor := ui.Dim
		switch entry.Level {
		case 1: // Info
			levelColor = ui.Cyan
		case 2: // Warn
			levelColor = ui.Yellow
		case 3: // Error
			levelColor = ui.Red
		}

		fmt.Printf("  %s %s %s\n",
			r.ui.Color(ui.Dim, entry.Time.Format("15:04:05")),
			r.ui.Color(levelColor, entry.Level.String()),
			entry.Message)
	}

	return nil
}
