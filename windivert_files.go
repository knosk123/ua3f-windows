package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func prepareWinDivertFiles(exeDir string) error {
	if err := writeEmbeddedIfMissing(exeDir, "windivert_amd64.dll"); err != nil {
		return err
	}
	if err := writeEmbeddedIfMissing(exeDir, "windivert_amd64.sys"); err != nil {
		return err
	}
	if err := copyIfMissing(
		filepath.Join(exeDir, "windivert_amd64.dll"),
		filepath.Join(exeDir, "WinDivert.dll"),
	); err != nil {
		return err
	}
	if err := copyIfMissing(
		filepath.Join(exeDir, "windivert_amd64.sys"),
		filepath.Join(exeDir, "WinDivert64.sys"),
	); err != nil {
		return err
	}
	return nil
}

func writeEmbeddedIfMissing(exeDir, name string) error {
	target := filepath.Join(exeDir, name)
	if info, err := os.Stat(target); err == nil && info.Size() > 0 {
		return nil
	}

	data, err := embeddedWinDivert.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read embedded resource %s: %w", name, err)
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}
	return nil
}

func copyIfMissing(src, dst string) error {
	if info, err := os.Stat(dst); err == nil && info.Size() > 0 {
		return nil
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
