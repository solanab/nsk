package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/config"
)

func writeConfig(t *testing.T, dir, body string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func isolateXDG(t *testing.T) (string, string) {
	t.Helper()

	configHome := t.TempDir()
	stateHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_STATE_HOME", stateHome)
	t.Setenv("HOME", t.TempDir())

	for _, key := range []string{
		"NSK_USERNAME",
		"NSK_SERVER_ADDR", "NSK_SERVER_TOKEN",
		"NSK_CLIENT_URL", "NSK_CLIENT_TOKEN",
	} {
		t.Setenv(key, "")
	}

	return configHome, stateHome
}

func TestLoadUsesXDGConfigHome(t *testing.T) {
	xdg, state := isolateXDG(t)
	dir := filepath.Join(xdg, config.AppName)
	writeConfig(t, dir, "\n[account]\nusername = \"alice\"\n")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Account.Username != "alice" || cfg.Dir != dir {
		t.Fatalf("%#v", cfg)
	}

	if cfg.CookieFile() != filepath.Join(state, config.AppName, config.Cookie) {
		t.Fatalf("cookie %s", cfg.CookieFile())
	}
}

func TestLoadMissingFileDefaults(t *testing.T) {
	_, state := isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)

	cfg, err := config.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Server.Addr != "" || cfg.Client.URL != "" {
		t.Fatalf("%#v %#v", cfg.Server, cfg.Client)
	}

	if cfg.HasClient() {
		t.Fatal("HasClient")
	}

	if cfg.CookieFile() != filepath.Join(state, config.AppName, config.Cookie) {
		t.Fatalf("cookie %s", cfg.CookieFile())
	}

	if cfg.ListenAddr() != config.DefaultAddr {
		t.Fatalf("listen %s", cfg.ListenAddr())
	}
}

func TestLoadTOMLAndEnvOverride(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, `
[account]
username = "alice"
cookie = "sess.json"

[server]
addr = "10.1.1.1:9200"
token = "server-file"
`)
	t.Setenv("NSK_USERNAME", "bob")
	t.Setenv("NSK_SERVER_ADDR", "10.1.1.1:9300")
	t.Setenv("NSK_SERVER_TOKEN", "server-env")

	cfg, err := config.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff("bob", cfg.Account.Username); diff != "" {
		t.Fatal(diff)
	}

	if cfg.Server.Addr != "10.1.1.1:9300" || cfg.Server.Token != "server-env" {
		t.Fatalf("server %#v", cfg.Server)
	}

	if cfg.CookieFile() != filepath.Join(cfg.StateDir, "sess.json") {
		t.Fatalf("cookie %s", cfg.CookieFile())
	}
}

func TestLoadClientEnv(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, `
[client]
url = "http://10.1.1.1:9200"
token = "from-file"
`)
	t.Setenv("NSK_CLIENT_URL", "http://10.1.1.2:9200")

	cfg, err := config.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Client.URL != "http://10.1.1.2:9200" || cfg.Client.Token != "from-file" {
		t.Fatalf("client %#v", cfg.Client)
	}

	if !cfg.HasClient() {
		t.Fatal("HasClient")
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, "listen = [")

	if _, err := config.LoadDir(dir); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestServerAndClientMutex(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, `
[server]
addr = "10.1.1.1:9200"
token = "s"

[client]
url = "http://10.1.1.1:9200"
token = "c"
`)

	_, err := config.LoadDir(dir)
	if err == nil || !strings.Contains(err.Error(), "[server] 与 [client]") {
		t.Fatalf("mutex: %v", err)
	}
}

func TestServerAndClientMutexFromEnv(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, "\n[client]\nurl = \"http://10.1.1.1:9200\"\n")
	t.Setenv("NSK_SERVER_TOKEN", "bind")

	if _, err := config.LoadDir(dir); err == nil {
		t.Fatal("expected mutex from env")
	}
}

func TestServerTokenOnlyDefaultsAddr(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	t.Setenv("NSK_SERVER_TOKEN", "bind")

	cfg, err := config.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Server.Token != "bind" || cfg.Server.Addr != config.DefaultAddr {
		t.Fatalf("%#v", cfg.Server)
	}
}

func TestLoadPathExplicit(t *testing.T) {
	isolateXDG(t)

	path := filepath.Join(t.TempDir(), "custom.toml")
	if err := os.WriteFile(path, []byte("\n[account]\nusername = \"abs\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadPath(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Account.Username != "abs" || cfg.Path != path {
		t.Fatalf("%#v", cfg)
	}
}

func TestLoadPathMissingErrors(t *testing.T) {
	isolateXDG(t)

	if _, err := config.LoadPath(filepath.Join(t.TempDir(), "nope.toml")); err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestLoadPathEmptyUsesXDG(t *testing.T) {
	xdg, _ := isolateXDG(t)
	writeConfig(t, filepath.Join(xdg, config.AppName), "\n[account]\nusername = \"xdg\"\n")

	cfg, err := config.LoadPath("")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Account.Username != "xdg" {
		t.Fatalf("%#v", cfg)
	}
}

func TestAbsoluteCookie(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	absCookie := filepath.Join(t.TempDir(), "abs.json")
	writeConfig(t, dir, "\n[account]\ncookie = \""+absCookie+"\"\n")

	cfg, err := config.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.CookieFile() != absCookie {
		t.Fatalf("abs %s", cfg.CookieFile())
	}
}

func TestLoadResolveDirError(t *testing.T) {
	isolateXDG(t)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Cleanup(config.StubUserHomeDir(func() (string, error) { return "", errNoHome }))

	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "无法解析 XDG 配置目录") {
		t.Fatalf("%v", err)
	}
}

func TestLoadResolveStateDirError(t *testing.T) {
	isolateXDG(t)
	t.Setenv("XDG_STATE_HOME", "")
	t.Cleanup(config.StubUserHomeDir(func() (string, error) { return "", errNoHome }))

	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "无法解析 XDG 状态目录") {
		t.Fatalf("%v", err)
	}
}

func TestLoadReadFileError(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(config.StubReadFile(func(string) ([]byte, error) { return nil, os.ErrPermission }))

	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, "\n[account]\nusername = \"alice\"\n")

	if _, err := config.LoadDir(dir); err == nil {
		t.Fatal("expected read error")
	}
}

func TestLoadPathAbsError(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(config.StubAbsPath(func(string) (string, error) { return "", errAbsPath }))

	_, err := config.LoadPath("relative.toml")
	if err == nil || !strings.Contains(err.Error(), "无效 --config") {
		t.Fatalf("%v", err)
	}
}
