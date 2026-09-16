package app_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
)

const (
	cmdWhoami = "whoami"
	cmdCookie = "cookie"
	flagText  = "--text"
	nameAlice = "alice"
	nameBob   = "bob"
	levelTwo  = "Lv.2"
)

type fakeAccount struct {
	info    *client.UserInfo
	whoErr  error
	json    string
	jsonErr error
	pjwt    string
	pjwtErr error
}

func (fake *fakeAccount) WhoAmI() (*client.UserInfo, error) {
	return fake.info, fake.whoErr
}

func (fake *fakeAccount) ExportCookiesJSON() (string, error) {
	return fake.json, fake.jsonErr
}

func (fake *fakeAccount) ExportPJWT() (string, error) {
	return fake.pjwt, fake.pjwtErr
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

func writeTOML(t *testing.T, dir, body string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func stubAccount(t *testing.T, fake *fakeAccount) {
	t.Helper()
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		return fake, nil
	}))
}

func userFake(id int, name string, chicken int, level string) *fakeAccount {
	fake := new(fakeAccount)
	info := new(client.UserInfo)
	info.ID = id
	info.Name = name
	info.Chicken = chicken
	info.Level = level
	fake.info = info

	return fake
}

func runCmd(t *testing.T, args []string) (int, string, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer

	code := app.Run(args, &stdout, &stderr)

	return code, stdout.String(), stderr.String()
}

func TestWhoamiJSON(t *testing.T) {
	isolateXDG(t)
	stubAccount(t, userFake(42, nameAlice, 7, levelTwo))

	code, stdout, stderr := runCmd(t, []string{cmdWhoami})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got client.UserInfo
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}

	if got.ID != 42 || got.Name != nameAlice || got.Chicken != 7 || got.Level != levelTwo {
		t.Fatalf("%#v", got)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestWhoamiText(t *testing.T) {
	isolateXDG(t)
	stubAccount(t, userFake(42, nameAlice, 7, levelTwo))

	code, stdout, stderr := runCmd(t, []string{flagText, cmdWhoami})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "id=42 name=alice chicken=7 level=Lv.2\n" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestWhoamiTextMin(t *testing.T) {
	isolateXDG(t)
	stubAccount(t, userFake(1, nameBob, 0, ""))

	code, stdout, _ := runCmd(t, []string{flagText, cmdWhoami})
	if code != 0 {
		t.Fatalf("code=%d", code)
	}

	if stdout != "id=1 name=bob\n" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestCookieJSON(t *testing.T) {
	isolateXDG(t)

	fake := new(fakeAccount)
	fake.json = `[{"name":"pjwt","value":"tok"}]`
	stubAccount(t, fake)

	code, stdout, _ := runCmd(t, []string{cmdCookie})
	if code != 0 {
		t.Fatalf("code=%d stdout=%q", code, stdout)
	}

	if stdout != "[{\"name\":\"pjwt\",\"value\":\"tok\"}]\n" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestCookieOnly(t *testing.T) {
	isolateXDG(t)

	fake := new(fakeAccount)
	fake.pjwt = "pjwt=tok"
	stubAccount(t, fake)

	code, stdout, _ := runCmd(t, []string{cmdCookie, "--only"})
	if code != 0 {
		t.Fatalf("code=%d", code)
	}

	if stdout != "pjwt=tok\n" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestClientConfigFatal(t *testing.T) {
	xdg, state := isolateXDG(t)
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

	for _, args := range [][]string{{cmdWhoami}, {cmdCookie}} {
		code, stdout, stderr := runCmd(t, args)
		if code != 1 {
			t.Fatalf("%v code=%d", args, code)
		}

		if !strings.Contains(stderr, "nsk server 在后续票才可用") {
			t.Fatalf("%v stderr=%q", args, stderr)
		}

		if strings.Contains(stdout, "{") || strings.Contains(stderr, "无效 cookie") {
			t.Fatalf("%v read cookie: stdout=%q stderr=%q", args, stdout, stderr)
		}
	}
}

func TestMissingCookieFileCLI(t *testing.T) {
	_, state := isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{cmdWhoami})
	if code != 1 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	wantPath := filepath.Join(state, config.AppName, config.Cookie)
	if !strings.Contains(stderr, wantPath) {
		t.Fatalf("stderr=%q want path %s", stderr, wantPath)
	}

	if !strings.Contains(stderr, "Cookie-Editor JSON") || !strings.Contains(stderr, "pjwt=") {
		t.Fatalf("stderr=%q", stderr)
	}

	if strings.Contains(stdout, "{") {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestWhoamiErrorNoJSON(t *testing.T) {
	isolateXDG(t)

	fake := new(fakeAccount)
	fake.whoErr = client.ErrExpiredCookie
	stubAccount(t, fake)

	code, stdout, stderr := runCmd(t, []string{cmdWhoami})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "pjwt 无效或过期") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}
