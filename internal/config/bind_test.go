package config_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/solanab/nsk/internal/config"
)

func TestRejectUnspecifiedBind(t *testing.T) {
	isolateXDG(t)

	dir := filepath.Join(t.TempDir(), config.AppName)
	for _, addr := range []string{"0.0.0.0:9200", "[::]:9200"} {
		writeConfig(t, dir, "\n[server]\naddr = \""+addr+"\"\ntoken = \"t\"\n")

		_, err := config.LoadDir(dir)
		if err == nil || !strings.Contains(err.Error(), "拒绝绑定") {
			t.Fatalf("addr %s: %v", addr, err)
		}
	}
}

func TestRejectBareIPv6Unspecified(t *testing.T) {
	t.Parallel()

	_, err := config.NormalizeHost("::")
	if err == nil || !strings.Contains(err.Error(), "拒绝绑定") {
		t.Fatalf("%v", err)
	}
}

func TestNonLoopbackRequiresToken(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, "\n[server]\naddr = \"10.1.1.1:9200\"\n")

	_, err := config.LoadDir(dir)
	if err == nil || !strings.Contains(err.Error(), "必须设置 token") {
		t.Fatalf("%v", err)
	}
}

func TestLoopbackAllowsEmptyToken(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, "\n[server]\naddr = \"127.0.0.1:9200\"\n")

	cfg, err := config.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ListenAddr() != "127.0.0.1:9200" {
		t.Fatalf("%s", cfg.ListenAddr())
	}
}

func TestParseAddrEmptyHost(t *testing.T) {
	t.Parallel()

	host, port, err := config.ParseAddr(":9200")
	if err != nil {
		t.Fatal(err)
	}

	if host != "127.0.0.1" || port != 9200 {
		t.Fatalf("%s %d", host, port)
	}
}

func TestParseAddrInvalid(t *testing.T) {
	t.Parallel()

	if _, _, err := config.ParseAddr("not-an-addr"); err == nil {
		t.Fatal("expected error")
	}

	if _, _, err := config.ParseAddr("127.0.0.1:http"); err == nil {
		t.Fatal("expected port error")
	}
}

func TestValidateListenPort(t *testing.T) {
	t.Parallel()

	if err := config.ValidateListen("127.0.0.1", 0, ""); err == nil {
		t.Fatal("expected port error")
	}
}

func TestCanonicalListenOK(t *testing.T) {
	t.Parallel()

	got, err := config.CanonicalListen("127.0.0.1:9200", "")
	if err != nil {
		t.Fatal(err)
	}

	if got != "127.0.0.1:9200" {
		t.Fatalf("%s", got)
	}
}

func TestCanonicalListenRejects(t *testing.T) {
	t.Parallel()

	if _, err := config.CanonicalListen("not-an-addr", ""); err == nil {
		t.Fatal("expected error")
	}

	if _, err := config.CanonicalListen("10.1.1.1:9200", ""); err == nil {
		t.Fatal("expected token error")
	}
}

func TestLocalhostIsLoopback(t *testing.T) {
	isolateXDG(t)
	dir := filepath.Join(t.TempDir(), config.AppName)
	writeConfig(t, dir, "\n[server]\naddr = \"localhost:9200\"\n")

	if _, err := config.LoadDir(dir); err != nil {
		t.Fatal(err)
	}
}
