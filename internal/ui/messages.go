package ui

import "fmt"

// Message templates for consistent user feedback.

// Messages provides developer-friendly feedback messages.
type Messages struct {
	renderer *Renderer
}

// NewMessages creates a new Messages instance.
func NewMessages(r *Renderer) *Messages {
	return &Messages{renderer: r}
}

// Git operation messages
func (m *Messages) GitInitSuccess(path string) {
	m.renderer.Success("Initialized empty Git repository")
	m.renderer.Hint("Next: Add files with 'add .' then 'commit -m \"Initial commit\"'")
}

func (m *Messages) GitCloneSuccess(repo, path string) {
	m.renderer.Success(fmt.Sprintf("Cloned %s", repo))
	m.renderer.Info(fmt.Sprintf("  → %s", path))
}

func (m *Messages) GitAddSuccess(count int) {
	if count == 1 {
		m.renderer.Success("Staged 1 file")
	} else {
		m.renderer.Success(fmt.Sprintf("Staged %d files", count))
	}
}

func (m *Messages) GitCommitSuccess(hash, message string) {
	shortHash := hash
	if len(hash) > 7 {
		shortHash = hash[:7]
	}
	m.renderer.Success(fmt.Sprintf("Committed: %s", shortHash))
	m.renderer.Info(fmt.Sprintf("  %s", message))
}

func (m *Messages) GitPushSuccess(branch, remote string) {
	m.renderer.Success(fmt.Sprintf("Pushed %s → %s", branch, remote))
}

func (m *Messages) GitPullSuccess(branch string, changes int) {
	if changes == 0 {
		m.renderer.Success("Already up to date")
	} else {
		m.renderer.Success(fmt.Sprintf("Pulled %d changes to %s", changes, branch))
	}
}

func (m *Messages) GitFetchSuccess(remote string) {
	m.renderer.Success(fmt.Sprintf("Fetched from %s", remote))
}

// Branch messages
func (m *Messages) BranchCreated(name string) {
	m.renderer.Success(fmt.Sprintf("Created branch: %s", name))
	m.renderer.Hint("Switch to it with 'branch switch " + name + "'")
}

func (m *Messages) BranchSwitched(name string) {
	m.renderer.Success(fmt.Sprintf("Switched to: %s", name))
}

func (m *Messages) BranchDeleted(name string) {
	m.renderer.Success(fmt.Sprintf("Deleted branch: %s", name))
}

// Account messages
func (m *Messages) AccountAdded(name string) {
	m.renderer.Success(fmt.Sprintf("Account '%s' added", name))
	m.renderer.Hint("Switch to it with 'account switch " + name + "'")
}

func (m *Messages) AccountSwitched(name, username string) {
	m.renderer.Success(fmt.Sprintf("Now using: %s (%s)", name, username))
}

func (m *Messages) AccountRemoved(name string) {
	m.renderer.Success(fmt.Sprintf("Account '%s' removed", name))
}

// Config messages
func (m *Messages) ConfigUpdated(key string, value interface{}) {
	m.renderer.Success(fmt.Sprintf("Set %s = %v", key, value))
}

// Protected repo messages
func (m *Messages) RepoProtected() {
	m.renderer.Success("Repository protected")
	m.renderer.Hint("Push operations will require confirmation")
}

func (m *Messages) RepoUnprotected() {
	m.renderer.Success("Protection removed")
}

func (m *Messages) ProtectedPushWarning() {
	m.renderer.Warning("This repository is protected")
	m.renderer.Info("  Pushing requires explicit confirmation")
}

// Stash messages
func (m *Messages) StashSaved(message string) {
	m.renderer.Success(fmt.Sprintf("Stashed: %s", message))
}

func (m *Messages) StashPopped() {
	m.renderer.Success("Stash applied and removed")
}

// Tag messages
func (m *Messages) TagCreated(name string) {
	m.renderer.Success(fmt.Sprintf("Created tag: %s", name))
	m.renderer.Hint("Push with 'push --tags'")
}

func (m *Messages) TagDeleted(name string) {
	m.renderer.Success(fmt.Sprintf("Deleted tag: %s", name))
}

// Error messages with hints
func (m *Messages) NotAGitRepo() {
	m.renderer.Error("Not a git repository")
	m.renderer.Hint("Initialize with 'init' or clone an existing repo")
}

func (m *Messages) NoAccountConfigured() {
	m.renderer.Warning("No account configured")
	m.renderer.Hint("Add one with 'account add <name>'")
}

func (m *Messages) NothingToCommit() {
	m.renderer.Info("Nothing to commit, working tree clean")
}

func (m *Messages) NoChanges() {
	m.renderer.Info("No changes detected")
}

func (m *Messages) NothingStaged() {
	m.renderer.Warning("Nothing staged for commit")
	m.renderer.Hint("Stage files with 'add <file>' or 'add .'")
}

func (m *Messages) CommitMessageRequired() {
	m.renderer.Error("Commit message required")
	m.renderer.Hint("Use 'commit -m \"your message\"'")
}

func (m *Messages) PushFailed(reason string) {
	m.renderer.Error("Push failed")
	m.renderer.Info("  " + reason)
	m.renderer.Hint("Try 'pull' first to sync with remote")
}

func (m *Messages) PullConflict() {
	m.renderer.Warning("Conflicts detected during pull")
	m.renderer.Hint("Resolve conflicts, then 'add' and 'commit'")
}

func (m *Messages) AuthenticationFailed() {
	m.renderer.Error("Authentication failed")
	m.renderer.Hint("Check your PAT with 'account current' or add new account")
}

// General messages
func (m *Messages) OperationCancelled() {
	m.renderer.Info("Operation cancelled")
}

func (m *Messages) Goodbye() {
	m.renderer.Info("Goodbye!")
}

func (m *Messages) CommandNotFound(cmd string) {
	m.renderer.Error(fmt.Sprintf("Unknown command: %s", cmd))
	m.renderer.Hint("Type 'help' for available commands")
}
