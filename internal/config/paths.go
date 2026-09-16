package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// XDG names for nsk config and cookie files.
const (
	AppName  = "nsk"
	FileName = "config.toml"
	Cookie   = "cookie.json"
)

// userHomeDir is the HOME lookup seam for missing-HOME tests.
var userHomeDir = os.UserHomeDir //nolint:gochecknoglobals

func fileExists(path string) bool {
	fileInfo, err := os.Stat(path)

	return err == nil && !fileInfo.IsDir()
}

// ResolveDir returns $XDG_CONFIG_HOME/nsk, else $HOME/.config/nsk.
func ResolveDir() (string, error) {
	if value := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); value != "" {
		return filepath.Join(value, AppName), nil
	}

	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法解析 XDG 配置目录: %w", err)
	}

	return filepath.Join(home, ".config", AppName), nil
}

// ResolveStateDir returns $XDG_STATE_HOME/nsk, else $HOME/.local/state/nsk.
func ResolveStateDir() (string, error) {
	if value := strings.TrimSpace(os.Getenv("XDG_STATE_HOME")); value != "" {
		return filepath.Join(value, AppName), nil
	}

	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法解析 XDG 状态目录: %w", err)
	}

	return filepath.Join(home, ".local", "state", AppName), nil
}

func resolvePath(dir, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "." {
		return filepath.Join(dir, Cookie)
	}

	if filepath.IsAbs(raw) {
		return raw
	}

	return filepath.Join(dir, raw)
}
