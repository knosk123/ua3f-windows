package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultAppConfigUsesWechatPreset(t *testing.T) {
	cfg := DefaultAppConfig()

	if cfg.Preset != "wechat" {
		t.Fatalf("expected default preset to be wechat, got %q", cfg.Preset)
	}
	if cfg.UA != presets["wechat"] {
		t.Fatalf("expected default UA to use wechat preset, got %q", cfg.UA)
	}
	if cfg.TTL != 64 {
		t.Fatalf("expected default TTL to be 64, got %d", cfg.TTL)
	}
	if cfg.Ports != "80" {
		t.Fatalf("expected default ports to be 80, got %q", cfg.Ports)
	}
	if cfg.Log != "info" {
		t.Fatalf("expected default log level to be info, got %q", cfg.Log)
	}
}

func TestLoadAppConfigReturnsDefaultsWhenFileMissing(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadAppConfig(filepath.Join(dir, "ua3f-config.json"))
	if err != nil {
		t.Fatalf("LoadAppConfig returned error: %v", err)
	}

	if cfg.Preset != "wechat" || cfg.UA != presets["wechat"] {
		t.Fatalf("expected default config on missing file, got %#v", cfg)
	}
}

func TestSaveAndLoadAppConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ua3f-config.json")

	want := AppConfig{
		Preset: "pc",
		UA:     presets["pc"],
		TTL:    65,
		Ports:  "80,8080",
		Log:    "debug",
	}
	if err := SaveAppConfig(path, want); err != nil {
		t.Fatalf("SaveAppConfig returned error: %v", err)
	}

	got, err := LoadAppConfig(path)
	if err != nil {
		t.Fatalf("LoadAppConfig returned error: %v", err)
	}

	if got != want {
		t.Fatalf("round trip mismatch:\nwant=%#v\ngot=%#v", want, got)
	}
}

func TestAppConfigEffectiveUAUsesPresetOnly(t *testing.T) {
	wechat := AppConfig{Preset: "wechat", UA: "ignored", TTL: 64, Ports: "80", Log: "info"}
	if got := wechat.EffectiveUA(); got != presets["wechat"] {
		t.Fatalf("expected preset UA, got %q", got)
	}

	custom := AppConfig{Preset: "custom", UA: "Mozilla/5.0 custom", TTL: 64, Ports: "80", Log: "info"}
	if got := custom.EffectiveUA(); got != presets["wechat"] {
		t.Fatalf("expected custom UA to fall back to wechat preset, got %q", got)
	}
}

func TestWechatPresetUsesShortUA(t *testing.T) {
	want := "Mozilla/5.0 (Linux; Android 15; RMX6688 Build/AP3A.240617.008; wv) AppleWebKit/537.36"

	if got := presets["wechat"]; got != want {
		t.Fatalf("wechat preset mismatch:\nwant=%q\ngot=%q", want, got)
	}
}

func TestAppStateSnapshotAndLogs(t *testing.T) {
	state := NewAppState(DefaultAppConfig())
	state.SetStatus(StatusStarting, "booting")
	state.AppendLog("[+] started")

	snap := state.Snapshot()
	if snap.Status != StatusStarting {
		t.Fatalf("expected status %q, got %q", StatusStarting, snap.Status)
	}
	if snap.LastError != "booting" {
		t.Fatalf("expected last error/message to be recorded, got %q", snap.LastError)
	}
	if len(snap.Logs) != 1 || snap.Logs[0] != "[+] started" {
		t.Fatalf("unexpected logs: %#v", snap.Logs)
	}
}

func TestStartWebServerServesConfigAndPage(t *testing.T) {
	exeDir := t.TempDir()
	app, err := NewApp(exeDir)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	server, url, err := startWebServer(app)
	if err != nil {
		t.Fatalf("startWebServer returned error: %v", err)
	}
	defer server.Close()

	resp, err := http.Get(url + "/api/config")
	if err != nil {
		t.Fatalf("GET /api/config failed: %v", err)
	}
	defer resp.Body.Close()

	var cfg AppConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		t.Fatalf("decode config failed: %v", err)
	}
	if cfg.Preset != "wechat" {
		t.Fatalf("expected wechat preset, got %q", cfg.Preset)
	}

	pageResp, err := http.Get(url + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer pageResp.Body.Close()
	if pageResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from root page, got %d", pageResp.StatusCode)
	}
}

func TestLaunchLoggerWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ua3f-launch.log")
	logger := NewLaunchLogger(path)

	logger.Printf("hello %s", "world")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read launch log: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Fatalf("expected launch log to contain message, got %q", string(data))
	}
}

func TestComposeWindowsCommandLinePreservesQuotedArgs(t *testing.T) {
	got := composeWindowsCommandLine([]string{"-ua", "wechat", "-ports", "80,8080"})

	if !strings.Contains(got, "-ua wechat") {
		t.Fatalf("expected UA preset in command line, got %q", got)
	}
	if !strings.Contains(got, "-ports 80,8080") {
		t.Fatalf("expected ports in command line, got %q", got)
	}
}
