// Package command provides command routing and context for CLI commands.
package command

import (
	"fmt"
	"strings"

	"github.com/arfrfrr/arngit/internal/core"
	"github.com/arfrfrr/arngit/internal/ui"
)

// Handler is a function that handles a command.
type Handler func(ctx *Context) error

// Command represents a CLI command.
type Command struct {
	Name        string
	Description string
	Usage       string
	Handler     Handler
	SubCommands map[string]*Command
}

// Router routes CLI commands to their handlers.
type Router struct {
	engine   *core.Engine
	commands map[string]*Command
	ui       *ui.Renderer
}

// NewRouter creates a new command router.
func NewRouter(engine *core.Engine) *Router {
	r := &Router{
		engine:   engine,
		commands: make(map[string]*Command),
		ui:       ui.NewRenderer(engine.Config().ColorOutput),
	}

	r.registerCommands()
	r.RegisterGitCommands()
	r.RegisterSystemCommands()
	r.RegisterGitHubCommands()
	r.RegisterAutomationCommands()
	r.RegisterAnalyticsCommands()
	return r
}

// registerCommands registers all available commands.
func (r *Router) registerCommands() {
	// Version command
	r.Register(&Command{
		Name:        "version",
		Description: "Show version information",
		Usage:       "arngit version",
		Handler:     r.handleVersion,
	})

	// Help command
	r.Register(&Command{
		Name:        "help",
		Description: "Show help information",
		Usage:       "arngit help [command]",
		Handler:     r.handleHelp,
	})

	// Config command
	r.Register(&Command{
		Name:        "config",
		Description: "Manage configuration",
		Usage:       "arngit config [get|set] [key] [value]",
		Handler:     r.handleConfig,
	})

	// Account command with subcommands
	r.Register(&Command{
		Name:        "account",
		Description: "Manage GitHub accounts",
		Usage:       "arngit account <subcommand>",
		Handler:     r.handleAccountHelp,
		SubCommands: map[string]*Command{
			"add": {
				Name:        "add",
				Description: "Add a new account",
				Usage:       "arngit account add <name>",
				Handler:     r.handleAccountAdd,
			},
			"remove": {
				Name:        "remove",
				Description: "Remove an account",
				Usage:       "arngit account remove <name>",
				Handler:     r.handleAccountRemove,
			},
			"list": {
				Name:        "list",
				Description: "List all accounts",
				Usage:       "arngit account list",
				Handler:     r.handleAccountList,
			},
			"switch": {
				Name:        "switch",
				Description: "Switch to a different account",
				Usage:       "arngit account switch <name>",
				Handler:     r.handleAccountSwitch,
			},
			"current": {
				Name:        "current",
				Description: "Show current account",
				Usage:       "arngit account current",
				Handler:     r.handleAccountCurrent,
			},
			"check": {
				Name:        "check",
				Description: "Validate PAT and show scopes",
				Usage:       "arngit account check",
				Handler:     r.handleAccountCheck,
			},
			"test": {
				Name:        "test",
				Description: "Test PAT validity (alias for check)",
				Usage:       "arngit account test",
				Handler:     r.handleAccountCheck,
			},
		},
	})
}

// Register adds a command to the router.
func (r *Router) Register(cmd *Command) {
	r.commands[cmd.Name] = cmd
}

// Execute parses and executes a command.
func (r *Router) Execute(args []string) error {
	// No args: show dashboard
	if len(args) == 0 {
		return r.showDashboard()
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	// Find command
	cmd, exists := r.commands[cmdName]
	if !exists {
		r.ui.Error(fmt.Sprintf("Unknown command: %s", cmdName))
		r.ui.Hint("Run 'arngit help' for available commands")
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Check for subcommand
	if len(cmdArgs) > 0 && cmd.SubCommands != nil {
		if subCmd, ok := cmd.SubCommands[cmdArgs[0]]; ok {
			cmd = subCmd
			cmdArgs = cmdArgs[1:]
		}
	}

	// Create context
	ctx := &Context{
		Engine: r.engine,
		UI:     r.ui,
		Args:   cmdArgs,
	}

	// Execute handler
	return cmd.Handler(ctx)
}

// showDashboard displays the main dashboard and enters interactive mode.
func (r *Router) showDashboard() error {
	// Show dashboard first
	ui.ShowDashboard(r.engine, r.ui)

	// Enter simple interactive mode (stable)
	return r.runInteractiveMode()
}

// runInteractiveMode runs the interactive REPL.
func (r *Router) runInteractiveMode() error {
	fmt.Println()
	r.ui.Info("Interactive mode. Type 'help' for commands, 'exit' to quit.")
	fmt.Println()

	for {
		// Show prompt
		input := r.ui.Prompt(r.ui.Color(ui.BrightCyan, "Arngit"))

		// Trim and check for empty input
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Check for exit commands
		if input == "exit" || input == "quit" || input == "q" {
			r.ui.Info("Goodbye!")
			return nil
		}

		// Clear screen command
		if input == "clear" || input == "cls" {
			fmt.Print("\033[2J\033[H")
			continue
		}

		// Parse and execute command
		args := strings.Fields(input)
		if err := r.executeCommand(args); err != nil {
			// Error already printed by executeCommand
		}

		fmt.Println() // Add spacing between commands
	}
}

// executeCommand executes a single command.
func (r *Router) executeCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	// Find command
	cmd, exists := r.commands[cmdName]
	if !exists {
		r.ui.Error(fmt.Sprintf("Unknown command: %s", cmdName))
		r.ui.Hint("Type 'help' for available commands")
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Check for subcommand
	if len(cmdArgs) > 0 && cmd.SubCommands != nil {
		if subCmd, ok := cmd.SubCommands[cmdArgs[0]]; ok {
			cmd = subCmd
			cmdArgs = cmdArgs[1:]
		}
	}

	// Create context
	ctx := &Context{
		Engine: r.engine,
		UI:     r.ui,
		Args:   cmdArgs,
	}

	// Execute handler
	return cmd.Handler(ctx)
}

// handleVersion shows version information.
func (r *Router) handleVersion(ctx *Context) error {
	ctx.UI.Logo()
	ctx.UI.Info(fmt.Sprintf("Version: %s", ctx.Engine.Version()))
	ctx.UI.Info(fmt.Sprintf("Build:   %s", ctx.Engine.BuildTime()))
	ctx.UI.Info(fmt.Sprintf("Commit:  %s", ctx.Engine.GitCommit()))
	return nil
}

// handleHelp shows help information.
func (r *Router) handleHelp(ctx *Context) error {
	if len(ctx.Args) > 0 {
		// Show detailed help for specific command
		return r.ShowCommandHelp(ctx, ctx.Args[0])
	}

	// Show structured help
	r.ShowStructuredHelp(ctx)
	return nil
}

// handleConfig manages configuration.
func (r *Router) handleConfig(ctx *Context) error {
	if len(ctx.Args) == 0 {
		// Show all config
		ctx.UI.Title("Configuration")
		config := ctx.Engine.Config()

		items := []struct{ key, value string }{
			{"default_account", config.DefaultAccount},
			{"theme", config.Theme},
			{"update_channel", config.UpdateChannel},
			{"update_interval", fmt.Sprintf("%d hours", config.UpdateInterval)},
			{"color_output", fmt.Sprintf("%v", config.ColorOutput)},
			{"default_branch", config.DefaultBranch},
		}

		for _, item := range items {
			val := item.value
			if val == "" {
				val = "(not set)"
			}
			ctx.UI.Info(fmt.Sprintf("  %-18s %s", item.key+":", val))
		}
		return nil
	}

	action := ctx.Args[0]

	switch action {
	case "get":
		if len(ctx.Args) < 2 {
			return fmt.Errorf("usage: arngit config get <key>")
		}
		key := ctx.Args[1]
		value := ctx.Engine.Config().Get(key)
		if value == nil {
			return fmt.Errorf("unknown config key: %s", key)
		}
		ctx.UI.Info(fmt.Sprintf("%s = %v", key, value))

	case "set":
		if len(ctx.Args) < 3 {
			return fmt.Errorf("usage: arngit config set <key> <value>")
		}
		key := ctx.Args[1]
		value := strings.Join(ctx.Args[2:], " ")

		// Parse value based on key type
		var parsedValue interface{}
		switch key {
		case "update_interval":
			var v int
			fmt.Sscanf(value, "%d", &v)
			parsedValue = v
		case "compact_mode", "color_output", "auto_stage", "sign_commits", "push_after_commit":
			parsedValue = value == "true" || value == "1" || value == "yes"
		default:
			parsedValue = value
		}

		if !ctx.Engine.Config().Set(key, parsedValue) {
			return fmt.Errorf("unknown config key: %s", key)
		}
		if err := ctx.Engine.Config().Save(); err != nil {
			return err
		}
		ctx.UI.Success(fmt.Sprintf("Set %s = %v", key, parsedValue))

	default:
		return fmt.Errorf("unknown action: %s (use 'get' or 'set')", action)
	}

	return nil
}

// handleAccountHelp shows account subcommand help.
func (r *Router) handleAccountHelp(ctx *Context) error {
	ctx.UI.Title("Account Management")
	ctx.UI.Info("Manage multiple GitHub accounts with encrypted PAT storage.\n")
	ctx.UI.Info("Subcommands:")
	ctx.UI.Info("  add      Add a new account")
	ctx.UI.Info("  remove   Remove an account")
	ctx.UI.Info("  list     List all accounts")
	ctx.UI.Info("  switch   Switch to a different account")
	ctx.UI.Info("  current  Show current account")
	ctx.UI.Info("  check    Validate PAT and show scopes")
	ctx.UI.Info("  test     Test PAT validity")
	return nil
}

// handleAccountAdd adds a new account.
func (r *Router) handleAccountAdd(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit account add <name>")
	}

	name := ctx.Args[0]

	// Prompt for details
	ctx.UI.Title("Add GitHub Account")
	ctx.UI.Info(fmt.Sprintf("Account name: %s\n", name))

	username := ctx.UI.Prompt("GitHub username")
	email := ctx.UI.Prompt("Email")
	pat := ctx.UI.PromptSecret("Personal Access Token (PAT)")

	if err := ctx.Engine.Accounts().Add(name, username, email, pat); err != nil {
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Account '%s' added successfully", name))
	return nil
}

// handleAccountRemove removes an account.
func (r *Router) handleAccountRemove(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit account remove <name>")
	}

	name := ctx.Args[0]

	if !ctx.UI.Confirm(fmt.Sprintf("Remove account '%s'?", name)) {
		ctx.UI.Info("Cancelled")
		return nil
	}

	if err := ctx.Engine.Accounts().Remove(name); err != nil {
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Account '%s' removed", name))
	return nil
}

// handleAccountList lists all accounts.
func (r *Router) handleAccountList(ctx *Context) error {
	accounts := ctx.Engine.Accounts().List()

	if len(accounts) == 0 {
		ctx.UI.Info("No accounts configured")
		ctx.UI.Hint("Use 'arngit account add <name>' to add an account")
		return nil
	}

	ctx.UI.Title("GitHub Accounts")
	current := ctx.Engine.Accounts().CurrentName()

	for _, name := range accounts {
		acc := ctx.Engine.Accounts().Get(name)
		marker := "  "
		if name == current {
			marker = "* "
		}
		ctx.UI.Info(fmt.Sprintf("%s%-12s %s <%s>", marker, name, acc.Username, acc.Email))
	}

	return nil
}

// handleAccountSwitch switches to a different account.
func (r *Router) handleAccountSwitch(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: arngit account switch <name>")
	}

	name := ctx.Args[0]

	if err := ctx.Engine.Accounts().Switch(name); err != nil {
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Switched to account '%s'", name))
	return nil
}

// handleAccountCurrent shows the current account.
func (r *Router) handleAccountCurrent(ctx *Context) error {
	acc := ctx.Engine.Accounts().Current()

	if acc == nil {
		ctx.UI.Info("No account selected")
		ctx.UI.Hint("Use 'arngit account add <name>' to add an account")
		return nil
	}

	ctx.UI.Title("Current Account")
	ctx.UI.Info(fmt.Sprintf("  Name:     %s", acc.Name))
	ctx.UI.Info(fmt.Sprintf("  Username: %s", acc.Username))
	ctx.UI.Info(fmt.Sprintf("  Email:    %s", acc.Email))

	return nil
}
