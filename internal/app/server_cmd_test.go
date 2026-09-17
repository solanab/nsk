package app_test

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
	"github.com/solanab/nsk/internal/server"
)

func TestServerHelpSkipsAccount(t *testing.T) {
	t.Cleanup(app.StubOpenForum(func(string) (client.Forum, error) {
		t.Fatal("server --help opened Forum")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdServer, flagHelp})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Usage: nsk server") {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestServerOnClientFatal(t *testing.T) {
	xdg, state := isolateXDG(t)
	writeClientTOML(t, xdg, state)

	code, stdout, stderr := runCmd(t, []string{cmdServer})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "client 机器不能跑 nsk server") {
		t.Fatalf("stderr=%q", stderr)
	}

	if strings.Contains(stderr, "无效 cookie") || stdout != "" {
		t.Fatalf("stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestServerMissingCookie(t *testing.T) {
	_, state := isolateXDG(t)

	code, _, stderr := runCmd(t, []string{cmdServer})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	wantPath := filepath.Join(state, config.AppName, config.Cookie)
	if !strings.Contains(stderr, wantPath) {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestServerTokenFlag(t *testing.T) {
	isolateXDG(t)
	stubCmdForum(t)
	t.Cleanup(app.StubServeHTTP(func(*server.Server) error {
		return nil
	}))

	code, _, stderr := runCmd(t, []string{cmdServer, "--token", "cli-tok"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestServerListenStub(t *testing.T) {
	isolateXDG(t)
	stubCmdForum(t)

	var gotAddr string

	t.Cleanup(app.StubServeHTTP(func(srv *server.Server) error {
		gotAddr = srv.Addr()

		return nil
	}))

	code, stdout, stderr := runCmd(t, []string{cmdServer, "--addr", "127.0.0.1:9201"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}

	if gotAddr != "127.0.0.1:9201" || !strings.Contains(stderr, "http://127.0.0.1:9201") {
		t.Fatalf("addr=%s stderr=%q", gotAddr, stderr)
	}
}

func TestServerBindError(t *testing.T) {
	isolateXDG(t)
	stubCmdForum(t)

	code, _, stderr := runCmd(t, []string{cmdServer, "--addr", "0.0.0.0:9200"})
	if code != 1 || !strings.Contains(stderr, "拒绝绑定") {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestServerServeError(t *testing.T) {
	isolateXDG(t)
	stubCmdForum(t)
	t.Cleanup(app.StubServeHTTP(func(*server.Server) error {
		return staticError("listen fail")
	}))

	code, _, stderr := runCmd(t, []string{cmdServer})
	if code != 1 || !strings.Contains(stderr, "listen fail") {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestServerStderrWriteError(t *testing.T) {
	isolateXDG(t)
	stubCmdForum(t)

	if code := app.Run([]string{cmdServer}, io.Discard, &failWriter{remainingWrites: 0}); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestServerConfigAddrAndToken(t *testing.T) {
	xdg, _ := isolateXDG(t)
	writeTOML(t, filepath.Join(xdg, config.AppName), `
[server]
addr = "127.0.0.1:9202"
token = "from-toml"
`)
	stubCmdForum(t)
	t.Cleanup(app.StubServeHTTP(func(srv *server.Server) error {
		if srv.Addr() != "127.0.0.1:9202" {
			t.Fatalf("%s", srv.Addr())
		}

		return nil
	}))

	code, _, stderr := runCmd(t, []string{cmdServer})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestServerConfigLoadError(t *testing.T) {
	isolateXDG(t)

	missing := filepath.Join(t.TempDir(), "nope.toml")

	code, _, stderr := runCmd(t, []string{flagConfig, missing, cmdServer})
	if code != 1 || !strings.Contains(stderr, "找不到配置文件") {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestPingRemote(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", func(writer http.ResponseWriter, _ *http.Request) {
		encodeHealth(t, writer, true, true)
	})
	healthServer := httptest.NewServer(mux)
	t.Cleanup(healthServer.Close)

	acc, err := app.PingRemote(healthServer.URL, "tok")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := acc.ExportCookiesJSON(); err == nil || !strings.Contains(err.Error(), "不能导出 cookie") {
		t.Fatalf("%v", err)
	}

	if _, err := acc.ExportPJWT(); err == nil || !strings.Contains(err.Error(), "不能导出 cookie") {
		t.Fatalf("%v", err)
	}

	down := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		encodeHealth(t, writer, false, false)
	}))
	t.Cleanup(down.Close)

	if _, err := app.PingRemote(down.URL, ""); err == nil {
		t.Fatal("expected not ready")
	}
}

func encodeHealth(t *testing.T, writer http.ResponseWriter, ok, ready bool) {
	t.Helper()

	if err := json.NewEncoder(writer).Encode(map[string]any{"ok": ok, "ready": ready}); err != nil {
		t.Fatal(err)
	}
}

func TestOpenLocalForumError(t *testing.T) {
	if _, err := app.OpenLocalForum(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected error")
	}
}

func TestWrapForumOK(t *testing.T) {
	t.Parallel()

	if _, err := app.WrapForum(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func stubCmdForum(t *testing.T) {
	t.Helper()
	t.Cleanup(app.StubOpenForum(func(string) (client.Forum, error) {
		return &fakeForum{fakeAccount: new(fakeAccount)}, nil
	}))
}

func TestServeListenOK(t *testing.T) {
	t.Cleanup(server.StubListenHTTP(func(*http.Server) error {
		return nil
	}))

	forum := &fakeForum{fakeAccount: new(fakeAccount)}

	srv, err := server.New(server.Config{Forum: forum})
	if err != nil {
		t.Fatal(err)
	}

	if err := app.ServeListen(srv); err != nil {
		t.Fatal(err)
	}
}

func TestServeListenError(t *testing.T) {
	cfg := new(net.ListenConfig)

	listener, err := cfg.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Error(err)
		}
	})

	forum := &fakeForum{fakeAccount: new(fakeAccount)}

	srv, err := server.New(server.Config{Addr: listener.Addr().String(), Forum: forum})
	if err != nil {
		t.Fatal(err)
	}

	if err := app.ServeListen(srv); err == nil {
		t.Fatal("expected in-use error")
	}
}

func TestClientWhoamiRemoteJSON(t *testing.T) {
	xdg, _ := isolateXDG(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", func(writer http.ResponseWriter, _ *http.Request) {
		encodeHealth(t, writer, true, true)
	})
	mux.HandleFunc("/api/v1/whoami", func(writer http.ResponseWriter, _ *http.Request) {
		info := new(client.UserInfo)
		info.ID = 8

		info.Name = "remote-alice"
		if err := json.NewEncoder(writer).Encode(info); err != nil {
			t.Fatal(err)
		}
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	writeTOML(t, filepath.Join(xdg, config.AppName), `
[client]
url = "`+ts.URL+`"
token = "tok"
`)

	code, stdout, stderr := runCmd(t, []string{cmdWhoami})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got client.UserInfo
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}

	if got.Name != "remote-alice" {
		t.Fatalf("%#v", got)
	}
}
