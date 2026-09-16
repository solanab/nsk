package config_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solanab/nsk/internal/config"
)

var (
	errNoHome  = errors.New("no home")
	errAbsPath = errors.New("abs")
)

func TestResolveDirUsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")

	got, err := config.ResolveDir()
	if err != nil {
		t.Fatal(err)
	}

	if got != filepath.Join("/tmp/xdg-config", config.AppName) {
		t.Fatalf("got %s", got)
	}
}

func TestResolveDirDefaultDotConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := config.ResolveDir()
	if err != nil {
		t.Fatal(err)
	}

	if got != filepath.Join(home, ".config", config.AppName) {
		t.Fatalf("got %s", got)
	}
}

func TestResolveStateDirUsesXDGStateHome(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/xdg-state")

	got, err := config.ResolveStateDir()
	if err != nil {
		t.Fatal(err)
	}

	if got != filepath.Join("/tmp/xdg-state", config.AppName) {
		t.Fatalf("got %s", got)
	}
}

func TestResolveStateDirDefault(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := config.ResolveStateDir()
	if err != nil {
		t.Fatal(err)
	}

	if got != filepath.Join(home, ".local", "state", config.AppName) {
		t.Fatalf("got %s", got)
	}
}

func TestResolvePathDotUsesDefaultCookie(t *testing.T) {
	t.Parallel()

	if got := config.ResolvePath("/state", "."); got != "/state/cookie.json" {
		t.Fatalf("got %s", got)
	}

	if got := config.ResolvePath("/state", "  "); got != "/state/cookie.json" {
		t.Fatalf("empty %s", got)
	}
}

func TestFileExistsRejectsDirectory(t *testing.T) {
	t.Parallel()

	if config.FileExists(t.TempDir()) {
		t.Fatal("directory must not count as a file")
	}
}

func TestResolveDirHomeError(t *testing.T) {
	t.Cleanup(config.StubUserHomeDir(func() (string, error) { return "", errNoHome }))
	t.Setenv("XDG_CONFIG_HOME", "")

	_, err := config.ResolveDir()
	if err == nil || !strings.Contains(err.Error(), "无法解析 XDG 配置目录") {
		t.Fatalf("%v", err)
	}
}

func TestResolveStateDirHomeError(t *testing.T) {
	t.Cleanup(config.StubUserHomeDir(func() (string, error) { return "", errNoHome }))
	t.Setenv("XDG_STATE_HOME", "")

	_, err := config.ResolveStateDir()
	if err == nil || !strings.Contains(err.Error(), "无法解析 XDG 状态目录") {
		t.Fatalf("%v", err)
	}
}
