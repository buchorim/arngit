package command

import (
	"fmt"
	"strings"

	"github.com/arfrfrr/arngit/internal/ui"
)

// handleStashHelp shows stash help.
func (r *Router) handleStashHelp(ctx *Context) error {
	ctx.UI.Title("Stash Management")
	ctx.UI.Info("Subcommands:")
	ctx.UI.Info("  save     Stash changes")
	ctx.UI.Info("  list     List stashes")
	ctx.UI.Info("  pop      Pop a stash")
	return nil
}

// handleStashSave stashes changes.
func (r *Router) handleStashSave(ctx *Context) error {
	svc := r.getGitService()

	message := ""
	if len(ctx.Args) > 0 {
		message = strings.Join(ctx.Args, " ")
	}

	if err := svc.Stash(message); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success("Changes stashed")
	return nil
}

// handleStashList lists stashes.
func (r *Router) handleStashList(ctx *Context) error {
	svc := r.getGitService()

	stashes, err := svc.StashList()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if len(stashes) == 0 {
		ctx.UI.Info("No stashes")
		return nil
	}

	ctx.UI.Title("Stashes")

	for _, s := range stashes {
		fmt.Printf("  %s %s: %s\n",
			r.ui.Color(ui.Yellow, fmt.Sprintf("stash@{%d}", s.Index)),
			r.ui.Color(ui.Cyan, s.Branch),
			s.Message)
	}

	return nil
}

// handleStashPop pops a stash.
func (r *Router) handleStashPop(ctx *Context) error {
	svc := r.getGitService()

	index := -1
	if len(ctx.Args) > 0 {
		fmt.Sscanf(ctx.Args[0], "%d", &index)
	}

	if err := svc.StashPop(index); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success("Stash applied and removed")
	return nil
}
