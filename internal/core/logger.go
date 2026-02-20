package core

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// LogLevel represents log severity.
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// LogEntry represents a single log entry.
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   LogLevel  `json:"level"`
	Message string    `json:"message"`
}

// LevelString returns the string representation of log level.
func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarn:
		return "WARN"
	case LogError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides application logging with in-memory buffer.
type Logger struct {
	file     *os.File
	entries  []LogEntry
	maxSize  int
	mu       sync.RWMutex
	minLevel LogLevel
}

// NewLogger creates a new logger.
func NewLogger(path string, maxEntries int) (*Logger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:     file,
		entries:  make([]LogEntry, 0, maxEntries),
		maxSize:  maxEntries,
		minLevel: LogInfo,
	}, nil
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// log writes a log entry.
func (l *Logger) log(level LogLevel, message string) {
	if level < l.minLevel {
		return
	}

	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Add to in-memory buffer
	if len(l.entries) >= l.maxSize {
		// Remove oldest entries
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, entry)

	// Write to file
	if l.file != nil {
		line := fmt.Sprintf("[%s] %s: %s\n",
			entry.Time.Format("2006-01-02 15:04:05"),
			entry.Level.String(),
			entry.Message)
		l.file.WriteString(line)
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(message string) {
	l.log(LogDebug, message)
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LogDebug, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func (l *Logger) Info(message string) {
	l.log(LogInfo, message)
}

// Infof logs a formatted info message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LogInfo, fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func (l *Logger) Warn(message string) {
	l.log(LogWarn, message)
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LogWarn, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func (l *Logger) Error(message string) {
	l.log(LogError, message)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LogError, fmt.Sprintf(format, args...))
}

// Entries returns the in-memory log entries.
func (l *Logger) Entries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Return a copy
	entries := make([]LogEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// EntriesSince returns entries after a given time.
func (l *Logger) EntriesSince(since time.Time) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, entry := range l.entries {
		if entry.Time.After(since) {
			result = append(result, entry)
		}
	}
	return result
}

// EntriesByLevel returns entries of a specific level.
func (l *Logger) EntriesByLevel(level LogLevel) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, entry := range l.entries {
		if entry.Level == level {
			result = append(result, entry)
		}
	}
	return result
}

// Clear clears the in-memory entries.
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]LogEntry, 0, l.maxSize)
}

// Close closes the logger.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Count returns the number of entries.
func (l *Logger) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}

// Last returns the last N entries.
func (l *Logger) Last(n int) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if n >= len(l.entries) {
		entries := make([]LogEntry, len(l.entries))
		copy(entries, l.entries)
		return entries
	}

	start := len(l.entries) - n
	entries := make([]LogEntry, n)
	copy(entries, l.entries[start:])
	return entries
}
