package app_test

import (
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
)

var (
	errNoAccount  = staticError("no account")
	errExportJSON = staticError("export json")
	errExportPJWT = staticError("export pjwt")
)

type staticError string

func (err staticError) Error() string { return string(err) }

func TestWhoamiNilUser(t *testing.T) {
	isolateXDG(t)
	stubAccount(t, new(fakeAccount))

	code, stdout, stderr := runCmd(t, []string{cmdWhoami})
	if code != 1 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stderr, "pjwt 无效或过期") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestWhoamiMarshalError(t *testing.T) {
	isolateXDG(t)
	stubAccount(t, userFake(1, nameBob, 0, ""))
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdWhoami})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "编码 JSON") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestWhoamiStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubAccount(t, userFake(1, nameBob, 0, ""))

	if code := app.Run([]string{cmdWhoami}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestWhoamiTextStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubAccount(t, userFake(1, nameBob, 0, ""))

	if code := app.Run([]string{flagText, cmdWhoami}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestWhoamiStderrWriteError(t *testing.T) {
	isolateXDG(t)

	fake := new(fakeAccount)
	fake.whoErr = client.ErrExpiredCookie
	stubAccount(t, fake)

	if code := app.Run([]string{cmdWhoami}, io.Discard, &failWriter{remainingWrites: 0}); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestOpenAccountError(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdWhoami})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "no account") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestConfigLoadError(t *testing.T) {
	isolateXDG(t)

	missing := filepath.Join(t.TempDir(), "nope.toml")

	code, stdout, stderr := runCmd(t, []string{"--config", missing, cmdWhoami})
	if code != 1 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stderr, "找不到配置文件") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestCookieExportError(t *testing.T) {
	isolateXDG(t)

	fake := new(fakeAccount)
	fake.jsonErr = errExportJSON
	stubAccount(t, fake)

	code, stdout, stderr := runCmd(t, []string{cmdCookie})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "export json") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestCookieOnlyError(t *testing.T) {
	isolateXDG(t)

	fake := new(fakeAccount)
	fake.pjwtErr = errExportPJWT
	stubAccount(t, fake)

	code, stdout, stderr := runCmd(t, []string{cmdCookie, "--only"})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "export pjwt") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestWrapAccountOK(t *testing.T) {
	t.Parallel()

	_, err := app.WrapAccount(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCookieStdoutWriteError(t *testing.T) {
	isolateXDG(t)

	fake := new(fakeAccount)
	fake.json = "[]"
	stubAccount(t, fake)

	if code := app.Run([]string{cmdCookie}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}
