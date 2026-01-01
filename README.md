# arngit - Arinara Git Wrapper

A powerful git wrapper with multi-account management, auto-push, and repository protection.

## Features

- **Multi-Account Management** - Switch between multiple git accounts seamlessly
- **Auto-Push** - Automatically push changes based on configurable thresholds
- **Repository Protection** - Confirmation dialog for sensitive repositories  
- **Global Access** - Use from any directory after installation
- **Secure Credential Storage** - Uses Windows Credential Manager

## Installation

### From Source

```bash
# Clone or navigate to project directory
cd arngit

# Build
go build -o arngit.exe ./cmd/arngit

# Install (run as administrator if needed)
scripts\install.bat
```

### Manual Installation

1. Build: `go build -o arngit.exe ./cmd/arngit`
2. Copy `arngit.exe` to a directory in your PATH
3. Restart terminal

## Usage

### Account Management

```bash
# Add a new account
arngit account add personal
# Prompts for: username, email, PAT

# List all accounts (* = active)
arngit account list

# Switch active account
arngit account switch work

# Remove an account
arngit account remove old-account

# Show current account
arngit account current
```

### Push & Pull

```bash
# Push with protection check
arngit push

# Force push (bypass protection)
arngit push --force

# Push to specific remote/branch
arngit push -r origin -b main

# Pull from remote
arngit pull
```

### Repository Protection

Protected repositories show a confirmation dialog before push.

```bash
# Protect current repository
arngit protect

# Protect with reason
arngit protect --reason "Contains API keys"

# List protected repositories
arngit protect --list

# Remove protection
arngit unprotect
```

### Auto-Push Watcher

Automatically push when thresholds are reached.

```bash
# Start watcher with default settings
arngit watch

# Push after 5 commits
arngit watch --commits 5

# Push every 30 minutes
arngit watch --time 30m

# Push after 100KB of changes
arngit watch --size 100KB

# Stop watcher
arngit watch --stop
```

### Configuration

```bash
# Show all config
arngit config

# Set threshold type (commits, time, size)
arngit config set threshold.type commits
arngit config set threshold.value 5

# Set default branch
arngit config set default_branch main
```

### Repository Initialization

```bash
# Init new repo with active account settings
arngit init
```

## Configuration

Config is stored in `%APPDATA%\arngit\`:

- `config.json` - Global settings
- `accounts.json` - Account information
- `protected.json` - Protected repositories
- `watcher.log` - Watcher activity log

### Config Options

| Key | Default | Description |
|-----|---------|-------------|
| `threshold.type` | `commits` | Auto-push trigger: `commits`, `time`, or `size` |
| `threshold.value` | `5` | Threshold value |
| `watcher_interval` | `10s` | How often watcher checks for changes |
| `auto_init_remote` | `true` | Auto-setup remote on init |
| `default_branch` | `main` | Default branch name |
| `log_level` | `info` | Logging level |

## Security

- PAT (Personal Access Token) is stored in Windows Credential Manager
- Protected repositories require user confirmation before push
- Credentials are never stored in plain text

## Requirements

- Windows 10/11
- Git installed and in PATH
- Go 1.21+ (for building from source)

## License

MIT License

## Author

Arinara
