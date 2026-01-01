//go:build windows

// Package dialog provides native Windows dialogs.
package dialog

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW = user32.NewProc("MessageBoxW")
)

const (
	MB_OK              = 0x00000000
	MB_OKCANCEL        = 0x00000001
	MB_YESNO           = 0x00000004
	MB_YESNOCANCEL     = 0x00000003
	MB_ICONWARNING     = 0x00000030
	MB_ICONQUESTION    = 0x00000020
	MB_ICONERROR       = 0x00000010
	MB_ICONINFORMATION = 0x00000040
	MB_DEFBUTTON2      = 0x00000100

	IDYES    = 6
	IDNO     = 7
	IDOK     = 1
	IDCANCEL = 2
)

// messageBox displays a Windows message box and returns the button clicked.
func messageBox(title, message string, flags uint32) int {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)

	return int(ret)
}

// Confirm shows a Yes/No confirmation dialog.
// Returns true if user clicks Yes.
func Confirm(title, message string) bool {
	result := messageBox(title, message, MB_YESNO|MB_ICONQUESTION|MB_DEFBUTTON2)
	return result == IDYES
}

// ConfirmWarning shows a Yes/No warning dialog.
// Returns true if user clicks Yes.
func ConfirmWarning(title, message string) bool {
	result := messageBox(title, message, MB_YESNO|MB_ICONWARNING|MB_DEFBUTTON2)
	return result == IDYES
}

// Alert shows an information dialog.
func Alert(title, message string) {
	messageBox(title, message, MB_OK|MB_ICONINFORMATION)
}

// Error shows an error dialog.
func Error(title, message string) {
	messageBox(title, message, MB_OK|MB_ICONERROR)
}

// Warning shows a warning dialog.
func Warning(title, message string) {
	messageBox(title, message, MB_OK|MB_ICONWARNING)
}

// ConfirmProtectedPush shows the protection confirmation dialog for pushes.
func ConfirmProtectedPush(repoName, branch, reason string, commitCount int) bool {
	title := "⚠️ Protected Repository"

	message := fmt.Sprintf(
		"Repository: %s\nBranch: %s",
		repoName, branch,
	)

	if reason != "" {
		message += fmt.Sprintf("\nReason: %s", reason)
	}

	message += fmt.Sprintf("\n\n%d commit(s) akan di-push.\n\nLanjutkan?", commitCount)

	return ConfirmWarning(title, message)
}
