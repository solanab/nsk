package app_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
)

func TestStructureMarshalError(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdStructure})
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

func TestCatsMarshalError(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdCats})
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

func TestStructureCategoriesError(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(app.StubLoadCategories(func() ([]client.Category, error) {
		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdStructure})
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

func TestCatsCategoriesError(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(app.StubLoadCategories(func() ([]client.Category, error) {
		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdCats})
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

func TestStructureAndCatsSkipAccount(t *testing.T) {
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

	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("structure/cats opened Account")

		return nil, errNoAccount
	}))

	for _, args := range [][]string{
		nil,
		{cmdStructure},
		{cmdCats},
		{flagText, cmdStructure},
		{flagText, cmdCats},
	} {
		code, stdout, stderr := runCmd(t, args)
		if code != 0 {
			t.Fatalf("%v code=%d stderr=%q", args, code, stderr)
		}

		if strings.Contains(stderr, "nsk server 在后续票才可用") ||
			strings.Contains(stderr, "无效 cookie") {
			t.Fatalf("%v opened client: stderr=%q", args, stderr)
		}

		if stdout == "" {
			t.Fatalf("%v empty stdout", args)
		}
	}
}

func TestStructureNoConfig(t *testing.T) {
	isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{cmdStructure})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	assertSiteStructure(t, decodeStructure(t, stdout))

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestStructureStdoutWriteError(t *testing.T) {
	isolateXDG(t)

	if code := app.Run([]string{cmdStructure}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestCatsTextStdoutWriteError(t *testing.T) {
	isolateXDG(t)

	if code := app.Run([]string{flagText, cmdCats}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}
