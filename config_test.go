package main

import (
	"flag"
	"testing"
)

func TestParseFlagsSupportsPresetUA(t *testing.T) {
	cfg, err := parseFlags([]string{"-ua", "pc", "-ports", "80,8080", "-ttl", "64"})
	if err != nil {
		t.Fatalf("parseFlags returned error: %v", err)
	}

	if cfg.UA != "pc" {
		t.Fatalf("expected UA flag value to stay as preset key before resolveUA, got %q", cfg.UA)
	}
	if cfg.TTL != 64 {
		t.Fatalf("expected TTL=64, got %d", cfg.TTL)
	}
	if len(cfg.UA_Ports) != 2 || cfg.UA_Ports[0] != 80 || cfg.UA_Ports[1] != 8080 {
		t.Fatalf("unexpected ports: %#v", cfg.UA_Ports)
	}
}

func TestParseFlagsRejectsTTLAbove255(t *testing.T) {
	_, err := parseFlags([]string{"-ttl", "300"})
	if err == nil {
		t.Fatal("expected TTL range error, got nil")
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
