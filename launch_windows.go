//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func relaunchAsAdmin() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	argLine := composeWindowsCommandLine(os.Args[1:])

	launchLogger.Printf("shell execute runas file=%s args=%q cwd=%s", exePath, argLine, exeDir)

	verbPtr, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	filePtr, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return err
	}
	cwdPtr, err := windows.UTF16PtrFromString(exeDir)
	if err != nil {
		return err
	}

	var argsPtr *uint16
	if argLine != "" {
		argsPtr, err = windows.UTF16PtrFromString(argLine)
		if err != nil {
			return err
		}
	}

	return windows.ShellExecute(0, verbPtr, filePtr, argsPtr, cwdPtr, 1)
}

func openInEdge(url string) error {
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			launchLogger.Printf("opening edge via path=%s", candidate)
			return exec.Command(candidate, url).Start()
		}
	}

	edgeURI := "microsoft-edge:" + url
	launchLogger.Printf("opening edge via uri=%s", edgeURI)
	if err := exec.Command("cmd", "/c", "start", "", edgeURI).Start(); err == nil {
		return nil
	}

	launchLogger.Printf("edge uri fallback failed, using default browser")
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func composeWindowsCommandLine(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return windows.ComposeCommandLine(args)
}
