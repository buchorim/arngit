// Package ui provides styled console output utilities.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Colors
var (
	green  = color.New(color.FgGreen, color.Bold)
	red    = color.New(color.FgRed, color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)
	blue   = color.New(color.FgCyan, color.Bold)
	white  = color.New(color.FgWhite, color.Bold)
	dim    = color.New(color.FgHiBlack)
)

// Symbols
const (
	SymbolSuccess = "✓"
	SymbolError   = "✗"
	SymbolWarning = "⚠"
	SymbolInfo    = "→"
	SymbolBullet  = "•"
	SymbolArrow   = "›"
)

// Global spinner
var currentSpinner *spinner.Spinner

// Success prints a success message.
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	green.Printf("  %s %s\n", SymbolSuccess, msg)
}

// Error prints an error message.
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	red.Printf("  %s %s\n", SymbolError, msg)
}

// Warning prints a warning message.
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	yellow.Printf("  %s %s\n", SymbolWarning, msg)
}

// Info prints an info message.
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	blue.Printf("  %s %s\n", SymbolInfo, msg)
}

// Print prints a regular message.
func Print(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s\n", msg)
}

// Dim prints dimmed text.
func Dim(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	dim.Printf("  %s\n", msg)
}

// Header prints a styled header.
func Header(title string) {
	fmt.Println()
	white.Printf("  ┌─ %s ", title)
	fmt.Println(strings.Repeat("─", max(0, 40-len(title))))
	fmt.Println()
}

// Divider prints a divider line.
func Divider() {
	dim.Println("  " + strings.Repeat("─", 45))
}

// NewLine prints an empty line.
func NewLine() {
	fmt.Println()
}

// StartSpinner starts a loading spinner.
func StartSpinner(msg string) {
	StopSpinner() // Stop any existing spinner

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "  "
	s.Suffix = " " + msg
	s.Color("cyan")
	s.Start()

	currentSpinner = s
}

// StopSpinner stops the current spinner.
func StopSpinner() {
	if currentSpinner != nil {
		currentSpinner.Stop()
		currentSpinner = nil
	}
}

// StopSpinnerSuccess stops spinner and shows success message.
func StopSpinnerSuccess(msg string) {
	StopSpinner()
	Success(msg)
}

// StopSpinnerError stops spinner and shows error message.
func StopSpinnerError(msg string) {
	StopSpinner()
	Error(msg)
}

// Table prints a formatted table.
func Table(headers []string, rows [][]string) {
	// Print header
	fmt.Println()
	dim.Print("  ")
	for _, h := range headers {
		fmt.Printf("%-16s", h)
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("  ")
		for i, cell := range row {
			if i == 0 && cell != "" {
				green.Printf("%-16s", cell)
			} else {
				fmt.Printf("%-16s", cell)
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

// KeyValue prints a key-value pair.
func KeyValue(key, value string) {
	dim.Printf("  %-12s", key+":")
	fmt.Printf(" %s\n", value)
}

// Bullet prints a bulleted item.
func Bullet(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", SymbolBullet, msg)
}

// SubItem prints a sub-item with indentation.
func SubItem(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	dim.Printf("    %s ", SymbolArrow)
	fmt.Println(msg)
}

// StatusLine prints a status with color based on success.
func StatusLine(label string, ok bool, details string) {
	if ok {
		green.Printf("  %s ", SymbolSuccess)
	} else {
		red.Printf("  %s ", SymbolError)
	}
	fmt.Printf("%-18s %s\n", label, details)
}

// ProgressDots prints a simple progress indicator.
func ProgressDots(count int) {
	for i := 0; i < count; i++ {
		fmt.Print(".")
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println()
}

// Banner prints the application banner.
func Banner() {
	blue.Println(`
    ╭─────────────────────────────────────────╮
    │                                         │
    │     ▄▀█ █▀█ █▄░█ █▀▀ █ ▀█▀              │
    │     █▀█ █▀▄ █░▀█ █▄█ █ ░█░              │
    │                                         │
    │     Arinara Git  ·  Modern Git Wrapper  │
    │                                         │
    ╰─────────────────────────────────────────╯`)
	fmt.Println()
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
