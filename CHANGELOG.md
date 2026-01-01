# Changelog

All notable changes to this project will be documented in this file.

## [2.2.0] - 2026-01-01

### Features

- **Stash Management** - `stash save/list/pop/apply/drop/clear`
- **Cherry-Pick** - `cherry-pick <commit>`
- **Revert** - `revert <commit>` with optional `--no-commit`
- **Squash** - `squash <count>` to squash last N commits
- **Diff** - `diff` with `--staged` and `--stat` options
- **Tag Management** - `tag list/create/delete`
- **Update** - `update` command (placeholder for self-update)
- **Doctor** - `doctor` health check for git and config

## [2.1.0] - 2026-01-01

### Features

- **Modern CLI UI** - Colored output, spinners, banners, styled tables
- **Automation**
  - `sync` - Fetch + rebase from remote
  - `hooks` - Git hooks management (install, uninstall, list)
- **Smart Features**
  - `changelog` - Auto-generate changelog from commits
  - `bump` - Semantic version bumping
  - `pr` - Pull request management via GitHub API
- **Analytics**
  - `stats` - Repository statistics and commit activity
  - `history` - Visual commit tree
  - `blame` - Enhanced file blame

## [2.0.0] - 2026-01-01

### Features

- **GitHub API Integration**
  - Repository management (`repo create`, `repo list`)
  - Release management (`release list`, `release create`, `release upload`)
  - Project and package management
- **Smart Commit** - Auto-generate commit messages based on changes
- **Status Dashboard** - Repository status with changes and commit info
- **Git Shortcuts**
  - `clone` - Clone with automatic account setup
  - `remote` - Remote management
  - `branch` - Branch management
- **PAT Validation** - Check and test GitHub PAT scopes

## [1.0.0] - 2026-01-01

### Features

- **Multi-Account Management** - Store and switch between multiple Git accounts
- **Secure Credential Storage** - PAT stored in Windows Credential Manager
- **Auto-Push Watcher** - Automatic push based on commit/time/size thresholds
- **Repository Protection** - Native dialog confirmation before push
- **Global CLI** - Access arngit from any directory
