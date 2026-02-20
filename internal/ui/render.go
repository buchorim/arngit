// Package ui provides terminal UI components for ArnGit.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Colors for terminal output
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"

	// Foreground colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
)

// Symbols for UI elements
const (
	SymbolCheck   = "✓"
	SymbolCross   = "✗"
	SymbolArrow   = "→"
	SymbolDot     = "●"
	SymbolCircle  = "○"
	SymbolStar    = "★"
	SymbolInfo    = "ℹ"
	SymbolWarning = "⚠"
	SymbolError   = "✖"
)

// Renderer handles terminal output with styling.
type Renderer struct {
	colorEnabled bool
	reader       *bufio.Reader
}

// NewRenderer creates a new UI renderer.
func NewRenderer(colorEnabled bool) *Renderer {
	return &Renderer{
		colorEnabled: colorEnabled,
		reader:       bufio.NewReader(os.Stdin),
	}
}

// Color applies color if enabled.
func (r *Renderer) Color(c, text string) string {
	if !r.colorEnabled {
		return text
	}
	return c + text + Reset
}

// color is an alias for Color (for internal use).
func (r *Renderer) color(c, text string) string {
	return r.Color(c, text)
}

// Logo prints the ArnGit logo.
func (r *Renderer) Logo() {
	logo := `
   _____                _______ __  
  /  _  \_______  ____ /  _____//__|_/  |_ 
 /  /_\  \_  __ \/    \/   \  __\|  \   __\
/    |    \  | \/   |  \    \_\  \  ||  |  
\____|__  /__|  |___|  /\______  /__||__|  
        \/           \/        \/          
`
	fmt.Println(r.color(BrightCyan, logo))
}

// Title prints a section title.
func (r *Renderer) Title(text string) {
	fmt.Println()
	fmt.Println(r.color(Bold+BrightCyan, text))
	fmt.Println(r.color(Dim, strings.Repeat("─", len(text)+2)))
}

// Info prints an info message.
func (r *Renderer) Info(text string) {
	fmt.Println(text)
}

// Success prints a success message.
func (r *Renderer) Success(text string) {
	symbol := r.color(BrightGreen, SymbolCheck)
	fmt.Printf("%s %s\n", symbol, r.color(Green, text))
}

// Warning prints a warning message.
func (r *Renderer) Warning(text string) {
	symbol := r.color(BrightYellow, SymbolWarning)
	fmt.Printf("%s %s\n", symbol, r.color(Yellow, text))
}

// Error prints an error message.
func (r *Renderer) Error(text string) {
	symbol := r.color(BrightRed, SymbolError)
	fmt.Printf("%s %s\n", symbol, r.color(Red, text))
}

// Hint prints a hint message.
func (r *Renderer) Hint(text string) {
	fmt.Printf("%s %s\n", r.color(Dim, "→"), r.color(Dim, text))
}

// Prompt prints a prompt and reads user input.
func (r *Renderer) Prompt(label string) string {
	fmt.Printf("%s: ", r.color(BrightCyan, label))
	input, _ := r.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// PromptSecret prints a prompt and reads password input.
func (r *Renderer) PromptSecret(label string) string {
	fmt.Printf("%s: ", r.color(BrightCyan, label))

	// Try to read password without echo
	if term.IsTerminal(int(os.Stdin.Fd())) {
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // New line after password
		if err == nil {
			return string(password)
		}
	}

	// Fallback to normal input
	input, _ := r.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// Confirm asks for a yes/no confirmation.
func (r *Renderer) Confirm(question string) bool {
	fmt.Printf("%s [y/N]: ", r.color(BrightYellow, question))
	input, _ := r.reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

// Box prints text in a styled box.
func (r *Renderer) Box(title, content string, color string) {
	if color == "" {
		color = Cyan
	}

	lines := strings.Split(content, "\n")
	maxLen := len(title)
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	width := maxLen + 4

	// Top border
	fmt.Println(r.color(color, "╭"+strings.Repeat("─", width)+"╮"))

	// Title
	if title != "" {
		padding := width - len(title) - 2
		fmt.Printf("%s %s%s %s\n",
			r.color(color, "│"),
			r.color(Bold+color, title),
			strings.Repeat(" ", padding),
			r.color(color, "│"))
		fmt.Println(r.color(color, "├"+strings.Repeat("─", width)+"┤"))
	}

	// Content
	for _, line := range lines {
		padding := width - len(line) - 2
		fmt.Printf("%s %s%s %s\n",
			r.color(color, "│"),
			line,
			strings.Repeat(" ", padding),
			r.color(color, "│"))
	}

	// Bottom border
	fmt.Println(r.color(color, "╰"+strings.Repeat("─", width)+"╯"))
}

// Table prints a simple table.
func (r *Renderer) Table(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	fmt.Print(r.color(Bold, "  "))
	for i, h := range headers {
		fmt.Printf(r.color(Bold, "%-*s  "), widths[i], h)
	}
	fmt.Println()

	// Print separator
	fmt.Print("  ")
	for i := range headers {
		fmt.Print(r.color(Dim, strings.Repeat("─", widths[i])) + "  ")
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("  ")
		for i, cell := range row {
			if i < len(widths) {
				fmt.Printf("%-*s  ", widths[i], cell)
			}
		}
		fmt.Println()
	}
}

// ProgressBar prints a progress bar.
func (r *Renderer) ProgressBar(current, total int, label string) {
	width := 30
	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Printf("\r%s %s %s %3.0f%%",
		r.color(Cyan, label),
		r.color(BrightCyan, bar),
		r.color(Dim, fmt.Sprintf("(%d/%d)", current, total)),
		percent*100)

	if current >= total {
		fmt.Println()
	}
}

// Spinner symbols for animation
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner prints a spinner frame.
func (r *Renderer) Spinner(frame int, text string) {
	symbol := spinnerFrames[frame%len(spinnerFrames)]
	fmt.Printf("\r%s %s", r.color(BrightCyan, symbol), text)
}

// ClearLine clears the current line.
func (r *Renderer) ClearLine() {
	fmt.Print("\r\033[K")
}

// Badge prints a colored badge.
func (r *Renderer) Badge(text, bgColor string) {
	fmt.Printf(" %s ", r.color(bgColor+White+Bold, " "+text+" "))
}
