package main

import (
	"flag"
	"strings"
	"testing"
)

func TestParseFlagsSupportsPresetUA(t *testing.T) {
	cfg, err := parseFlags([]string{"-ua", "pc", "-ttl", "64"})
	if err != nil {
		t.Fatalf("parseFlags returned error: %v", err)
	}

	if cfg.UA != "pc" {
		t.Fatalf("expected UA flag value to stay as preset key before resolveUA, got %q", cfg.UA)
	}
	if cfg.TTL != 64 {
		t.Fatalf("expected TTL=64, got %d", cfg.TTL)
	}
}

func TestParseFlagsRejectsTTLAbove255(t *testing.T) {
	_, err := parseFlags([]string{"-ttl", "300"})
	if err == nil {
		t.Fatal("expected TTL range error, got nil")
	}
}

func TestParseFlagsRejectsRemovedPortsOption(t *testing.T) {
	_, err := parseFlags([]string{"-ports", "80"})
	if err == nil {
		t.Fatal("expected removed -ports option to be rejected")
	}
	if !strings.Contains(err.Error(), "ports") {
		t.Fatalf("expected error to identify removed ports option, got %v", err)
	}
	if !strings.Contains(err.Error(), "no longer supported") {
		t.Fatalf("expected error to explain that ports are detected automatically, got %v", err)
	}
}

func TestResolveUAPresetAndLiteral(t *testing.T) {
	if got := resolveUA("wechat"); got != presets["wechat"] {
		t.Fatalf("expected wechat preset, got %q", got)
	}

	literal := "Mozilla/5.0 custom"
	if got := resolveUA(literal); got != presets["wechat"] {
		t.Fatalf("expected literal UA to fall back to wechat, got %q", got)
	}
}

func TestParseFlagsRejectsCustomUA(t *testing.T) {
	_, err := parseFlags([]string{"-ua", "Mozilla/5.0 custom"})
	if err == nil {
		t.Fatal("expected custom UA to be rejected")
	}
}

func TestParseFlagsHelpReturnsFlagErrHelp(t *testing.T) {
	_, err := parseFlags([]string{"-h"})
	if err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp, got %#v", err)
	}
}
