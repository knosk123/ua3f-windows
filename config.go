package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	TTL uint8
	UA  string
	Log string
}

func parseFlags(args []string) (*Config, error) {
	for _, arg := range args {
		if arg == "-ports" || strings.HasPrefix(arg, "-ports=") || arg == "--ports" || strings.HasPrefix(arg, "--ports=") {
			return nil, fmt.Errorf("-ports is no longer supported; HTTP is detected on all TCP ports")
		}
	}

	fs := flag.NewFlagSet("ua3f-win", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	ttlValue := fs.Uint("ttl", 64, "Target IPv4 TTL (1-255)")
	uaValue := fs.String("ua", "wechat", "User-Agent preset: pc/wechat")
	logValue := fs.String("log", "info", "Log level: debug/info/warn")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "ua3f-win: rewrite TTL and HTTP User-Agent on Windows\n\n")
		fmt.Fprintf(fs.Output(), "Usage:\n  ua3f-win.exe [options]\n\n")
		fmt.Fprintf(fs.Output(), "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nExamples:\n")
		fmt.Fprintf(fs.Output(), "  ua3f-win.exe\n")
		fmt.Fprintf(fs.Output(), "  ua3f-win.exe -ua pc\n")
		fmt.Fprintf(fs.Output(), "  ua3f-win.exe -ua wechat -log debug\n")
		fmt.Fprintf(fs.Output(), "  ua3f-win.exe -ttl 65\n")
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if *ttlValue == 0 || *ttlValue > 255 {
		return nil, fmt.Errorf("-ttl must be between 1 and 255")
	}

	logLevel := strings.ToLower(strings.TrimSpace(*logValue))
	switch logLevel {
	case "debug", "info", "warn":
	default:
		return nil, fmt.Errorf("-log must be one of: debug, info, warn")
	}

	uaPreset := strings.ToLower(strings.TrimSpace(*uaValue))
	if uaPreset != "pc" && uaPreset != "wechat" {
		return nil, fmt.Errorf("-ua must be one of: pc, wechat")
	}

	return &Config{
		TTL: uint8(*ttlValue),
		UA:  uaPreset,
		Log: logLevel,
	}, nil
}
