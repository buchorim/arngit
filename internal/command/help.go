package command

import (
	"fmt"
	"sort"
	"strings"

	"github.com/arfrfrr/arngit/internal/ui"
)

// CommandGroup represents a group of related commands.
type CommandGroup struct {
	Name     string
	Commands []*Command
}

// GetCommandGroups returns commands organized by category.
func (r *Router) GetCommandGroups() []*CommandGroup {
	groups := map[string]*CommandGroup{
		"git":        {Name: "Git Operations"},
		"branch":     {Name: "Branch & Tag"},
		"github":     {Name: "GitHub API"},
		"automation": {Name: "Automation"},
		"account":    {Name: "Account Management"},
		"system":     {Name: "System & Config"},
		"help":       {Name: "Help & Info"},
	}

	// Categorize commands
	gitCmds := []string{"init", "clone", "status", "add", "commit", "push", "pull", "fetch", "diff", "history", "sync"}
	branchCmds := []string{"branch", "remote", "stash", "tag"}
	githubCmds := []string{"repo", "release", "pr"}
	automationCmds := []string{"hooks", "changelog", "bump", "watch", "stats", "blame"}
	accountCmds := []string{"account"}
	systemCmds := []string{"config", "doctor", "storage", "logs", "protect", "unprotect", "update"}
	helpCmds := []string{"help", "version"}

	for name, cmd := range r.commands {
		switch {
		case contains(gitCmds, name):
			groups["git"].Commands = append(groups["git"].Commands, cmd)
		case contains(branchCmds, name):
			groups["branch"].Commands = append(groups["branch"].Commands, cmd)
		case contains(githubCmds, name):
			groups["github"].Commands = append(groups["github"].Commands, cmd)
		case contains(automationCmds, name):
			groups["automation"].Commands = append(groups["automation"].Commands, cmd)
		case contains(accountCmds, name):
			groups["account"].Commands = append(groups["account"].Commands, cmd)
		case contains(systemCmds, name):
			groups["system"].Commands = append(groups["system"].Commands, cmd)
		case contains(helpCmds, name):
			groups["help"].Commands = append(groups["help"].Commands, cmd)
		default:
			groups["system"].Commands = append(groups["system"].Commands, cmd)
		}
	}

	// Sort commands within each group
	for _, g := range groups {
		sort.Slice(g.Commands, func(i, j int) bool {
			return g.Commands[i].Name < g.Commands[j].Name
		})
	}

	// Return in display order
	return []*CommandGroup{
		groups["git"],
		groups["branch"],
		groups["github"],
		groups["automation"],
		groups["account"],
		groups["system"],
		groups["help"],
	}
}

// ShowStructuredHelp displays help with proper structure.
func (r *Router) ShowStructuredHelp(ctx *Context) {
	ctx.UI.Logo()
	fmt.Println()

	// Brief intro
	fmt.Println(r.ui.Color(ui.Bold, "  ArnGit") + r.ui.Color(ui.Dim, " - Developer Control Platform"))
	fmt.Println(r.ui.Color(ui.Dim, "  A modern Git workflow tool with multi-account support"))
	fmt.Println()

	// Group commands
	groups := r.GetCommandGroups()

	for _, group := range groups {
		if len(group.Commands) == 0 {
			continue
		}

		// Group header
		fmt.Println(r.ui.Color(ui.BrightCyan+ui.Bold, "  "+group.Name))
		fmt.Println(r.ui.Color(ui.Dim, "  "+strings.Repeat("─", len(group.Name)+2)))

		for _, cmd := range group.Commands {
			// Command name
			cmdName := r.ui.Color(ui.Yellow, fmt.Sprintf("    %-12s", cmd.Name))

			// Description
			desc := cmd.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			fmt.Printf("%s %s\n", cmdName, r.ui.Color(ui.Dim, desc))

			// Show subcommands if any
			if cmd.SubCommands != nil && len(cmd.SubCommands) > 0 {
				subNames := []string{}
				for name := range cmd.SubCommands {
					subNames = append(subNames, name)
				}
				sort.Strings(subNames)
				subList := r.ui.Color(ui.Dim, "      └─ "+strings.Join(subNames, ", "))
				fmt.Println(subList)
			}
		}
		fmt.Println()
	}

	// Footer
	fmt.Println(r.ui.Color(ui.Dim, "  Use 'arngit help <command>' for detailed information"))
	fmt.Println(r.ui.Color(ui.Dim, "  Use 'arngit' to enter interactive mode"))
	fmt.Println()
}

// ShowCommandHelp displays detailed help for a specific command.
func (r *Router) ShowCommandHelp(ctx *Context, cmdName string) error {
	cmd, exists := r.commands[cmdName]
	if !exists {
		ctx.UI.Error(fmt.Sprintf("Unknown command: %s", cmdName))
		ctx.UI.Hint("Run 'arngit help' for available commands")
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Command header
	fmt.Println()
	fmt.Println(r.ui.Color(ui.Bold+ui.BrightCyan, "  "+cmd.Name) + " - " + cmd.Description)
	fmt.Println()

	// Usage
	fmt.Println(r.ui.Color(ui.Bold, "  USAGE"))
	fmt.Println(r.ui.Color(ui.Dim, "    "+cmd.Usage))
	fmt.Println()

	// Subcommands
	if cmd.SubCommands != nil && len(cmd.SubCommands) > 0 {
		fmt.Println(r.ui.Color(ui.Bold, "  SUBCOMMANDS"))

		// Sort subcommand names
		names := []string{}
		for name := range cmd.SubCommands {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			sub := cmd.SubCommands[name]
			subName := r.ui.Color(ui.Yellow, fmt.Sprintf("    %-12s", name))
			fmt.Printf("%s %s\n", subName, r.ui.Color(ui.Dim, sub.Description))
		}
		fmt.Println()
	}

	// Examples based on command
	examples := r.getCommandExamples(cmdName)
	if len(examples) > 0 {
		fmt.Println(r.ui.Color(ui.Bold, "  EXAMPLES"))
		for _, ex := range examples {
			fmt.Println(r.ui.Color(ui.Dim, "    $ ") + r.ui.Color(ui.Cyan, ex))
		}
		fmt.Println()
	}

	return nil
}

// getCommandExamples returns example usage for a command.
func (r *Router) getCommandExamples(cmdName string) []string {
	examples := map[string][]string{
		"commit": {
			"arngit commit -m \"feat: add login\"",
			"arngit commit -m \"fix: resolve bug\"",
		},
		"push": {
			"arngit push",
			"arngit push origin main",
		},
		"branch": {
			"arngit branch list",
			"arngit branch new feature/login",
			"arngit branch switch main",
		},
		"account": {
			"arngit account add work",
			"arngit account list",
			"arngit account switch personal",
		},
		"config": {
			"arngit config",
			"arngit config set default_branch main",
		},
		"status": {
			"arngit status",
		},
		"clone": {
			"arngit clone https://github.com/user/repo.git",
		},
		"repo": {
			"arngit repo create myproject",
			"arngit repo create myproject -p",
			"arngit repo list",
		},
		"release": {
			"arngit release list user/repo",
			"arngit release create user/repo v1.0.0",
			"arngit release upload user/repo v1.0.0 ./app.exe",
		},
		"pr": {
			"arngit pr create user/repo Add login feature",
			"arngit pr list user/repo",
		},
		"hooks": {
			"arngit hooks list",
			"arngit hooks install pre-commit",
			"arngit hooks uninstall pre-commit",
		},
		"changelog": {
			"arngit changelog",
			"arngit changelog v1.0.0",
		},
		"bump": {
			"arngit bump",
			"arngit bump major",
			"arngit bump minor",
			"arngit bump patch",
		},
		"stats": {
			"arngit stats",
		},
		"blame": {
			"arngit blame main.go",
		},
	}
	return examples[cmdName]
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
