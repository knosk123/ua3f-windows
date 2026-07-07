package main

import (
	"path/filepath"
)

type App struct {
	exeDir     string
	configPath string
	state      *AppState
	runner     *Runner
}

func NewApp(exeDir string) (*App, error) {
	configPath := filepath.Join(exeDir, "ua3f-config.json")
	cfg, err := LoadAppConfig(configPath)
	if err != nil {
		return nil, err
	}

	state := NewAppState(cfg)
	app := &App{
		exeDir:     exeDir,
		configPath: configPath,
		state:      state,
	}
	app.runner = NewRunner(exeDir, state)
	return app, nil
}

func (a *App) SaveConfig(cfg AppConfig) error {
	cfg = cfg.Normalize()
	if err := SaveAppConfig(a.configPath, cfg); err != nil {
		return err
	}
	a.state.SetConfig(cfg)
	return nil
}

func (a *App) LoadConfig() AppConfig {
	return a.state.Config()
}

func (a *App) Snapshot() AppSnapshot {
	return a.state.Snapshot()
}

func (a *App) Start() error {
	cfg := a.state.Config()
	return a.runner.Start(cfg)
}

func (a *App) Stop() error {
	return a.runner.Stop()
}
