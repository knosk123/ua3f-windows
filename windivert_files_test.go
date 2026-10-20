package main

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

func TestDriverFilesRepairCorruptionAndPreserveValidFiles(t *testing.T) {
	dir := t.TempDir()
	if err := prepareWinDivertFiles(dir); err != nil {
		t.Fatal(err)
	}
	names := map[string]string{"windivert_amd64.dll": "windivert_amd64.dll", "windivert_amd64.sys": "windivert_amd64.sys", "WinDivert.dll": "windivert_amd64.dll", "WinDivert64.sys": "windivert_amd64.sys"}
	for name, embedded := range names {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name)
			expected, _ := embeddedWinDivert.ReadFile(embedded)
			before, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := prepareWinDivertFiles(dir); err != nil {
				t.Fatal(err)
			}
			after, _ := os.Stat(path)
			if !before.ModTime().Equal(after.ModTime()) {
				t.Fatal("valid file was rewritten")
			}
			// Same size and nonempty: the previous existence-only check accepted this.
			corrupt := bytes.Repeat([]byte{0xaa}, len(expected))
			if err := os.WriteFile(path, corrupt, 0644); err != nil {
				t.Fatal(err)
			}
			if err := prepareWinDivertFiles(dir); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(path)
			if err != nil || sha256.Sum256(got) != sha256.Sum256(expected) {
				t.Fatalf("repair failed: %v", err)
			}
		})
	}
	entries, _ := filepath.Glob(filepath.Join(dir, ".ua3f-driver-*"))
	if len(entries) > 0 {
		t.Fatalf("temporary files leaked: %v", entries)
	}
}
func TestDriverRepairRefusesDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "WinDivert.dll"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := prepareWinDivertFiles(dir); err == nil {
		t.Fatal("expected non-regular target rejection")
	}
}
func TestDriverRepairFailureKeepsOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WinDivert.dll")
	if err := os.WriteFile(path, []byte("corrupt"), 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0644)
	if err := ensureVerifiedFile(path, []byte("correct")); err == nil {
		t.Fatal("expected read-only file replacement to fail")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "corrupt" {
		t.Fatalf("original lost: %q %v", data, err)
	}
}
