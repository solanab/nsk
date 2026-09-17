package app_test

import (
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
)

const (
	notifyTitle = "奥特曼咋还不重置啊，出了这么大封号乌龙事件"
	notifyURL   = "https://www.nodeseek.com/post-763505-2#11"
)

func sampleNotify() []client.Notification {
	return []client.Notification{
		{
			PostID: 763505,
			Page:   2,
			Floor:  11,
			URL:    notifyURL,
			Text:   notifyTitle,
		},
	}
}

func stubNotify(t *testing.T, notes []client.Notification, err error) {
	t.Helper()

	fake := new(fakeAccount)
	fake.notify = notes
	fake.notifyErr = err
	stubAccount(t, fake)
}

func TestNotifyJSON(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, sampleNotify(), nil)

	code, stdout, stderr := runCmd(t, []string{cmdNotify})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got []client.Notification
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(sampleNotify(), got); diff != "" {
		t.Fatal(diff)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestNotifyJSONEmpty(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, []client.Notification{}, nil)

	code, stdout, stderr := runCmd(t, []string{cmdNotify})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "[]\n" {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestNotifyJSONNilSlice(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, nil, nil)

	code, stdout, stderr := runCmd(t, []string{cmdNotify})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "[]\n" {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestNotifyText(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, sampleNotify(), nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdNotify})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	want := "post_id=763505 page=2 floor=11 url=" + notifyURL + " text=" + notifyTitle + "\n"
	if stdout != want {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestNotifyTextMin(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, []client.Notification{{PostID: 7, URL: "https://www.nodeseek.com/post-7-1#0"}}, nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdNotify})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "post_id=7 url=https://www.nodeseek.com/post-7-1#0\n" {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestNotifyErrorNoJSON(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, sampleNotify(), client.ErrAuthRequired)

	code, stdout, stderr := runCmd(t, []string{cmdNotify})
	if code != 1 {
		t.Fatalf("code=%d", code)
	}

	if !strings.Contains(stderr, "需要登录") {
		t.Fatalf("stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestNotifyMarshalError(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, sampleNotify(), nil)
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdNotify})
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

func TestNotifyStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, sampleNotify(), nil)

	if code := app.Run([]string{cmdNotify}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestNotifyTextStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubNotify(t, sampleNotify(), nil)

	if code := app.Run([]string{flagText, cmdNotify}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestNotifyMissingCookieFile(t *testing.T) {
	_, state := isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{cmdNotify})
	if code != 1 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	wantPath := filepath.Join(state, config.AppName, config.Cookie)
	if !strings.Contains(stderr, wantPath) {
		t.Fatalf("stderr=%q want path %s", stderr, wantPath)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestNotifyHelpSkipsAccount(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("notify --help opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdNotify, flagHelp})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Usage: nsk notify") {
		t.Fatalf("stdout=%q", stdout)
	}

	if strings.Contains(stdout, "{") {
		t.Fatalf("stdout leaked JSON: %q", stdout)
	}
}

func TestNotifyVersionSkipsAccount(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("notify --version opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdNotify, flagVersion})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "nsk 0.0.0-dev") {
		t.Fatalf("stdout=%q", stdout)
	}

	if strings.Contains(stdout, "{") {
		t.Fatalf("stdout leaked JSON: %q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}
