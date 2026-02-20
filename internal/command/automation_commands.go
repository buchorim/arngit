package command

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/arfrfrr/arngit/internal/automation"
	"github.com/arfrfrr/arngit/internal/ui"
)

// RegisterAutomationCommands registers automation commands.
func (r *Router) RegisterAutomationCommands() {
	// hooks command
	r.Register(&Command{
		Name:        "hooks",
		Description: "Manage git hooks",
		Usage:       "arngit hooks <subcommand>",
		Handler:     r.handleHooksHelp,
		SubCommands: map[string]*Command{
			"install": {
				Name:        "install",
				Description: "Install a git hook",
				Usage:       "arngit hooks install <type>",
				Handler:     r.handleHooksInstall,
			},
			"uninstall": {
				Name:        "uninstall",
				Description: "Uninstall a git hook",
				Usage:       "arngit hooks uninstall <type>",
				Handler:     r.handleHooksUninstall,
			},
			"list": {
				Name:        "list",
				Description: "List installed hooks",
				Usage:       "arngit hooks list",
				Handler:     r.handleHooksList,
			},
		},
	})

	// changelog command
	r.Register(&Command{
		Name:        "changelog",
		Description: "Generate changelog from commits",
		Usage:       "arngit changelog [since-tag]",
		Handler:     r.handleChangelog,
	})

	// bump command
	r.Register(&Command{
		Name:        "bump",
		Description: "Bump version (major/minor/patch)",
		Usage:       "arngit bump [major|minor|patch]",
		Handler:     r.handleBump,
	})

	// watch command
	r.Register(&Command{
		Name:        "watch",
		Description: "Auto-push watcher",
		Usage:       "arngit watch",
		Handler:     r.handleWatch,
	})
}

// --- Hooks handlers ---

func (r *Router) handleHooksHelp(ctx *Context) error {
	ctx.UI.Title("Git Hooks Management")
	fmt.Println(r.ui.Color(ui.Dim, "  Install and manage git hooks"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "install <type>  "), r.ui.Color(ui.Dim, "Install a hook"))
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "uninstall <type>"), r.ui.Color(ui.Dim, "Remove a hook"))
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "list            "), r.ui.Color(ui.Dim, "List installed hooks"))
	fmt.Println()
	fmt.Println(r.ui.Color(ui.Dim, "  Available hook types:"))
	for _, h := range automation.AvailableHooks() {
		fmt.Printf("    %s  %s\n", r.ui.Color(ui.Cyan, h), r.ui.Color(ui.Dim, automation.HookDescription(h)))
	}
	fmt.Println()
	return nil
}

func (r *Router) handleHooksInstall(ctx *Context) error {
	if len(ctx.Args) < 1 {
		ctx.UI.Error("Hook type required")
		ctx.UI.Hint("Available: " + strings.Join(automation.AvailableHooks(), ", "))
		return fmt.Errorf("missing hook type")
	}

	hookType := ctx.Args[0]
	if !automation.ValidateHookType(hookType) {
		ctx.UI.Error(fmt.Sprintf("Invalid hook type: %s", hookType))
		ctx.UI.Hint("Available: " + strings.Join(automation.AvailableHooks(), ", "))
		return fmt.Errorf("invalid hook type")
	}

	cwd, _ := os.Getwd()
	if err := automation.InstallHook(cwd, hookType); err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to install hook: %v", err))
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Installed hook: %s", hookType))
	ctx.UI.Hint(automation.HookDescription(hookType))
	return nil
}

func (r *Router) handleHooksUninstall(ctx *Context) error {
	if len(ctx.Args) < 1 {
		ctx.UI.Error("Hook type required")
		return fmt.Errorf("missing hook type")
	}

	hookType := ctx.Args[0]
	cwd, _ := os.Getwd()
	if err := automation.UninstallHook(cwd, hookType); err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to uninstall hook: %v", err))
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Uninstalled hook: %s", hookType))
	return nil
}

func (r *Router) handleHooksList(ctx *Context) error {
	cwd, _ := os.Getwd()
	installed, err := automation.ListInstalledHooks(cwd)
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to list hooks: %v", err))
		return err
	}

	ctx.UI.Title("Installed Hooks")
	if len(installed) == 0 {
		ctx.UI.Info("  No hooks installed")
		ctx.UI.Hint("Install with: arngit hooks install <type>")
	} else {
		for _, h := range installed {
			fmt.Printf("  %s %s  %s\n",
				r.ui.Color(ui.BrightGreen, ui.SymbolCheck),
				r.ui.Color(ui.Cyan, h),
				r.ui.Color(ui.Dim, automation.HookDescription(h)))
		}
	}
	fmt.Println()
	return nil
}

// --- Changelog handler ---

func (r *Router) handleChangelog(ctx *Context) error {
	since := ""
	if len(ctx.Args) > 0 {
		since = ctx.Args[0]
	} else {
		// Try to get latest tag
		if tag, err := automation.GetLatestTag(); err == nil {
			since = tag
		}
	}

	commits, err := automation.GetCommitsSince(since)
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to get commits: %v", err))
		return err
	}

	if len(commits) == 0 {
		ctx.UI.Info("No commits found for changelog")
		return nil
	}

	version := "Unreleased"
	if since != "" {
		version = fmt.Sprintf("since %s", since)
	}

	entry := automation.GenerateChangelog(commits, version)
	output := automation.FormatChangelog(entry)

	fmt.Println(output)
	return nil
}

// --- Bump handler ---

func (r *Router) handleBump(ctx *Context) error {
	bumpType := ""
	if len(ctx.Args) > 0 {
		bumpType = ctx.Args[0]
	}

	current, err := automation.GetCurrentVersion()
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to get current version: %v", err))
		return err
	}

	// Auto-detect bump type if not specified
	if bumpType == "" {
		since := current.String()
		commits, _ := automation.GetCommitsSince(since)
		if len(commits) > 0 {
			bumpType = automation.DetermineBumpType(commits)
		} else {
			bumpType = "patch"
		}
		ctx.UI.Info(fmt.Sprintf("  Auto-detected bump type: %s", bumpType))
	}

	switch bumpType {
	case "major":
		current.BumpMajor()
	case "minor":
		current.BumpMinor()
	case "patch":
		current.BumpPatch()
	default:
		ctx.UI.Error(fmt.Sprintf("Invalid bump type: %s", bumpType))
		ctx.UI.Hint("Use: major, minor, or patch")
		return fmt.Errorf("invalid bump type")
	}

	newVersion := current.String()

	if err := automation.CreateVersionTag(newVersion, fmt.Sprintf("Release %s", newVersion)); err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to create tag: %v", err))
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Bumped version to: %s", newVersion))
	ctx.UI.Hint("Push tag with: arngit push --tags")
	return nil
}

// --- Watch handler ---

func (r *Router) handleWatch(ctx *Context) error {
	gitSvc := r.getGitService()
	if !gitSvc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a repo")
	}

	// Parse args: arngit watch [--commits N] [--time DURATION] [--size SIZE] [--interval DURATION]
	cfg := automation.WatcherConfig{
		ThresholdType:  automation.ThresholdCommits,
		ThresholdValue: "3",
		Interval:       10 * time.Second,
	}

	for i := 0; i < len(ctx.Args); i++ {
		switch ctx.Args[i] {
		case "--commits":
			if i+1 < len(ctx.Args) {
				cfg.ThresholdType = automation.ThresholdCommits
				cfg.ThresholdValue = ctx.Args[i+1]
				i++
			}
		case "--time":
			if i+1 < len(ctx.Args) {
				cfg.ThresholdType = automation.ThresholdTime
				cfg.ThresholdValue = ctx.Args[i+1]
				i++
			}
		case "--size":
			if i+1 < len(ctx.Args) {
				cfg.ThresholdType = automation.ThresholdSize
				cfg.ThresholdValue = ctx.Args[i+1]
				i++
			}
		case "--interval":
			if i+1 < len(ctx.Args) {
				d, err := time.ParseDuration(ctx.Args[i+1])
				if err == nil {
					cfg.Interval = d
				}
				i++
			}
		}
	}

	watcher, err := automation.NewWatcher(r.engine, gitSvc, cfg)
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to start watcher: %v", err))
		return err
	}

	// Setup callbacks
	watcher.OnPush = func(remote, branch string, commitCount int) {
		ctx.UI.Success(fmt.Sprintf("Pushed %d commits to %s/%s", commitCount, remote, branch))
	}
	watcher.OnError = func(err error) {
		ctx.UI.Error(fmt.Sprintf("Push failed: %v", err))
	}
	watcher.OnCheck = func(thresholdType automation.ThresholdType, current, threshold string) {
		// Silent check - only show on verbose
	}
	watcher.OnSkipped = func(reason string) {
		fmt.Printf("  %s %s\r", r.ui.Color(ui.Dim, "[watching]"), r.ui.Color(ui.Dim, reason))
	}

	ctx.UI.Title("Watch Mode")
	fmt.Printf("  %s %s\n", r.ui.Color(ui.Dim, "Config:"), r.ui.Color(ui.Cyan, watcher.Config()))
	fmt.Printf("  %s\n\n", r.ui.Color(ui.Dim, "Press Ctrl+C to stop"))

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println()
		ctx.UI.Info("Stopping watcher...")
		watcher.Stop()
	}()

	return watcher.Start()
}
