package core

import "fmt"

// ErrorCode represents a unique error identifier.
type ErrorCode string

// Error codes
const (
	ErrConfigLoad      ErrorCode = "CONFIG_LOAD"
	ErrConfigSave      ErrorCode = "CONFIG_SAVE"
	ErrAccountNotFound ErrorCode = "ACCOUNT_NOT_FOUND"
	ErrAccountExists   ErrorCode = "ACCOUNT_EXISTS"
	ErrPATEncrypt      ErrorCode = "PAT_ENCRYPT"
	ErrPATDecrypt      ErrorCode = "PAT_DECRYPT"
	ErrStorageInit     ErrorCode = "STORAGE_INIT"
	ErrGitNotFound     ErrorCode = "GIT_NOT_FOUND"
	ErrGitCommand      ErrorCode = "GIT_COMMAND"
	ErrGitNoRepo       ErrorCode = "GIT_NO_REPO"
	ErrGitRemote       ErrorCode = "GIT_REMOTE"
	ErrGitAuth         ErrorCode = "GIT_AUTH"
	ErrGitConflict     ErrorCode = "GIT_CONFLICT"
	ErrGitNoChanges    ErrorCode = "GIT_NO_CHANGES"
	ErrGitProtected    ErrorCode = "GIT_PROTECTED"
	ErrGitHook         ErrorCode = "GIT_HOOK"
	ErrGitBranch       ErrorCode = "GIT_BRANCH"
	ErrGitRebase       ErrorCode = "GIT_REBASE"
	ErrGitMerge        ErrorCode = "GIT_MERGE"
	ErrGitStash        ErrorCode = "GIT_STASH"
	ErrGitTag          ErrorCode = "GIT_TAG"
	ErrGitClone        ErrorCode = "GIT_CLONE"
	ErrGitFetch        ErrorCode = "GIT_FETCH"
	ErrGitPush         ErrorCode = "GIT_PUSH"
	ErrGitPull         ErrorCode = "GIT_PULL"
	ErrGitCommit       ErrorCode = "GIT_COMMIT"
	ErrGitStatus       ErrorCode = "GIT_STATUS"
	ErrGitDiff         ErrorCode = "GIT_DIFF"
	ErrGitLog          ErrorCode = "GIT_LOG"
	ErrAPIRequest      ErrorCode = "API_REQUEST"
	ErrAPIAuth         ErrorCode = "API_AUTH"
	ErrAPIRateLimit    ErrorCode = "API_RATE_LIMIT"
	ErrAPINotFound     ErrorCode = "API_NOT_FOUND"
	ErrPluginLoad      ErrorCode = "PLUGIN_LOAD"
	ErrPluginCrash     ErrorCode = "PLUGIN_CRASH"
	ErrUpdateCheck     ErrorCode = "UPDATE_CHECK"
	ErrUpdateApply     ErrorCode = "UPDATE_APPLY"
)

// ErrorInfo contains metadata about an error.
type ErrorInfo struct {
	Code    ErrorCode
	Message string
	Hint    string
}

// errorRegistry maps error codes to their info.
var errorRegistry = map[ErrorCode]ErrorInfo{
	ErrConfigLoad: {
		Code:    ErrConfigLoad,
		Message: "Failed to load configuration",
		Hint:    "Check if ~/.arngit/config/config.yaml exists and is valid YAML",
	},
	ErrConfigSave: {
		Code:    ErrConfigSave,
		Message: "Failed to save configuration",
		Hint:    "Check write permissions for ~/.arngit/config/",
	},
	ErrAccountNotFound: {
		Code:    ErrAccountNotFound,
		Message: "Account not found",
		Hint:    "Use 'arngit account list' to see available accounts",
	},
	ErrAccountExists: {
		Code:    ErrAccountExists,
		Message: "Account already exists",
		Hint:    "Use a different name or remove the existing account first",
	},
	ErrPATEncrypt: {
		Code:    ErrPATEncrypt,
		Message: "Failed to encrypt PAT",
		Hint:    "This is an internal error. Please report this issue.",
	},
	ErrPATDecrypt: {
		Code:    ErrPATDecrypt,
		Message: "Failed to decrypt PAT",
		Hint:    "The PAT may have been encrypted on a different machine. Re-add the account.",
	},
	ErrStorageInit: {
		Code:    ErrStorageInit,
		Message: "Failed to initialize storage",
		Hint:    "Check write permissions for home directory",
	},
	ErrGitNotFound: {
		Code:    ErrGitNotFound,
		Message: "Git is not installed or not in PATH",
		Hint:    "Install Git from https://git-scm.com/",
	},
	ErrGitNoRepo: {
		Code:    ErrGitNoRepo,
		Message: "Not a git repository",
		Hint:    "Run 'arngit init' to initialize a repository, or navigate to an existing repo",
	},
	ErrGitAuth: {
		Code:    ErrGitAuth,
		Message: "Git authentication failed",
		Hint:    "Check your PAT permissions. Ensure 'repo' scope is enabled.",
	},
	ErrGitConflict: {
		Code:    ErrGitConflict,
		Message: "Merge conflict detected",
		Hint:    "Resolve conflicts manually, then run 'arngit commit'",
	},
	ErrGitNoChanges: {
		Code:    ErrGitNoChanges,
		Message: "No changes to commit",
		Hint:    "Make some changes first, or use --allow-empty",
	},
	ErrGitProtected: {
		Code:    ErrGitProtected,
		Message: "Repository is protected",
		Hint:    "Confirm the operation when prompted, or unprotect the repo first",
	},
	ErrAPIRateLimit: {
		Code:    ErrAPIRateLimit,
		Message: "GitHub API rate limit exceeded",
		Hint:    "Wait a few minutes before retrying, or use a PAT for higher limits",
	},
	ErrAPINotFound: {
		Code:    ErrAPINotFound,
		Message: "Resource not found on GitHub",
		Hint:    "Check the repository/resource name and your access permissions",
	},
	ErrPluginLoad: {
		Code:    ErrPluginLoad,
		Message: "Failed to load plugin",
		Hint:    "Check if the plugin is compatible with this version of ArnGit",
	},
	ErrPluginCrash: {
		Code:    ErrPluginCrash,
		Message: "Plugin crashed",
		Hint:    "The plugin has been disabled. Report this to the plugin author.",
	},
	ErrUpdateCheck: {
		Code:    ErrUpdateCheck,
		Message: "Failed to check for updates",
		Hint:    "Check your internet connection and try again later",
	},
	ErrUpdateApply: {
		Code:    ErrUpdateApply,
		Message: "Failed to apply update",
		Hint:    "Try downloading the update manually from GitHub releases",
	},
}

// AppError represents an application error with context.
type AppError struct {
	Code    ErrorCode
	Message string
	Hint    string
	Cause   error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewError creates a new AppError from an error code.
func NewError(code ErrorCode, cause error) *AppError {
	info, ok := errorRegistry[code]
	if !ok {
		info = ErrorInfo{
			Code:    code,
			Message: string(code),
			Hint:    "No additional information available",
		}
	}

	return &AppError{
		Code:    code,
		Message: info.Message,
		Hint:    info.Hint,
		Cause:   cause,
	}
}

// NewErrorf creates a new AppError with a formatted message.
func NewErrorf(code ErrorCode, format string, args ...interface{}) *AppError {
	info, ok := errorRegistry[code]
	if !ok {
		info = ErrorInfo{
			Code: code,
			Hint: "No additional information available",
		}
	}

	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Hint:    info.Hint,
	}
}

// GetErrorHint returns the hint for an error code.
func GetErrorHint(code ErrorCode) string {
	if info, ok := errorRegistry[code]; ok {
		return info.Hint
	}
	return ""
}
