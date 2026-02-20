// Package automation - File system watcher with threshold-based auto-push.
package automation

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/arfrfrr/arngit/internal/core"
	"github.com/arfrfrr/arngit/internal/git"
)

// ThresholdType defines how auto-push is triggered.
type ThresholdType string

const (
	ThresholdCommits ThresholdType = "commits"
	ThresholdTime    ThresholdType = "time"
	ThresholdSize    ThresholdType = "size"
)

// Watcher monitors a repository and auto-pushes based on thresholds.
type Watcher struct {
	gitSvc         *git.Service
	engine         *core.Engine
	thresholdType  ThresholdType
	thresholdValue string
	interval       time.Duration
	remote         string
	branch         string

	lastPushTime time.Time
	ctx          context.Context
	cancel       context.CancelFunc

	// Callbacks
	OnPush    func(remote, branch string, commitCount int)
	OnError   func(err error)
	OnCheck   func(thresholdType ThresholdType, current, threshold string)
	OnSkipped func(reason string)
}

// WatcherConfig holds watcher configuration.
type WatcherConfig struct {
	ThresholdType  ThresholdType
	ThresholdValue string
	Interval       time.Duration
	Remote         string
	Branch         string
}

// NewWatcher creates a new repository watcher.
func NewWatcher(engine *core.Engine, gitSvc *git.Service, cfg WatcherConfig) (*Watcher, error) {
	if !gitSvc.IsRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	// Default branch from current
	if cfg.Branch == "" {
		branch, err := gitSvc.CurrentBranch()
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

	if cfg.ThresholdType == "" {
		cfg.ThresholdType = ThresholdCommits
	}

	if cfg.ThresholdValue == "" {
		switch cfg.ThresholdType {
		case ThresholdCommits:
			cfg.ThresholdValue = "3"
		case ThresholdTime:
			cfg.ThresholdValue = "5m"
		case ThresholdSize:
			cfg.ThresholdValue = "1MB"
		}
	}

	return &Watcher{
		gitSvc:         gitSvc,
		engine:         engine,
		thresholdType:  cfg.ThresholdType,
		thresholdValue: cfg.ThresholdValue,
		interval:       cfg.Interval,
		remote:         cfg.Remote,
		branch:         cfg.Branch,
		lastPushTime:   time.Now(),
	}, nil
}

// Start begins watching the repository. Blocks until Stop() is called.
func (w *Watcher) Start() error {
	w.ctx, w.cancel = context.WithCancel(context.Background())

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return nil
		case <-ticker.C:
			shouldPush, reason := w.checkThreshold()
			if shouldPush {
				count, _ := w.gitSvc.PendingCommitCount(w.remote, w.branch)
				if err := w.push(); err != nil {
					if w.OnError != nil {
						w.OnError(err)
					}
				} else {
					if w.OnPush != nil {
						w.OnPush(w.remote, w.branch, count)
					}
				}
			} else if reason != "" {
				if w.OnSkipped != nil {
					w.OnSkipped(reason)
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

// Config returns the watcher configuration summary.
func (w *Watcher) Config() string {
	return fmt.Sprintf("threshold=%s:%s interval=%s remote=%s branch=%s",
		w.thresholdType, w.thresholdValue, w.interval, w.remote, w.branch)
}

// checkThreshold determines if a push should be triggered.
func (w *Watcher) checkThreshold() (bool, string) {
	switch w.thresholdType {
	case ThresholdCommits:
		count, err := w.gitSvc.PendingCommitCount(w.remote, w.branch)
		if err != nil {
			return false, ""
		}
		threshold, _ := strconv.Atoi(w.thresholdValue)
		if w.OnCheck != nil {
			w.OnCheck(ThresholdCommits, strconv.Itoa(count), w.thresholdValue)
		}
		if count >= threshold {
			return true, fmt.Sprintf("%d commits pending (threshold: %d)", count, threshold)
		}
		if count > 0 {
			return false, fmt.Sprintf("%d/%d commits", count, threshold)
		}

	case ThresholdTime:
		duration, err := time.ParseDuration(w.thresholdValue)
		if err != nil {
			return false, ""
		}
		elapsed := time.Since(w.lastPushTime)
		if w.OnCheck != nil {
			w.OnCheck(ThresholdTime, elapsed.Round(time.Second).String(), duration.String())
		}
		if elapsed >= duration {
			count, _ := w.gitSvc.PendingCommitCount(w.remote, w.branch)
			if count > 0 {
				return true, fmt.Sprintf("time threshold reached (%s)", w.thresholdValue)
			}
		}

	case ThresholdSize:
		size, err := w.gitSvc.UnpushedChangesSize(w.remote, w.branch)
		if err != nil {
			return false, ""
		}
		threshold := ParseSize(w.thresholdValue)
		if w.OnCheck != nil {
			w.OnCheck(ThresholdSize, FormatSize(size), FormatSize(threshold))
		}
		if size >= threshold {
			return true, fmt.Sprintf("size threshold reached (%s)", FormatSize(size))
		}
	}

	return false, ""
}

// push executes the push.
func (w *Watcher) push() error {
	// Check protection
	repoName, _ := w.gitSvc.RepoName()
	if w.engine.ProtectedRepos().IsProtected(repoName) {
		return fmt.Errorf("repository '%s' is protected - skipping auto-push", repoName)
	}

	// Execute push
	if err := w.gitSvc.Push(w.remote, w.branch, false, false); err != nil {
		return err
	}

	w.lastPushTime = time.Now()
	return nil
}

// ParseSize converts size string to bytes.
func ParseSize(s string) int64 {
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

// FormatSize formats bytes to human-readable.
func FormatSize(bytes int64) string {
	if bytes >= 1024*1024*1024 {
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
	}
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	if bytes >= 1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%d B", bytes)
}
