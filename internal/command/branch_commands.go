package command

import (
	"fmt"

	"github.com/arfrfrr/arngit/internal/ui"
)

// handleBranchHelp shows branch help.
func (r *Router) handleBranchHelp(ctx *Context) error {
	ctx.UI.Title("Branch Management")
	ctx.UI.Info("Subcommands:")
	ctx.UI.Info("  list     List all branches")
	ctx.UI.Info("  new      Create a new branch")
	ctx.UI.Info("  switch   Switch to a branch")
	ctx.UI.Info("  delete   Delete a branch")
	return nil
}

// handleBranchList lists branches.
func (r *Router) handleBranchList(ctx *Context) error {
	svc := r.getGitService()

	all := false
	for _, arg := range ctx.Args {
		if arg == "-a" || arg == "--all" {
			all = true
		}
	}

	branches, err := svc.Branches(all)
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Title("Branches")

	for _, b := range branches {
		marker := "  "
		if b.Current {
			marker = "* "
		}

		name := b.Name
		if b.Remote {
			name = r.ui.Color(ui.Red, name)
		} else if b.Current {
			name = r.ui.Color(ui.BrightGreen, name)
		}

		fmt.Printf("%s%s %s\n", marker, name, r.ui.Color(ui.Dim, b.Commit[:7]))
	}

	return nil
}

// handleBranchNew creates a new branch.
func (r *Router) handleBranchNew(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit branch new <name>")
	}

	svc := r.getGitService()
	name := ctx.Args[0]

	if err := svc.CreateBranch(name, true); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Created and switched to branch '%s'", name))
	return nil
}

// handleBranchSwitch switches branch.
func (r *Router) handleBranchSwitch(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit branch switch <name>")
	}

	svc := r.getGitService()
	name := ctx.Args[0]

	if err := svc.SwitchBranch(name); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Switched to branch '%s'", name))
	return nil
}

// handleBranchDelete deletes a branch.
func (r *Router) handleBranchDelete(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit branch delete <name>")
	}

	svc := r.getGitService()
	name := ctx.Args[0]
	force := false

	for _, arg := range ctx.Args {
		if arg == "-f" || arg == "--force" {
			force = true
		}
	}

	if err := svc.DeleteBranch(name, force); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Deleted branch '%s'", name))
	return nil
}
