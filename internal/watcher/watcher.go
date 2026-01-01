// Package watcher provides auto-push functionality.
package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arinara/arngit/internal/account"
	"github.com/arinara/arngit/internal/config"
	"github.com/arinara/arngit/internal/dialog"
	"github.com/arinara/arngit/internal/git"
	"github.com/arinara/arngit/internal/protect"
)

// ThresholdType defines how auto-push is triggered.
type ThresholdType string

const (
	ThresholdCommits ThresholdType = "commits"
	ThresholdTime    ThresholdType = "time"
	ThresholdSize    ThresholdType = "size"
)

// Watcher monitors a repository for changes and auto-pushes.
type Watcher struct {
	RepoPath       string
	ThresholdType  ThresholdType
	ThresholdValue string
	Interval       time.Duration
	Remote         string
	Branch         string
	Logger         *log.Logger

	lastPushTime time.Time
	ctx          context.Context
	cancel       context.CancelFunc
}

// WatcherConfig holds watcher configuration.
type WatcherConfig struct {
	RepoPath       string
	ThresholdType  ThresholdType
	ThresholdValue string
	Interval       time.Duration
	Remote         string
	Branch         string
}

// NewWatcher creates a new repository watcher.
func NewWatcher(cfg WatcherConfig) (*Watcher, error) {
	g := git.New(cfg.RepoPath)
	if !g.IsRepo() {
		return nil, fmt.Errorf("'%s' is not a git repository", cfg.RepoPath)
	}

	// Get current branch if not specified
	if cfg.Branch == "" {
		branch, err := g.GetCurrentBranch()
		if err != nil {
			return nil, err
		}
		cfg.Branch = branch
	}

	if cfg.Remote == "" {
		cfg.Remote = "origin"
	}

	if cfg.Interval == 0 {
		cfg.Interval = 10 * time.Second
	}

	// Setup logger
	logPath, _ := getLogPath()
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logFile = os.Stdout
	}

	return &Watcher{
		RepoPath:       cfg.RepoPath,
		ThresholdType:  cfg.ThresholdType,
		ThresholdValue: cfg.ThresholdValue,
		Interval:       cfg.Interval,
		Remote:         cfg.Remote,
		Branch:         cfg.Branch,
		Logger:         log.New(logFile, "[watcher] ", log.LstdFlags),
		lastPushTime:   time.Now(),
	}, nil
}

// getLogPath returns the path to watcher.log.
func getLogPath() (string, error) {
	dir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "watcher.log"), nil
}

// Start begins watching the repository.
func (w *Watcher) Start() error {
	w.ctx, w.cancel = context.WithCancel(context.Background())

	w.Logger.Printf("Starting watcher for %s (threshold: %s %s, interval: %s)",
		w.RepoPath, w.ThresholdValue, w.ThresholdType, w.Interval)

	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.Logger.Println("Watcher stopped")
			return nil
		case <-ticker.C:
			if shouldPush, reason := w.checkThreshold(); shouldPush {
				w.Logger.Printf("Threshold reached: %s", reason)
				if err := w.push(); err != nil {
					w.Logger.Printf("Push failed: %v", err)
				}
			}
		}
	}
}

// Stop stops the watcher.
func (w *Watcher) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

// checkThreshold determines if a push should be triggered.
func (w *Watcher) checkThreshold() (bool, string) {
	g := git.New(w.RepoPath)

	switch w.ThresholdType {
	case ThresholdCommits:
		count, err := g.GetPendingCommitCount(w.Remote, w.Branch)
		if err != nil {
			return false, ""
		}
		threshold, _ := strconv.Atoi(w.ThresholdValue)
		if count >= threshold {
			return true, fmt.Sprintf("%d commits pending (threshold: %d)", count, threshold)
		}

	case ThresholdTime:
		duration, err := time.ParseDuration(w.ThresholdValue)
		if err != nil {
			return false, ""
		}
		if time.Since(w.lastPushTime) >= duration {
			count, _ := g.GetPendingCommitCount(w.Remote, w.Branch)
			if count > 0 {
				return true, fmt.Sprintf("time threshold reached (%s)", w.ThresholdValue)
			}
		}

	case ThresholdSize:
		size, err := g.GetUnpushedChangesSize(w.Remote, w.Branch)
		if err != nil {
			return false, ""
		}
		threshold := parseSize(w.ThresholdValue)
		if size >= threshold {
			return true, fmt.Sprintf("size threshold reached (%d bytes)", size)
		}
	}

	return false, ""
}

// parseSize converts size string to bytes.
func parseSize(s string) int64 {
	s = strings.ToUpper(strings.TrimSpace(s))
	multiplier := int64(1)

	if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	}

	value, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return value * multiplier
}

// push executes the push with protection check.
func (w *Watcher) push() error {
	g := git.New(w.RepoPath)

	// Check protection
	protStore, err := protect.LoadStore()
	if err != nil {
		return err
	}

	if prot, isProtected := protStore.GetProtection(w.RepoPath); isProtected {
		repoName, _ := g.GetRepoName()
		commitCount, _ := g.GetPendingCommitCount(w.Remote, w.Branch)

		// Show confirmation dialog
		if !dialog.ConfirmProtectedPush(repoName, w.Branch, prot.Reason, commitCount) {
			w.Logger.Println("Push cancelled by user (protected repository)")
			return nil
		}
	}

	// Get credentials
	accStore, err := account.LoadStore()
	if err != nil {
		return err
	}

	username, pat, err := accStore.GetActiveCredentials()
	if err != nil {
		return err
	}

	// Execute push
	if err := g.PushWithCredentials(w.Remote, w.Branch, username, pat); err != nil {
		return err
	}

	w.lastPushTime = time.Now()
	w.Logger.Printf("Successfully pushed to %s/%s", w.Remote, w.Branch)

	return nil
}

// WritePIDFile writes the watcher PID to a file.
func WritePIDFile(repoPath string, pid int) error {
	dir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	pidDir := filepath.Join(dir, "watchers")
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return err
	}

	// Use repo path hash as filename
	hash := hashPath(repoPath)
	pidFile := filepath.Join(pidDir, hash+".pid")

	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

// ReadPIDFile reads the watcher PID for a repository.
func ReadPIDFile(repoPath string) (int, error) {
	dir, err := config.GetConfigDir()
	if err != nil {
		return 0, err
	}

	hash := hashPath(repoPath)
	pidFile := filepath.Join(dir, "watchers", hash+".pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(data))
}

// RemovePIDFile removes the PID file for a repository.
func RemovePIDFile(repoPath string) error {
	dir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	hash := hashPath(repoPath)
	pidFile := filepath.Join(dir, "watchers", hash+".pid")

	return os.Remove(pidFile)
}

// hashPath creates a simple hash of a path for use as filename.
func hashPath(path string) string {
	absPath, _ := filepath.Abs(path)
	// Simple hash: replace problematic characters
	result := strings.ReplaceAll(absPath, "\\", "_")
	result = strings.ReplaceAll(result, ":", "_")
	result = strings.ReplaceAll(result, "/", "_")
	return result
}
