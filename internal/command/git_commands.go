package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/arfrfrr/arngit/internal/git"
	"github.com/arfrfrr/arngit/internal/ui"
)

// RegisterGitCommands registers all git-related commands.
func (r *Router) RegisterGitCommands() {
	// Init command
	r.Register(&Command{
		Name:        "init",
		Description: "Initialize a new git repository",
		Usage:       "arngit init [--bare]",
		Handler:     r.handleInit,
	})

	// Clone command
	r.Register(&Command{
		Name:        "clone",
		Description: "Clone a repository",
		Usage:       "arngit clone <url> [destination]",
		Handler:     r.handleClone,
	})

	// Status command
	r.Register(&Command{
		Name:        "status",
		Description: "Show repository status",
		Usage:       "arngit status",
		Handler:     r.handleStatus,
	})

	// Add command
	r.Register(&Command{
		Name:        "add",
		Description: "Stage files for commit",
		Usage:       "arngit add [files...] or arngit add -A",
		Handler:     r.handleAdd,
	})

	// Commit command
	r.Register(&Command{
		Name:        "commit",
		Description: "Create a commit",
		Usage:       "arngit commit -m <message>",
		Handler:     r.handleCommit,
	})

	// Push command
	r.Register(&Command{
		Name:        "push",
		Description: "Push commits to remote",
		Usage:       "arngit push [remote] [branch]",
		Handler:     r.handlePush,
	})

	// Pull command
	r.Register(&Command{
		Name:        "pull",
		Description: "Pull changes from remote",
		Usage:       "arngit pull [remote] [branch]",
		Handler:     r.handlePull,
	})

	// Fetch command
	r.Register(&Command{
		Name:        "fetch",
		Description: "Fetch from remote",
		Usage:       "arngit fetch [remote]",
		Handler:     r.handleFetch,
	})

	// Diff command
	r.Register(&Command{
		Name:        "diff",
		Description: "Show changes",
		Usage:       "arngit diff [--staged] [file]",
		Handler:     r.handleDiff,
	})

	// Log/History command
	r.Register(&Command{
		Name:        "history",
		Description: "Show commit history",
		Usage:       "arngit history [-n N]",
		Handler:     r.handleHistory,
	})

	// Branch command
	r.Register(&Command{
		Name:        "branch",
		Description: "Manage branches",
		Usage:       "arngit branch <subcommand>",
		Handler:     r.handleBranchHelp,
		SubCommands: map[string]*Command{
			"list": {
				Name:        "list",
				Description: "List all branches",
				Usage:       "arngit branch list [-a]",
				Handler:     r.handleBranchList,
			},
			"new": {
				Name:        "new",
				Description: "Create a new branch",
				Usage:       "arngit branch new <name>",
				Handler:     r.handleBranchNew,
			},
			"switch": {
				Name:        "switch",
				Description: "Switch to a branch",
				Usage:       "arngit branch switch <name>",
				Handler:     r.handleBranchSwitch,
			},
			"delete": {
				Name:        "delete",
				Description: "Delete a branch",
				Usage:       "arngit branch delete <name>",
				Handler:     r.handleBranchDelete,
			},
		},
	})

	// Remote command
	r.Register(&Command{
		Name:        "remote",
		Description: "Manage remotes",
		Usage:       "arngit remote <subcommand>",
		Handler:     r.handleRemoteHelp,
		SubCommands: map[string]*Command{
			"list": {
				Name:        "list",
				Description: "List all remotes",
				Usage:       "arngit remote list",
				Handler:     r.handleRemoteList,
			},
			"add": {
				Name:        "add",
				Description: "Add a remote",
				Usage:       "arngit remote add <name> <url>",
				Handler:     r.handleRemoteAdd,
			},
			"remove": {
				Name:        "remove",
				Description: "Remove a remote",
				Usage:       "arngit remote remove <name>",
				Handler:     r.handleRemoteRemove,
			},
		},
	})

	// Stash command
	r.Register(&Command{
		Name:        "stash",
		Description: "Manage stashes",
		Usage:       "arngit stash <subcommand>",
		Handler:     r.handleStashHelp,
		SubCommands: map[string]*Command{
			"save": {
				Name:        "save",
				Description: "Stash changes",
				Usage:       "arngit stash save [message]",
				Handler:     r.handleStashSave,
			},
			"list": {
				Name:        "list",
				Description: "List stashes",
				Usage:       "arngit stash list",
				Handler:     r.handleStashList,
			},
			"pop": {
				Name:        "pop",
				Description: "Pop a stash",
				Usage:       "arngit stash pop [index]",
				Handler:     r.handleStashPop,
			},
		},
	})

	// Tag command
	r.Register(&Command{
		Name:        "tag",
		Description: "Manage tags",
		Usage:       "arngit tag <subcommand>",
		Handler:     r.handleTagHelp,
		SubCommands: map[string]*Command{
			"list": {
				Name:        "list",
				Description: "List all tags",
				Usage:       "arngit tag list",
				Handler:     r.handleTagList,
			},
			"create": {
				Name:        "create",
				Description: "Create a tag",
				Usage:       "arngit tag create <name> [-m message]",
				Handler:     r.handleTagCreate,
			},
			"delete": {
				Name:        "delete",
				Description: "Delete a tag",
				Usage:       "arngit tag delete <name>",
				Handler:     r.handleTagDelete,
			},
		},
	})

	// Sync command (fetch + rebase)
	r.Register(&Command{
		Name:        "sync",
		Description: "Sync with remote (fetch + pull --rebase)",
		Usage:       "arngit sync",
		Handler:     r.handleSync,
	})
}

// getGitService returns the git service.
func (r *Router) getGitService() *git.Service {
	return git.NewService(r.engine)
}

// handleInit initializes a repository.
func (r *Router) handleInit(ctx *Context) error {
	svc := r.getGitService()

	bare := false
	for _, arg := range ctx.Args {
		if arg == "--bare" {
			bare = true
		}
	}

	if err := svc.Init(bare); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success("Initialized git repository")
	return nil
}

// handleClone clones a repository.
func (r *Router) handleClone(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit clone <url> [destination]")
	}

	svc := r.getGitService()
	url := ctx.Args[0]
	dest := ""
	if len(ctx.Args) > 1 {
		dest = ctx.Args[1]
	}

	ctx.UI.Info(fmt.Sprintf("Cloning %s...", url))

	if err := svc.Clone(url, dest, 0); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success("Repository cloned")
	return nil
}

// handleStatus shows repository status.
func (r *Router) handleStatus(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		ctx.UI.Hint("Run 'arngit init' to initialize")
		return fmt.Errorf("not a git repository")
	}

	status, err := svc.Status()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	// Branch info
	ctx.UI.Title(fmt.Sprintf("On branch %s", status.Branch))

	if status.Ahead > 0 || status.Behind > 0 {
		info := []string{}
		if status.Ahead > 0 {
			info = append(info, fmt.Sprintf("%d ahead", status.Ahead))
		}
		if status.Behind > 0 {
			info = append(info, fmt.Sprintf("%d behind", status.Behind))
		}
		ctx.UI.Info(fmt.Sprintf("  %s", strings.Join(info, ", ")))
	}

	// Conflicts
	if len(status.Conflicts) > 0 {
		fmt.Println()
		ctx.UI.Warning("Conflicts:")
		for _, f := range status.Conflicts {
			fmt.Printf("  %s %s\n", r.ui.Color(ui.Red, "C"), f.Path)
		}
	}

	// Staged changes
	if len(status.Staged) > 0 {
		fmt.Println()
		ctx.UI.Info("Staged changes:")
		for _, f := range status.Staged {
			color := ui.Green
			symbol := "M"
			switch f.Status {
			case "A":
				symbol = "A"
				color = ui.BrightGreen
			case "D":
				symbol = "D"
				color = ui.Red
			case "R":
				symbol = "R"
				color = ui.Cyan
			}
			fmt.Printf("  %s %s\n", r.ui.Color(color, symbol), f.Path)
		}
	}

	// Modified files
	if len(status.Modified) > 0 {
		fmt.Println()
		ctx.UI.Info("Modified:")
		for _, f := range status.Modified {
			color := ui.Yellow
			symbol := "M"
			if f.Status == "D" {
				symbol = "D"
				color = ui.Red
			}
			fmt.Printf("  %s %s\n", r.ui.Color(color, symbol), f.Path)
		}
	}

	// Untracked files
	if len(status.Untracked) > 0 {
		fmt.Println()
		ctx.UI.Info("Untracked:")
		for _, f := range status.Untracked {
			fmt.Printf("  %s %s\n", r.ui.Color(ui.Dim, "?"), f.Path)
		}
	}

	if !status.HasChanges() {
		ctx.UI.Success("Working tree clean")
	}

	return nil
}

// handleAdd stages files.
func (r *Router) handleAdd(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a git repository")
	}

	files := ctx.Args
	if err := svc.Add(files...); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if len(files) == 0 || (len(files) == 1 && files[0] == "-A") {
		ctx.UI.Success("All changes staged")
	} else {
		ctx.UI.Success(fmt.Sprintf("Staged %d file(s)", len(files)))
	}

	return nil
}

// handleCommit creates a commit.
func (r *Router) handleCommit(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a git repository")
	}

	// Parse message from -m flag
	message := ""
	for i, arg := range ctx.Args {
		if arg == "-m" && i+1 < len(ctx.Args) {
			message = ctx.Args[i+1]
			break
		}
	}

	if message == "" {
		message = ctx.UI.Prompt("Commit message")
	}

	if message == "" {
		return fmt.Errorf("commit message required")
	}

	// Check if auto-stage is enabled
	if ctx.Engine.Config().AutoStage {
		svc.Add()
	}

	if err := svc.Commit(message, false); err != nil {
		if strings.Contains(err.Error(), "nothing to commit") {
			ctx.UI.Warning("Nothing to commit")
			ctx.UI.Hint("Stage changes with 'arngit add' first")
			return nil
		}
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Committed: %s", message))

	// Auto-push if configured
	if ctx.Engine.Config().PushAfterCommit {
		ctx.UI.Info("Auto-pushing...")
		if err := svc.Push("", "", false, false); err != nil {
			ctx.UI.Warning("Push failed: " + err.Error())
		} else {
			ctx.UI.Success("Pushed")
		}
	}

	return nil
}

// handlePush pushes commits.
func (r *Router) handlePush(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		ctx.UI.Hint("Initialize with 'init' or clone a repo")
		return fmt.Errorf("not a git repository")
	}

	remote := ""
	branch := ""
	force := false
	setUpstream := false

	for i, arg := range ctx.Args {
		switch arg {
		case "-f", "--force":
			force = true
		case "-u", "--set-upstream":
			setUpstream = true
		default:
			if remote == "" && !strings.HasPrefix(arg, "-") {
				remote = arg
			} else if branch == "" && !strings.HasPrefix(arg, "-") {
				branch = ctx.Args[i]
			}
		}
	}

	// Check if repository is protected
	wd, _ := os.Getwd()
	protection := ctx.Engine.ProtectedRepos().GetProtection(wd)
	if protection != nil {
		ctx.UI.Warning("This repository is protected")

		// Check password if set
		if protection.Password != "" {
			password := ctx.UI.PromptSecret("Enter protection password")
			if !ctx.Engine.ProtectedRepos().VerifyAccess(wd, password) {
				ctx.UI.Error("Invalid password")
				return fmt.Errorf("access denied")
			}
		} else {
			// No password, just confirm
			if !ctx.UI.Confirm("Proceed with push?") {
				ctx.UI.Info("Push cancelled")
				return nil
			}
		}
	}

	// Get current branch for display
	currentBranch, _ := svc.CurrentBranch()
	if branch == "" {
		branch = currentBranch
	}
	if remote == "" {
		remote = "origin"
	}

	ctx.UI.Info(fmt.Sprintf("Pushing %s → %s/%s...", currentBranch, remote, branch))

	if err := svc.Push(remote, branch, force, setUpstream); err != nil {
		ctx.UI.Error("Push failed")
		ctx.UI.Info("  " + err.Error())
		ctx.UI.Hint("Try 'pull' first to sync with remote")
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Pushed %s → %s", currentBranch, remote))
	return nil
}

// handlePull pulls changes.
func (r *Router) handlePull(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a git repository")
	}

	remote := ""
	branch := ""
	rebase := false

	for _, arg := range ctx.Args {
		switch arg {
		case "-r", "--rebase":
			rebase = true
		default:
			if remote == "" && !strings.HasPrefix(arg, "-") {
				remote = arg
			} else if branch == "" && !strings.HasPrefix(arg, "-") {
				branch = arg
			}
		}
	}

	ctx.UI.Info("Pulling...")

	if err := svc.Pull(remote, branch, rebase); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success("Pulled successfully")
	return nil
}

// handleFetch fetches from remote.
func (r *Router) handleFetch(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a git repository")
	}

	remote := ""
	prune := false

	for _, arg := range ctx.Args {
		switch arg {
		case "-p", "--prune":
			prune = true
		default:
			if !strings.HasPrefix(arg, "-") {
				remote = arg
			}
		}
	}

	ctx.UI.Info("Fetching...")

	if err := svc.Fetch(remote, prune); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success("Fetched successfully")
	return nil
}

// handleDiff shows diff.
func (r *Router) handleDiff(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a git repository")
	}

	staged := false
	file := ""

	for _, arg := range ctx.Args {
		switch arg {
		case "--staged", "--cached":
			staged = true
		default:
			if !strings.HasPrefix(arg, "-") {
				file = arg
			}
		}
	}

	diff, err := svc.Diff(staged, file)
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if diff == "" {
		ctx.UI.Info("No changes")
		return nil
	}

	// Print with syntax highlighting
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+"):
			fmt.Println(r.ui.Color(ui.Green, line))
		case strings.HasPrefix(line, "-"):
			fmt.Println(r.ui.Color(ui.Red, line))
		case strings.HasPrefix(line, "@@"):
			fmt.Println(r.ui.Color(ui.Cyan, line))
		default:
			fmt.Println(line)
		}
	}

	return nil
}

// handleHistory shows commit history.
func (r *Router) handleHistory(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a git repository")
	}

	n := 10
	graph := false
	for i, arg := range ctx.Args {
		switch arg {
		case "-n":
			if i+1 < len(ctx.Args) {
				fmt.Sscanf(ctx.Args[i+1], "%d", &n)
			}
		case "--graph":
			graph = true
		}
	}

	if graph {
		return r.handleHistoryGraph(ctx, n)
	}

	commits, err := svc.Log(n, false)
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if len(commits) == 0 {
		ctx.UI.Info("No commits yet")
		return nil
	}

	ctx.UI.Title("Commit History")

	for _, c := range commits {
		hash := r.ui.Color(ui.Yellow, c.Hash[:7])
		date := r.ui.Color(ui.Dim, c.Date.Format("2006-01-02 15:04"))
		author := r.ui.Color(ui.Cyan, c.Author)

		fmt.Printf("%s %s %s\n", hash, date, author)
		fmt.Printf("  %s\n\n", c.Message)
	}

	return nil
}

// handleHistoryGraph shows commit history with ASCII graph.
func (r *Router) handleHistoryGraph(ctx *Context, n int) error {
	svc := r.getGitService()

	// Use git log --graph directly for proper graph rendering
	out, err := svc.LogGraph(n)
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if out == "" {
		ctx.UI.Info("No commits yet")
		return nil
	}

	ctx.UI.Title("Commit Graph")

	// Colorize the graph output
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		switch {
		case strings.Contains(line, "* "):
			// Colorize hash and message
			fmt.Println(r.ui.Color(ui.Yellow, line))
		case strings.HasPrefix(line, "|") || strings.HasPrefix(line, "\\") || strings.HasPrefix(line, "/"):
			fmt.Println(r.ui.Color(ui.Dim, line))
		default:
			fmt.Println(line)
		}
	}

	return nil
}

// handleSync syncs with remote.
func (r *Router) handleSync(ctx *Context) error {
	svc := r.getGitService()

	if !svc.IsRepo() {
		ctx.UI.Error("Not a git repository")
		return fmt.Errorf("not a git repository")
	}

	ctx.UI.Info("Fetching...")
	if err := svc.Fetch("", true); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Info("Rebasing...")
	if err := svc.Pull("", "", true); err != nil {
		ctx.UI.Warning("Rebase failed, trying merge...")
		if err := svc.Pull("", "", false); err != nil {
			ctx.UI.Error(err.Error())
			return err
		}
	}

	ctx.UI.Success("Synced successfully")
	return nil
}
