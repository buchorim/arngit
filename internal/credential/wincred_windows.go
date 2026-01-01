//go:build windows

// Package credential provides Windows Credential Manager integration.
package credential

import (
	"syscall"
	"unsafe"
)

var (
	advapi32           = syscall.NewLazyDLL("advapi32.dll")
	procCredWriteW     = advapi32.NewProc("CredWriteW")
	procCredReadW      = advapi32.NewProc("CredReadW")
	procCredDeleteW    = advapi32.NewProc("CredDeleteW")
	procCredFree       = advapi32.NewProc("CredFree")
	procCredEnumerateW = advapi32.NewProc("CredEnumerateW")
)

const (
	CRED_TYPE_GENERIC          = 1
	CRED_PERSIST_LOCAL_MACHINE = 2
)

type credential struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        uint64
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

// Store saves a credential to Windows Credential Manager.
func Store(target, username, password string) error {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}

	usernamePtr, err := syscall.UTF16PtrFromString(username)
	if err != nil {
		return err
	}

	passwordBytes := []byte(password)

	cred := credential{
		Type:               CRED_TYPE_GENERIC,
		TargetName:         targetPtr,
		CredentialBlobSize: uint32(len(passwordBytes)),
		CredentialBlob:     &passwordBytes[0],
		Persist:            CRED_PERSIST_LOCAL_MACHINE,
		UserName:           usernamePtr,
	}

	ret, _, err := procCredWriteW.Call(
		uintptr(unsafe.Pointer(&cred)),
		0,
	)

	if ret == 0 {
		return err
	}

	return nil
}

// Retrieve gets a credential from Windows Credential Manager.
func Retrieve(target string) (username, password string, err error) {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return "", "", err
	}

	var credPtr *credential

	ret, _, err := procCredReadW.Call(
		uintptr(unsafe.Pointer(targetPtr)),
		CRED_TYPE_GENERIC,
		0,
		uintptr(unsafe.Pointer(&credPtr)),
	)

	if ret == 0 {
		return "", "", err
	}

	defer procCredFree.Call(uintptr(unsafe.Pointer(credPtr)))

	// Extract username
	if credPtr.UserName != nil {
		username = syscall.UTF16ToString((*[256]uint16)(unsafe.Pointer(credPtr.UserName))[:])
	}

	// Extract password
	if credPtr.CredentialBlobSize > 0 {
		passwordBytes := make([]byte, credPtr.CredentialBlobSize)
		for i := uint32(0); i < credPtr.CredentialBlobSize; i++ {
			passwordBytes[i] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(credPtr.CredentialBlob)) + uintptr(i)))
		}
		password = string(passwordBytes)
	}

	return username, password, nil
}

// Delete removes a credential from Windows Credential Manager.
func Delete(target string) error {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}

	ret, _, err := procCredDeleteW.Call(
		uintptr(unsafe.Pointer(targetPtr)),
		CRED_TYPE_GENERIC,
		0,
	)

	if ret == 0 {
		return err
	}

	return nil
}

// Exists checks if a credential exists.
func Exists(target string) bool {
	_, _, err := Retrieve(target)
	return err == nil
}
