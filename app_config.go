package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AppConfig struct {
	Preset string `json:"preset"`
	UA     string `json:"ua"`
	TTL    uint8  `json:"ttl"`
	Log    string `json:"log"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Preset: "wechat",
		UA:     presets["wechat"],
		TTL:    64,
		Log:    "info",
	}
}

func (c AppConfig) Normalize() AppConfig {
	out := c
	if out.Preset == "" {
		out.Preset = "wechat"
	}
	out.Preset = strings.ToLower(strings.TrimSpace(out.Preset))
	if out.TTL == 0 {
		out.TTL = 64
	}
	out.Log = strings.ToLower(strings.TrimSpace(out.Log))
	if out.Log == "" {
		out.Log = "info"
	}
	switch out.Log {
	case "debug", "info", "warn":
	default:
		out.Log = "info"
	}
	if out.Preset != "pc" && out.Preset != "wechat" {
		out.Preset = "wechat"
	}
	out.UA = resolveUA(out.Preset)
	return out
}

func (c AppConfig) EffectiveUA() string {
	return c.Normalize().UA
}

func (c AppConfig) ToRuntimeConfig() *Config {
	norm := c.Normalize()
	return &Config{
		TTL: norm.TTL,
		UA:  norm.EffectiveUA(),
		Log: norm.Log,
	}
}

func LoadAppConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultAppConfig(), nil
		}
		return AppConfig{}, err
	}
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("decode app config: %w", err)
	}
	return cfg.Normalize(), nil
}

func SaveAppConfig(path string, cfg AppConfig) error {
	cfg = cfg.Normalize()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
