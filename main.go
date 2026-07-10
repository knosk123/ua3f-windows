package main

import (
	"embed"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

//go:embed windivert_amd64.dll
//go:embed windivert_amd64.sys
var embeddedWinDivert embed.FS

func main() {
	log.SetFlags(log.LstdFlags)
	if exeDir, err := executableDir(); err == nil {
		launchLogger = NewLaunchLogger(filepath.Join(exeDir, "ua3f-launch.log"))
	}
	launchLogger.Printf("process start pid=%d args=%q", os.Getpid(), os.Args)

	if shouldRunLegacyCLI(os.Args[1:]) {
		launchLogger.Printf("mode=legacy-cli")
		runLegacyCLI()
		return
	}

	admin := isAdmin() || isAdminFallback()
	launchLogger.Printf("ui mode detected admin=%t", admin)
	if !admin {
		launchLogger.Printf("requesting elevation")
		if err := relaunchAsAdmin(); err != nil {
			launchLogger.Printf("relaunch as admin failed: %v", err)
			log.Fatalf("[-] failed to relaunch as administrator: %v", err)
		}
		launchLogger.Printf("relaunch command started successfully, current process exiting")
		return
	}

	exeDir, err := executableDir()
	if err != nil {
		launchLogger.Printf("failed to resolve executable directory: %v", err)
		log.Fatalf("[-] failed to resolve executable directory: %v", err)
	}
	launchLogger.Printf("running elevated with exeDir=%s", exeDir)

	app, err := NewApp(exeDir)
	if err != nil {
		launchLogger.Printf("app init failed: %v", err)
		log.Fatalf("[-] failed to initialize app: %v", err)
	}
	app.state.AppendLog("[+] ui mode started")
	launchLogger.Printf("app initialized")

	server, url, quit, err := startWebServer(app)
	if err != nil {
		launchLogger.Printf("web server start failed: %v", err)
		log.Fatalf("[-] failed to start web server: %v", err)
	}
	log.Printf("[+] web ui: %s", url)
	app.state.AppendLog("[+] web ui: " + url)
	launchLogger.Printf("web server started url=%s", url)

	if err := openInEdge(url); err != nil {
		launchLogger.Printf("open edge failed: %v", err)
		log.Printf("[-] failed to open Edge automatically: %v", err)
		app.state.AppendLog("[-] failed to open Edge automatically: " + err.Error())
	} else {
		launchLogger.Printf("open edge command started successfully")
	}

	launchLogger.Printf("entering waitForShutdown")
	waitForShutdown(server, app, quit)
}

func shouldRunLegacyCLI(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-ua") ||
			strings.HasPrefix(arg, "-ttl") ||
			strings.HasPrefix(arg, "-ports") ||
			strings.HasPrefix(arg, "-log") ||
			arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func runLegacyCLI() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(2)
	}
	cfg.UA = resolveUA(cfg.UA)

	if !isAdmin() && !isAdminFallback() {
		log.Printf("[-] administrator privileges are required")
		log.Printf("    run ua3f-win.exe from an elevated cmd or PowerShell window")
		os.Exit(1)
	}

	exeDir, err := executableDir()
	if err != nil {
		log.Fatalf("[-] failed to resolve executable directory: %v", err)
	}

	state := NewAppState(DefaultAppConfig())
	runner := NewRunner(exeDir, state)
	appCfg := AppConfig{
		Preset: detectPresetKey(cfg.UA),
		UA:     cfg.UA,
		TTL:    cfg.TTL,
		Log:    cfg.Log,
	}.Normalize()
	state.SetConfig(appCfg)

	if err := runner.Start(appCfg); err != nil {
		log.Fatalf("[-] failed to start WinDivert: %v", err)
	}

	log.Printf("[*] press Ctrl+C to stop")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	_ = runner.Stop()
}

func waitForShutdown(server *http.Server, app *App, quit <-chan struct{}) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case <-sigCh:
		launchLogger.Printf("shutdown signal received")
	case <-quit:
		launchLogger.Printf("shutdown requested from web ui")
	}

	app.state.AppendLog("[*] shutdown requested")
	_ = app.Stop()
	_ = server.Close()
}

func detectPresetKey(ua string) string {
	for key, value := range presets {
		if ua == value {
			return key
		}
	}
	return "wechat"
}
