package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type LaunchLogger struct {
	path string
	mu   sync.Mutex
}

func NewLaunchLogger(path string) *LaunchLogger {
	return &LaunchLogger{path: path}
}

func (l *LaunchLogger) Printf(format string, args ...any) {
	if l == nil || l.path == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	line := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(f, "%s %s\n", time.Now().Format("2006-01-02 15:04:05.000"), line)
}

var launchLogger = NewLaunchLogger("")
