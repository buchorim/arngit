package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/arfrfrr/arngit/internal/github"
	"github.com/arfrfrr/arngit/internal/ui"
)

// RegisterGitHubCommands registers GitHub API commands.
func (r *Router) RegisterGitHubCommands() {
	// repo command
	r.Register(&Command{
		Name:        "repo",
		Description: "Manage GitHub repositories",
		Usage:       "arngit repo <subcommand>",
		Handler:     r.handleRepoHelp,
		SubCommands: map[string]*Command{
			"create": {
				Name:        "create",
				Description: "Create a new repository",
				Usage:       "arngit repo create <name> [-p]",
				Handler:     r.handleRepoCreate,
			},
			"list": {
				Name:        "list",
				Description: "List your repositories",
				Usage:       "arngit repo list",
				Handler:     r.handleRepoList,
			},
			"delete": {
				Name:        "delete",
				Description: "Delete a repository",
				Usage:       "arngit repo delete <owner/repo>",
				Handler:     r.handleRepoDelete,
			},
		},
	})

	// release command
	r.Register(&Command{
		Name:        "release",
		Description: "Manage GitHub releases",
		Usage:       "arngit release <subcommand>",
		Handler:     r.handleReleaseHelp,
		SubCommands: map[string]*Command{
			"list": {
				Name:        "list",
				Description: "List releases",
				Usage:       "arngit release list <owner/repo>",
				Handler:     r.handleReleaseList,
			},
			"create": {
				Name:        "create",
				Description: "Create a new release",
				Usage:       "arngit release create <owner/repo> <tag>",
				Handler:     r.handleReleaseCreate,
			},
			"upload": {
				Name:        "upload",
				Description: "Upload asset to a release",
				Usage:       "arngit release upload <owner/repo> <tag> <file>",
				Handler:     r.handleReleaseUpload,
			},
		},
	})

	// pr command
	r.Register(&Command{
		Name:        "pr",
		Description: "Manage pull requests",
		Usage:       "arngit pr <subcommand>",
		Handler:     r.handlePRHelp,
		SubCommands: map[string]*Command{
			"create": {
				Name:        "create",
				Description: "Create a pull request",
				Usage:       "arngit pr create <owner/repo> <title>",
				Handler:     r.handlePRCreate,
			},
			"list": {
				Name:        "list",
				Description: "List pull requests",
				Usage:       "arngit pr list <owner/repo>",
				Handler:     r.handlePRList,
			},
		},
	})
}

// getGitHubClient creates a GitHub client from the current account.
func (r *Router) getGitHubClient() (*github.Client, error) {
	acc := r.engine.Accounts().Current()
	if acc == nil {
		return nil, fmt.Errorf("no account configured. Use 'arngit account add <name>'")
	}

	pat, err := r.engine.Accounts().GetPAT(acc.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get PAT: %w", err)
	}

	return github.NewClient(acc.Username, pat), nil
}

// parseOwnerRepo splits "owner/repo" string.
func parseOwnerRepo(s string) (string, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format, expected owner/repo")
	}
	return parts[0], parts[1], nil
}

// --- Repo handlers ---

func (r *Router) handleRepoHelp(ctx *Context) error {
	ctx.UI.Title("Repository Management")
	fmt.Println(r.ui.Color(ui.Dim, "  Manage your GitHub repositories"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "create <name> [-p]"), r.ui.Color(ui.Dim, "Create a new repository (-p for private)"))
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "list            "), r.ui.Color(ui.Dim, "List your repositories"))
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "delete <o/r>    "), r.ui.Color(ui.Dim, "Delete a repository"))
	fmt.Println()
	return nil
}

func (r *Router) handleRepoCreate(ctx *Context) error {
	if len(ctx.Args) < 1 {
		ctx.UI.Error("Repository name required")
		ctx.UI.Hint("Usage: arngit repo create <name> [-p]")
		return fmt.Errorf("missing repo name")
	}

	client, err := r.getGitHubClient()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	name := ctx.Args[0]
	private := false
	for _, arg := range ctx.Args[1:] {
		if arg == "-p" || arg == "--private" {
			private = true
		}
	}

	repo, err := client.CreateRepo(github.CreateRepoParams{
		Name:     name,
		Private:  private,
		AutoInit: true,
	})
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to create repository: %v", err))
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Created repository: %s", repo.FullName))
	ctx.UI.Hint(fmt.Sprintf("Clone: git clone %s", repo.CloneURL))
	return nil
}

func (r *Router) handleRepoList(ctx *Context) error {
	client, err := r.getGitHubClient()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	repos, err := client.ListUserRepos()
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to list repositories: %v", err))
		return err
	}

	if len(repos) == 0 {
		ctx.UI.Info("No repositories found")
		return nil
	}

	ctx.UI.Title("Your Repositories")
	headers := []string{"Name", "Visibility", "Language", "Stars"}
	var rows [][]string
	for _, repo := range repos {
		vis := "public"
		if repo.Private {
			vis = "private"
		}
		lang := repo.Language
		if lang == "" {
			lang = "-"
		}
		rows = append(rows, []string{
			repo.FullName,
			vis,
			lang,
			fmt.Sprintf("%d", repo.StargazersCount),
		})
	}
	ctx.UI.Table(headers, rows)
	fmt.Println()
	return nil
}

func (r *Router) handleRepoDelete(ctx *Context) error {
	if len(ctx.Args) < 1 {
		ctx.UI.Error("Repository required (owner/repo)")
		return fmt.Errorf("missing repo")
	}

	owner, repo, err := parseOwnerRepo(ctx.Args[0])
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if !ctx.UI.Confirm(fmt.Sprintf("Delete %s/%s? This cannot be undone", owner, repo)) {
		ctx.UI.Info("Cancelled")
		return nil
	}

	client, err := r.getGitHubClient()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	if err := client.DeleteRepo(owner, repo); err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to delete: %v", err))
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Deleted repository: %s/%s", owner, repo))
	return nil
}

// --- Release handlers ---

func (r *Router) handleReleaseHelp(ctx *Context) error {
	ctx.UI.Title("Release Management")
	fmt.Println(r.ui.Color(ui.Dim, "  Manage GitHub releases"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "list <o/r>             "), r.ui.Color(ui.Dim, "List releases"))
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "create <o/r> <tag>     "), r.ui.Color(ui.Dim, "Create a release"))
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "upload <o/r> <tag> <f> "), r.ui.Color(ui.Dim, "Upload asset"))
	fmt.Println()
	return nil
}

func (r *Router) handleReleaseList(ctx *Context) error {
	if len(ctx.Args) < 1 {
		ctx.UI.Error("Repository required (owner/repo)")
		return fmt.Errorf("missing repo")
	}

	owner, repo, err := parseOwnerRepo(ctx.Args[0])
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	client, err := r.getGitHubClient()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	releases, err := client.ListReleases(owner, repo)
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to list releases: %v", err))
		return err
	}

	if len(releases) == 0 {
		ctx.UI.Info("No releases found")
		return nil
	}

	ctx.UI.Title("Releases")
	headers := []string{"Tag", "Name", "Published", "Assets"}
	var rows [][]string
	for _, rel := range releases {
		rows = append(rows, []string{
			rel.TagName,
			rel.Name,
			rel.PublishedAt,
			fmt.Sprintf("%d", len(rel.Assets)),
		})
	}
	ctx.UI.Table(headers, rows)
	fmt.Println()
	return nil
}

func (r *Router) handleReleaseCreate(ctx *Context) error {
	if len(ctx.Args) < 2 {
		ctx.UI.Error("Usage: arngit release create <owner/repo> <tag>")
		return fmt.Errorf("missing args")
	}

	owner, repo, err := parseOwnerRepo(ctx.Args[0])
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}
	tag := ctx.Args[1]

	client, err := r.getGitHubClient()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	release, err := client.CreateRelease(owner, repo, github.CreateReleaseParams{
		TagName: tag,
		Name:    tag,
	})
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to create release: %v", err))
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Created release: %s", release.TagName))
	ctx.UI.Hint(release.HTMLURL)
	return nil
}

func (r *Router) handleReleaseUpload(ctx *Context) error {
	if len(ctx.Args) < 3 {
		ctx.UI.Error("Usage: arngit release upload <owner/repo> <tag> <file>")
		return fmt.Errorf("missing args")
	}

	owner, repo, err := parseOwnerRepo(ctx.Args[0])
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}
	tag := ctx.Args[1]
	filePath := ctx.Args[2]

	data, err := os.ReadFile(filePath)
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to read file: %v", err))
		return err
	}

	client, err := r.getGitHubClient()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	// Get the release to get upload URL
	release, err := client.GetLatestRelease(owner, repo)
	if err != nil || release.TagName != tag {
		ctx.UI.Error(fmt.Sprintf("Release '%s' not found", tag))
		return fmt.Errorf("release not found")
	}

	filename := filePath
	if idx := strings.LastIndex(filePath, string(os.PathSeparator)); idx >= 0 {
		filename = filePath[idx+1:]
	}

	asset, err := client.UploadReleaseAsset(release.UploadURL, filename, "application/octet-stream", data)
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to upload: %v", err))
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Uploaded: %s (%d bytes)", asset.Name, asset.Size))
	return nil
}

// --- PR handlers ---

func (r *Router) handlePRHelp(ctx *Context) error {
	ctx.UI.Title("Pull Request Management")
	fmt.Println(r.ui.Color(ui.Dim, "  Manage pull requests"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "create <o/r> <title>"), r.ui.Color(ui.Dim, "Create a PR from current branch"))
	fmt.Printf("  %s  %s\n", r.ui.Color(ui.Yellow, "list <o/r>          "), r.ui.Color(ui.Dim, "List open PRs"))
	fmt.Println()
	return nil
}

func (r *Router) handlePRCreate(ctx *Context) error {
	if len(ctx.Args) < 2 {
		ctx.UI.Error("Usage: arngit pr create <owner/repo> <title>")
		return fmt.Errorf("missing args")
	}

	owner, repo, err := parseOwnerRepo(ctx.Args[0])
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}
	title := strings.Join(ctx.Args[1:], " ")

	// Get current branch
	gitSvc := r.getGitService()
	branch, err := gitSvc.CurrentBranch()
	if err != nil {
		ctx.UI.Error("Failed to get current branch")
		return err
	}

	client, err := r.getGitHubClient()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	pr, err := client.CreatePR(owner, repo, github.CreatePRParams{
		Title: title,
		Head:  branch,
		Base:  "main",
	})
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to create PR: %v", err))
		return err
	}

	ctx.UI.Success(fmt.Sprintf("Created PR #%d: %s", pr.Number, pr.Title))
	ctx.UI.Hint(pr.HTMLURL)
	return nil
}

func (r *Router) handlePRList(ctx *Context) error {
	if len(ctx.Args) < 1 {
		ctx.UI.Error("Repository required (owner/repo)")
		return fmt.Errorf("missing repo")
	}

	owner, repo, err := parseOwnerRepo(ctx.Args[0])
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	client, err := r.getGitHubClient()
	if err != nil {
		ctx.UI.Error(err.Error())
		return err
	}

	prs, err := client.ListPRs(owner, repo, "open")
	if err != nil {
		ctx.UI.Error(fmt.Sprintf("Failed to list PRs: %v", err))
		return err
	}

	if len(prs) == 0 {
		ctx.UI.Info("No open pull requests")
		return nil
	}

	ctx.UI.Title("Open Pull Requests")
	headers := []string{"#", "Title", "Author", "Branch"}
	var rows [][]string
	for _, pr := range prs {
		rows = append(rows, []string{
			fmt.Sprintf("%d", pr.Number),
			pr.Title,
			pr.User.Login,
			pr.Head.Ref,
		})
	}
	ctx.UI.Table(headers, rows)
	fmt.Println()
	return nil
}
