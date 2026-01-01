//go:build windows

// Package input provides secure input utilities.
package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

const (
	ENABLE_ECHO_INPUT = 0x0004
)

// ReadLine reads a line from stdin with prompt.
func ReadLine(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// ReadLineRequired reads a line and ensures it's not empty.
func ReadLineRequired(prompt string) (string, error) {
	for {
		value, err := ReadLine(prompt)
		if err != nil {
			return "", err
		}
		if value != "" {
			return value, nil
		}
		fmt.Println("  This field is required. Please try again.")
	}
}

// ReadPassword reads a password without echoing to console.
func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Get stdin handle
	handle := syscall.Handle(os.Stdin.Fd())

	// Get current console mode
	var mode uint32
	procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))

	// Disable echo
	newMode := mode &^ ENABLE_ECHO_INPUT
	procSetConsoleMode.Call(uintptr(handle), uintptr(newMode))

	// Read password
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')

	// Restore console mode
	procSetConsoleMode.Call(uintptr(handle), uintptr(mode))

	// Print newline (since echo was disabled)
	fmt.Println()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(password), nil
}

// ReadPasswordRequired reads a password and ensures it's not empty.
func ReadPasswordRequired(prompt string) (string, error) {
	for {
		value, err := ReadPassword(prompt)
		if err != nil {
			return "", err
		}
		if value != "" {
			return value, nil
		}
		fmt.Println("  This field is required. Please try again.")
	}
}

// Confirm asks for yes/no confirmation.
func Confirm(prompt string, defaultYes bool) bool {
	suffix := " [y/N]: "
	if defaultYes {
		suffix = " [Y/n]: "
	}

	fmt.Print(prompt + suffix)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}
