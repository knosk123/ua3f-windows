package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TTL      uint8
	UA       string
	UA_Ports []uint16
	Log      string
}

func parseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("ua3f-win", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	ttlValue := fs.Uint("ttl", 64, "Target IPv4 TTL (1-255)")
	uaValue := fs.String("ua", "wechat", "User-Agent preset: pc/wechat")
	logValue := fs.String("log", "info", "Log level: debug/info/warn")
	portsValue := fs.String("ports", "80", "Comma-separated destination ports for UA rewrite")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "ua3f-win: rewrite TTL and HTTP User-Agent on Windows\n\n")
		fmt.Fprintf(fs.Output(), "Usage:\n  ua3f-win.exe [options]\n\n")
		fmt.Fprintf(fs.Output(), "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nExamples:\n")
		fmt.Fprintf(fs.Output(), "  ua3f-win.exe\n")
		fmt.Fprintf(fs.Output(), "  ua3f-win.exe -ua pc\n")
		fmt.Fprintf(fs.Output(), "  ua3f-win.exe -ua wechat -log debug\n")
		fmt.Fprintf(fs.Output(), "  ua3f-win.exe -ttl 64 -ports \"80,8080\"\n")
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if *ttlValue == 0 || *ttlValue > 255 {
		return nil, fmt.Errorf("-ttl must be between 1 and 255")
	}

	ports, err := parsePorts(*portsValue)
	if err != nil {
		return nil, err
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
		TTL:      uint8(*ttlValue),
		UA:       uaPreset,
		UA_Ports: ports,
		Log:      logLevel,
	}, nil
}

func parsePorts(raw string) ([]uint16, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []uint16{80}, nil
	}

	parts := strings.Split(raw, ",")
	ports := make([]uint16, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		value, err := strconv.ParseUint(part, 10, 16)
		if err != nil || value == 0 {
			return nil, fmt.Errorf("invalid port value: %q", part)
		}
		ports = append(ports, uint16(value))
	}

	if len(ports) == 0 {
		return []uint16{80}, nil
	}
	return ports, nil
}
