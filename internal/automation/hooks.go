// Package automation - Git hooks management
package automation

import (
	"os"
	"path/filepath"
	"strings"
)

// Hook types
const (
	HookPreCommit  = "pre-commit"
	HookPostCommit = "post-commit"
	HookPrePush    = "pre-push"
	HookCommitMsg  = "commit-msg"
)

// AvailableHooks returns all available hook types.
func AvailableHooks() []string {
	return []string{
		HookPreCommit,
		HookPostCommit,
		HookPrePush,
		HookCommitMsg,
	}
}

// HookTemplate returns a template for a hook type.
func HookTemplate(hookType string) string {
	templates := map[string]string{
		HookPreCommit: `#!/bin/sh
# arngit pre-commit hook
# Run tests and linting before commit

echo "Running pre-commit checks..."

# Check for TODO/FIXME comments (warning only)
if git diff --cached --name-only | xargs grep -l "TODO\|FIXME" 2>/dev/null; then
    echo "⚠ Warning: Found TODO/FIXME comments"
fi

# Check for large files
MAX_SIZE=1048576  # 1MB
for file in $(git diff --cached --name-only); do
    if [ -f "$file" ]; then
        size=$(wc -c < "$file")
        if [ $size -gt $MAX_SIZE ]; then
            echo "✗ Error: $file is larger than 1MB"
            exit 1
        fi
    fi
done

echo "✓ Pre-commit checks passed"
exit 0
`,
		HookCommitMsg: `#!/bin/sh
# arngit commit-msg hook
# Validate commit message format

MSG_FILE=$1
MSG=$(cat "$MSG_FILE")

# Check minimum length
if [ ${#MSG} -lt 10 ]; then
    echo "✗ Commit message too short (min 10 chars)"
    exit 1
fi

# Check for type prefix (feat:, fix:, chore:, etc.)
if ! echo "$MSG" | grep -qE "^(feat|fix|chore|docs|style|refactor|test|ci|build):"; then
    echo "⚠ Consider using conventional commit format: type: message"
fi

exit 0
`,
		HookPrePush: `#!/bin/sh
# arngit pre-push hook
# Run checks before pushing

REMOTE=$1
URL=$2

echo "Pushing to $REMOTE ($URL)..."

# You can add custom checks here
# For example: run tests, check branch name, etc.

exit 0
`,
		HookPostCommit: `#!/bin/sh
# arngit post-commit hook
# Actions after commit

echo "✓ Commit created: $(git rev-parse --short HEAD)"
`,
	}

	if tmpl, ok := templates[hookType]; ok {
		return tmpl
	}
	return "#!/bin/sh\n# Custom hook\nexit 0\n"
}

// InstallHook installs a git hook.
func InstallHook(repoPath, hookType string) error {
	hooksDir := filepath.Join(repoPath, ".git", "hooks")

	// Create hooks directory if not exists
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, hookType)
	template := HookTemplate(hookType)

	// Write hook file
	if err := os.WriteFile(hookPath, []byte(template), 0755); err != nil {
		return err
	}

	return nil
}

// UninstallHook removes a git hook.
func UninstallHook(repoPath, hookType string) error {
	hookPath := filepath.Join(repoPath, ".git", "hooks", hookType)
	return os.Remove(hookPath)
}

// ListInstalledHooks returns installed hooks.
func ListInstalledHooks(repoPath string) ([]string, error) {
	hooksDir := filepath.Join(repoPath, ".git", "hooks")

	var installed []string
	for _, hook := range AvailableHooks() {
		hookPath := filepath.Join(hooksDir, hook)
		if _, err := os.Stat(hookPath); err == nil {
			installed = append(installed, hook)
		}
	}

	return installed, nil
}

// IsHookInstalled checks if a hook is installed.
func IsHookInstalled(repoPath, hookType string) bool {
	hookPath := filepath.Join(repoPath, ".git", "hooks", hookType)
	_, err := os.Stat(hookPath)
	return err == nil
}

// GetHookContent returns the content of a hook.
func GetHookContent(repoPath, hookType string) (string, error) {
	hookPath := filepath.Join(repoPath, ".git", "hooks", hookType)
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// HookDescription returns a description for a hook type.
func HookDescription(hookType string) string {
	descriptions := map[string]string{
		HookPreCommit:  "Runs before commit is created (validate, lint, test)",
		HookPostCommit: "Runs after commit is created (notifications, cleanup)",
		HookPrePush:    "Runs before push (validate, test, check branch)",
		HookCommitMsg:  "Validates commit message format",
	}

	if desc, ok := descriptions[hookType]; ok {
		return desc
	}
	return "Custom hook"
}

// ValidateHookType checks if hook type is valid.
func ValidateHookType(hookType string) bool {
	hookType = strings.ToLower(hookType)
	for _, h := range AvailableHooks() {
		if h == hookType {
			return true
		}
	}
	return false
}
