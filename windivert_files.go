package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var driverFilesMu sync.Mutex

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func prepareWinDivertFiles(exeDir string) error {
	driverFilesMu.Lock()
	defer driverFilesMu.Unlock()
	for _, file := range [][2]string{
		{"windivert_amd64.dll", "windivert_amd64.dll"},
		{"windivert_amd64.sys", "windivert_amd64.sys"},
		{"windivert_amd64.dll", "WinDivert.dll"},
		{"windivert_amd64.sys", "WinDivert64.sys"},
	} {
		data, err := embeddedWinDivert.ReadFile(file[0])
		if err != nil {
			return fmt.Errorf("read embedded resource %s: %w", file[0], err)
		}
		if err := ensureVerifiedFile(filepath.Join(exeDir, file[1]), data); err != nil {
			return err
		}
	}
	return nil
}

// The executable's embedded bytes are the source of truth for both file names.
// Never truncate an existing DLL/SYS: Windows may have it loaded in another process.
func ensureVerifiedFile(target string, data []byte) error {
	info, err := os.Lstat(target)
	if err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing to replace non-regular driver file %s", target)
		}
		existing, err := os.ReadFile(target)
		if err != nil {
			return fmt.Errorf("verify %s: %w", target, err)
		}
		if sha256.Sum256(existing) == sha256.Sum256(data) {
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect %s: %w", target, err)
	}
	f, err := os.CreateTemp(filepath.Dir(target), ".ua3f-driver-*")
	if err != nil {
		return fmt.Errorf("prepare %s: %w", target, err)
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	_, writeErr := f.Write(data)
	if writeErr == nil {
		writeErr = f.Sync()
	}
	closeErr := f.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("write temporary driver for %s: %w", target, err)
	}
	check, err := os.ReadFile(tmp)
	if err != nil || sha256.Sum256(check) != sha256.Sum256(data) {
		return fmt.Errorf("temporary driver verification failed for %s: %v", target, err)
	}
	if err := os.Rename(tmp, target); err != nil {
		return fmt.Errorf("repair %s failed (file may be in use; close other WinDivert tools and retry): %w", target, err)
	}
	check, err = os.ReadFile(target)
	if err != nil || sha256.Sum256(check) != sha256.Sum256(data) {
		return fmt.Errorf("driver verification failed after replacement of %s: %v", target, err)
	}
	return nil
}
