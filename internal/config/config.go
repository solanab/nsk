// Package config loads XDG TOML, NSK_* overlays, and Server bind rules.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var (
	errConfigMissing = errors.New("找不到配置文件")
	errServerClient  = errors.New("[server] 与 [client] 不能同时配置")
	readFile         = os.ReadFile  //nolint:gochecknoglobals // ReadFile test seam
	absPath          = filepath.Abs //nolint:gochecknoglobals // filepath.Abs test seam
)

// Config comes from $XDG_CONFIG_HOME/nsk/config.toml. NSK_* overlays same-named fields.
type Config struct {
	Dir      string  `toml:"-"`
	StateDir string  `toml:"-"`
	Path     string  `toml:"-"`
	Account  Account `toml:"account"`
	Server   Server  `toml:"server"`
	Client   Client  `toml:"client"`
}

// Account is the NodeSeek forum identity on this machine (pjwt cookie).
type Account struct {
	Username string `toml:"username"`
	Cookie   string `toml:"cookie"`
}

// Server is this machine holding Account for other nsk processes.
type Server struct {
	Addr  string `toml:"addr"`
	Token string `toml:"token"`
}

// Client is this machine connecting to another Server. No cookie.
type Client struct {
	URL   string `toml:"url"`
	Token string `toml:"token"`
}

// Load reads XDG config.toml; a missing file yields defaults.
func Load() (*Config, error) {
	dir, err := ResolveDir()
	if err != nil {
		return nil, err
	}

	return loadFrom(dir, filepath.Join(dir, FileName), false)
}

// LoadPath reads an explicit path. Empty path equals Load; a set path must exist.
func LoadPath(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Load()
	}

	abs, err := absPath(path)
	if err != nil {
		return nil, fmt.Errorf("无效 --config: %w", err)
	}

	return loadFrom(filepath.Dir(abs), abs, true)
}

func loadDir(dir string) (*Config, error) {
	return loadFrom(dir, filepath.Join(dir, FileName), false)
}

func loadFrom(dir, path string, requireFile bool) (*Config, error) {
	stateDir, err := ResolveStateDir()
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	cfg.Dir = dir
	cfg.Path = path
	cfg.StateDir = stateDir

	if fileExists(path) {
		data, err := readFile(path)
		if err != nil {
			return nil, err
		}

		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析 %s: %w", path, err)
		}

		cfg.Dir = dir
		cfg.Path = path
		cfg.StateDir = stateDir
	} else if requireFile {
		return nil, fmt.Errorf("%w: %s", errConfigMissing, path)
	}

	cfg.overlayEnv()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func envOr(current, key string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return strings.TrimSpace(current)
}

// CookieFile is the jar path: [account].cookie relative to the state directory, or absolute.
func (cfg *Config) CookieFile() string {
	return resolvePath(cfg.StateDir, cfg.Account.Cookie)
}

// ListenAddr is the Server bind address, defaulting to DefaultAddr.
func (cfg *Config) ListenAddr() string {
	if strings.TrimSpace(cfg.Server.Addr) == "" {
		return DefaultAddr
	}

	return strings.TrimSpace(cfg.Server.Addr)
}

func (cfg *Config) overlayEnv() {
	cfg.Account.Username = envOr(cfg.Account.Username, "NSK_USERNAME")
	cfg.Client.URL = envOr(cfg.Client.URL, "NSK_CLIENT_URL")
	cfg.Client.Token = envOr(cfg.Client.Token, "NSK_CLIENT_TOKEN")
	cfg.Server.Addr = envOr(cfg.Server.Addr, "NSK_SERVER_ADDR")
	cfg.Server.Token = envOr(cfg.Server.Token, "NSK_SERVER_TOKEN")
}

func (cfg *Config) hasServer() bool {
	return strings.TrimSpace(cfg.Server.Addr) != "" || strings.TrimSpace(cfg.Server.Token) != ""
}

func (cfg *Config) hasClient() bool {
	return strings.TrimSpace(cfg.Client.URL) != "" || strings.TrimSpace(cfg.Client.Token) != ""
}

func (cfg *Config) validate() error {
	if cfg.hasServer() && cfg.hasClient() {
		return fmt.Errorf("%w: %s", errServerClient, cfg.Path)
	}

	if !cfg.hasServer() {
		return nil
	}

	host, port, err := parseAddr(cfg.Server.Addr)
	if err != nil {
		return err
	}

	if strings.TrimSpace(cfg.Server.Addr) == "" {
		cfg.Server.Addr = DefaultAddr
	}

	return validateListen(host, port, cfg.Server.Token)
}
