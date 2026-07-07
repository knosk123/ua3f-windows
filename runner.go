package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/williamfhe/godivert"
)

type Runner struct {
	mu       sync.Mutex
	exeDir   string
	state    *AppState
	handle   *godivert.WinDivertHandle
	running  bool
	stopping bool
}

func NewRunner(exeDir string, state *AppState) *Runner {
	return &Runner{
		exeDir: exeDir,
		state:  state,
	}
}

func (r *Runner) Start(cfg AppConfig) error {
	runtimeCfg, err := cfg.ToRuntimeConfig()
	if err != nil {
		r.state.SetStatus(StatusError, err.Error())
		r.logf("[-] invalid config: %v", err)
		return err
	}

	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		r.logf("[*] runner already active")
		return nil
	}
	r.running = true
	r.stopping = false
	r.mu.Unlock()

	r.state.SetConfig(cfg.Normalize())
	r.state.SetCounters(0, 0, 0)
	r.state.SetStatus(StatusStarting, "")
	r.logf("[+] starting WinDivert")

	if err := prepareWinDivertFiles(r.exeDir); err != nil {
		r.finishWithError(err)
		return err
	}

	dllPath := filepath.Join(r.exeDir, "WinDivert.dll")
	godivert.LoadDLL(dllPath, dllPath)

	handle, err := godivert.NewWinDivertHandle("outbound && ip")
	if err != nil {
		r.finishWithError(err)
		return err
	}

	r.mu.Lock()
	r.handle = handle
	r.mu.Unlock()

	r.state.SetStatus(StatusRunning, "")
	r.logf("[+] WinDivert running")
	go r.loop(handle, runtimeCfg)
	return nil
}

func (r *Runner) loop(handle *godivert.WinDivertHandle, cfg *Config) {
	var pktCount, ttlCount, uaCount uint64
	var loopErr error

	for {
		pkt, err := handle.Recv()
		if err != nil {
			loopErr = err
			break
		}
		pktCount++

		ttlChanged, uaChanged := rewrite(pkt, cfg)
		if ttlChanged {
			ttlCount++
		}
		if uaChanged {
			uaCount++
		}
		if ttlChanged || uaChanged {
			handle.HelperCalcChecksum(pkt)
		}
		if _, err := handle.Send(pkt); err != nil {
			r.logf("[-] send failed: %v", err)
		}
		r.state.SetCounters(pktCount, ttlCount, uaCount)

		if cfg.Log == "debug" && pktCount%5000 == 0 {
			r.logf("[debug] packets=%d ttl_rewrites=%d ua_rewrites=%d", pktCount, ttlCount, uaCount)
		}
	}

	_ = handle.Close()

	r.mu.Lock()
	wasStopping := r.stopping
	r.running = false
	r.stopping = false
	r.handle = nil
	r.mu.Unlock()

	if wasStopping {
		r.state.SetStatus(StatusIdle, "")
		r.logf("[*] WinDivert stopped")
		return
	}

	if loopErr != nil {
		r.state.SetStatus(StatusError, loopErr.Error())
		r.logf("[-] WinDivert loop ended: %v", loopErr)
		return
	}

	r.state.SetStatus(StatusIdle, "")
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		r.logf("[*] runner already idle")
		return nil
	}
	r.stopping = true
	handle := r.handle
	r.mu.Unlock()

	r.state.SetStatus(StatusStopping, "")
	r.logf("[*] stopping WinDivert")
	if handle != nil {
		return handle.Close()
	}
	return nil
}

func (r *Runner) finishWithError(err error) {
	r.mu.Lock()
	r.running = false
	r.stopping = false
	r.handle = nil
	r.mu.Unlock()

	r.state.SetStatus(StatusError, err.Error())
	r.logf("[-] failed to start WinDivert: %v", err)
}

func (r *Runner) logf(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	log.Print(line)
	r.state.AppendLog(line)
}
