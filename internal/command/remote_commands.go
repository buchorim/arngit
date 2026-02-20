package command

import (
	"fmt"

	"github.com/arfrfrr/arngit/internal/ui"
)

// handleRemoteHelp shows remote help.
func (r *Router) handleRemoteHelp(ctx *Context) error {
	ctx.UI.Title("Remote Management")
	ctx.UI.Info("Subcommands:")
	ctx.UI.Info("  list     List all remotes")
	ctx.UI.Info("  add      Add a remote")
	ctx.UI.Info("  remove   Remove a remote")
	return nil
}

// handleRemoteList lists remotes.
func (r *Router) handleRemoteList(ctx *Context) error {
	svc := r.getGitService()

	remotes, err := svc.Remotes()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if len(remotes) == 0 {
		ctx.UI.Info("No remotes configured")
		return nil
	}

	ctx.UI.Title("Remotes")

	for _, remote := range remotes {
		fmt.Printf("  %s\n", r.ui.Color(ui.BrightCyan, remote.Name))
		fmt.Printf("    fetch: %s\n", remote.FetchURL)
		if remote.PushURL != remote.FetchURL {
			fmt.Printf("    push:  %s\n", remote.PushURL)
		}
	}

	return nil
}

// handleRemoteAdd adds a remote.
func (r *Router) handleRemoteAdd(ctx *Context) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("usage: arngit remote add <name> <url>")
	}

	svc := r.getGitService()
	name := ctx.Args[0]
	url := ctx.Args[1]

	if err := svc.AddRemote(name, url); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Added remote '%s'", name))
	return nil
}

// handleRemoteRemove removes a remote.
func (r *Router) handleRemoteRemove(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit remote remove <name>")
	}

	svc := r.getGitService()
	name := ctx.Args[0]

	if err := svc.RemoveRemote(name); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Removed remote '%s'", name))
	return nil
}
