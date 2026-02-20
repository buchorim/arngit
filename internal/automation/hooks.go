// Package automation provides git hooks management.
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

// HookTemplate returns a template for a hook type.
func HookTemplate(hookType string) string {
	templates := map[string]string{
		HookPreCommit: `#!/bin/sh
# arngit pre-commit hook
echo "Running pre-commit checks..."

# Check for large files
MAX_SIZE=1048576
for file in $(git diff --cached --name-only); do
    if [ -f "$file" ]; then
        size=$(wc -c < "$file")
        if [ $size -gt $MAX_SIZE ]; then
            echo "Error: $file is larger than 1MB"
            exit 1
        fi
    fi
done

echo "Pre-commit checks passed"
exit 0
`,
		HookCommitMsg: `#!/bin/sh
# arngit commit-msg hook
MSG_FILE=$1
MSG=$(cat "$MSG_FILE")

if [ ${#MSG} -lt 10 ]; then
    echo "Commit message too short (min 10 chars)"
    exit 1
fi

exit 0
`,
		HookPrePush: `#!/bin/sh
# arngit pre-push hook
REMOTE=$1
URL=$2
echo "Pushing to $REMOTE ($URL)..."
exit 0
`,
		HookPostCommit: `#!/bin/sh
# arngit post-commit hook
echo "Commit created: $(git rev-parse --short HEAD)"
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

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, hookType)
	template := HookTemplate(hookType)

	return os.WriteFile(hookPath, []byte(template), 0755)
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
