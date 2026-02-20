package command

import (
	"fmt"

	"github.com/arfrfrr/arngit/internal/ui"
)

// handleTagHelp shows tag help.
func (r *Router) handleTagHelp(ctx *Context) error {
	ctx.UI.Title("Tag Management")
	ctx.UI.Info("Subcommands:")
	ctx.UI.Info("  list     List all tags")
	ctx.UI.Info("  create   Create a tag")
	ctx.UI.Info("  delete   Delete a tag")
	return nil
}

// handleTagList lists tags.
func (r *Router) handleTagList(ctx *Context) error {
	svc := r.getGitService()

	tags, err := svc.Tags()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if len(tags) == 0 {
		ctx.UI.Info("No tags")
		return nil
	}

	ctx.UI.Title("Tags")

	for _, tag := range tags {
		fmt.Printf("  %s\n", r.ui.Color(ui.Yellow, tag))
	}

	return nil
}

// handleTagCreate creates a tag.
func (r *Router) handleTagCreate(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit tag create <name> [-m message]")
	}

	svc := r.getGitService()
	name := ctx.Args[0]
	message := ""

	for i, arg := range ctx.Args {
		if arg == "-m" && i+1 < len(ctx.Args) {
			message = ctx.Args[i+1]
		}
	}

	if err := svc.CreateTag(name, message); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Created tag '%s'", name))
	return nil
}

// handleTagDelete deletes a tag.
func (r *Router) handleTagDelete(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit tag delete <name>")
	}

	svc := r.getGitService()
	name := ctx.Args[0]

	if err := svc.DeleteTag(name); err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Deleted tag '%s'", name))
	return nil
}
