package app_test

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
)

func TestPostErrorNoJSON(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), client.ErrNotFound)

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10"})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "not found") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestPostAllErrorNoJSON(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), client.ErrRateLimited)

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10", flagAll})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "rate limited") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestPostNilResult(t *testing.T) {
	isolateXDG(t)
	stubPost(t, nil, nil)

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10"})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "not found") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestPostMarshalError(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10"})
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

func TestPostOutputJSON(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	dir := t.TempDir()
	path := filepath.Join(dir, "out.md")

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10", "-o", path})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got client.SavedView
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json %v stdout=%q", err, stdout)
	}

	want := client.SavedView{Saved: path, ID: 10, Title: titleHello}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}

	body, err := os.ReadFile(path) //nolint:gosec // testdata path
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != client.FormatPost(samplePost()) {
		t.Fatalf("file=%q", body)
	}
}

func TestPostOutputDefault(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	t.Chdir(t.TempDir())

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10", "-o"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got client.SavedView
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}

	if got.Saved != "post-10.md" {
		t.Fatalf("saved=%q", got.Saved)
	}

	if _, err := os.Stat("post-10.md"); err != nil {
		t.Fatal(err)
	}
}

func TestPostOutputText(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	path := filepath.Join(t.TempDir(), "t.md")

	code, stdout, stderr := runCmd(t, []string{flagText, cmdPost, "10", "-o", path})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if strings.Contains(stdout, "{") {
		t.Fatalf("JSON SavedView on --text: %q", stdout)
	}

	if stdout != "saved "+path+" id=10 title="+titleHello+"\n" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestPostOutputMarshalErrorNoFile(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	path := filepath.Join(t.TempDir(), "nope.md")
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10", "-o", path})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "编码 JSON") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("wrote file: %v", err)
	}
}

func TestPostOutputWriteError(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	t.Cleanup(app.StubWriteFile(func(string, []byte, os.FileMode) error {
		return errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10", "-o", "x.md"})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "写入 Markdown") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestPostStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)

	if code := app.Run([]string{cmdPost, "10"}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestPostTextStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)

	if code := app.Run([]string{flagText, cmdPost, "10"}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestPostOutputTextWriteError(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	t.Cleanup(app.StubWriteFile(func(string, []byte, os.FileMode) error {
		return errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{flagText, cmdPost, "10", "-o"})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "写入 Markdown") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestPostMissingCookieFile(t *testing.T) {
	_, state := isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10"})
	if code != 1 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	wantPath := filepath.Join(state, config.AppName, config.Cookie)
	if !strings.Contains(stderr, wantPath) {
		t.Fatalf("stderr=%q want path %s", stderr, wantPath)
	}

	if strings.Contains(stdout, "{") {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestPostOutputStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	t.Chdir(t.TempDir())

	if code := app.Run([]string{cmdPost, "10", "-o"}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestPostOutputTextStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)
	t.Chdir(t.TempDir())

	if code := app.Run(
		[]string{flagText, cmdPost, "10", "-o"},
		&failWriter{remainingWrites: 0},
		io.Discard,
	); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestPostHelpSkipsAccount(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("post --help opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdPost, flagHelp})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Usage: nsk post") {
		t.Fatalf("stdout=%q", stdout)
	}

	if strings.Contains(stdout, `"floors"`) || strings.Contains(stdout, `"categories"`) {
		t.Fatalf("stdout leaked JSON: %q", stdout)
	}
}
