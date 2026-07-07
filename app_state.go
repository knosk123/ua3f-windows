package main

import (
	"sync"
	"time"
)

const (
	StatusIdle     = "idle"
	StatusStarting = "starting"
	StatusRunning  = "running"
	StatusStopping = "stopping"
	StatusError    = "error"
)

type AppSnapshot struct {
	Status      string    `json:"status"`
	Config      AppConfig `json:"config"`
	LastError   string    `json:"lastError"`
	StartedAt   time.Time `json:"startedAt"`
	PacketCount uint64    `json:"packetCount"`
	TTLCount    uint64    `json:"ttlCount"`
	UACount     uint64    `json:"uaCount"`
	Logs        []string  `json:"logs"`
}

type AppState struct {
	mu          sync.RWMutex
	status      string
	config      AppConfig
	lastError   string
	startedAt   time.Time
	packetCount uint64
	ttlCount    uint64
	uaCount     uint64
	logs        []string
}

func NewAppState(cfg AppConfig) *AppState {
	return &AppState{
		status: StatusIdle,
		config: cfg.Normalize(),
		logs:   make([]string, 0, 128),
	}
}

func (s *AppState) SetConfig(cfg AppConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg.Normalize()
}

func (s *AppState) Config() AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *AppState) SetStatus(status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
	s.lastError = message
	if status == StatusRunning {
		s.startedAt = time.Now()
	}
	if status == StatusIdle {
		s.startedAt = time.Time{}
	}
}

func (s *AppState) SetCounters(packetCount, ttlCount, uaCount uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.packetCount = packetCount
	s.ttlCount = ttlCount
	s.uaCount = uaCount
}

func (s *AppState) AppendLog(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, line)
	if len(s.logs) > 500 {
		s.logs = append([]string(nil), s.logs[len(s.logs)-500:]...)
	}
}

func (s *AppState) Snapshot() AppSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := append([]string(nil), s.logs...)
	return AppSnapshot{
		Status:      s.status,
		Config:      s.config,
		LastError:   s.lastError,
		StartedAt:   s.startedAt,
		PacketCount: s.packetCount,
		TTLCount:    s.ttlCount,
		UACount:     s.uaCount,
		Logs:        logs,
	}
}
