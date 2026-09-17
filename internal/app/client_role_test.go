package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
)

type fakeForum struct {
	*fakeAccount
}

func (*fakeForum) Categories() ([]client.Category, error) {
	return nil, nil
}

func (*fakeForum) FormatPost(*client.PostDetail) string {
	return ""
}

func TestCookieOnClientFatal(t *testing.T) {
	xdg, state := isolateXDG(t)
	writeClientTOML(t, xdg, state)

	for _, args := range [][]string{{cmdCookie}, {cmdCookie, flagOnly}} {
		code, stdout, stderr := runCmd(t, args)
		if code != 1 {
			t.Fatalf("%v code=%d", args, code)
		}

		if !strings.Contains(stderr, "client 机器不能导出 cookie") {
			t.Fatalf("%v stderr=%q", args, stderr)
		}

		if strings.Contains(stdout, "{") || strings.Contains(stderr, "无效 cookie") {
			t.Fatalf("%v read cookie: stdout=%q stderr=%q", args, stdout, stderr)
		}
	}
}

func TestClientCommandsSkipCookieFile(t *testing.T) {
	xdg, state := isolateXDG(t)
	writeClientTOML(t, xdg, state)
	t.Cleanup(app.StubOpenRemote(func(string, string) (app.Account, error) {
		return nil, staticError("server 未就绪")
	}))

	for _, args := range [][]string{
		{cmdWhoami},
		{cmdList},
		{cmdList, slugTech},
		{cmdPost, "1"},
		{cmdSearch, "vps"},
		{cmdUser, "1"},
		{cmdNotify},
	} {
		code, stdout, stderr := runCmd(t, args)
		if code != 1 {
			t.Fatalf("%v code=%d", args, code)
		}

		if !strings.Contains(stderr, "server 未就绪") {
			t.Fatalf("%v stderr=%q", args, stderr)
		}

		if strings.Contains(stdout, "{") || strings.Contains(stderr, "无效 cookie") {
			t.Fatalf("%v read cookie: stdout=%q stderr=%q", args, stdout, stderr)
		}
	}
}

func writeClientTOML(t *testing.T, xdg, state string) {
	t.Helper()

	writeTOML(t, filepath.Join(xdg, config.AppName), `
[client]
url = "http://127.0.0.1:9200"
`)

	cookiePath := filepath.Join(state, config.AppName, config.Cookie)
	if err := os.MkdirAll(filepath.Dir(cookiePath), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(cookiePath, []byte("not-a-cookie"), 0o600); err != nil {
		t.Fatal(err)
	}
}
