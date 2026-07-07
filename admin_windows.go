//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func isAdmin() bool {
	adminSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return false
	}

	member, err := windows.GetCurrentProcessToken().IsMember(adminSID)
	if err != nil {
		return false
	}
	return member
}

func isAdminFallback() bool {
	f, err := os.Open(`\\.\PHYSICALDRIVE0`)
	if err == nil {
		f.Close()
		return true
	}
	return false
}
