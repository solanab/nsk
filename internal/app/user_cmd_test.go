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

func sampleUser() *client.UserInfo {
	return &client.UserInfo{ID: 42, Name: nameAlice, Chicken: 7, Level: levelTwo}
}

func stubUser(t *testing.T, info *client.UserInfo, err error) *fakeAccount {
	t.Helper()

	fake := new(fakeAccount)
	fake.user = info
	fake.userErr = err
	stubAccount(t, fake)

	return fake
}

func TestUserJSON(t *testing.T) {
	isolateXDG(t)
	stubUser(t, sampleUser(), nil)

	code, stdout, stderr := runCmd(t, []string{cmdUser, "42"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got client.UserInfo
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(sampleUser(), &got); diff != "" {
		t.Fatal(diff)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestUserText(t *testing.T) {
	isolateXDG(t)
	stubUser(t, sampleUser(), nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdUser, "42"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "id=42 name=alice chicken=7 level=Lv.2\n" {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestUserTextMin(t *testing.T) {
	isolateXDG(t)
	stubUser(t, &client.UserInfo{ID: 1, Name: nameBob}, nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdUser, "1"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "id=1 name=bob\n" {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestUserPassesID(t *testing.T) {
	isolateXDG(t)
	fake := stubUser(t, sampleUser(), nil)

	code, _, stderr := runCmd(t, []string{cmdUser, "42"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotUserID != 42 {
		t.Fatalf("id=%d", fake.gotUserID)
	}
}

func TestUserErrorNoJSON(t *testing.T) {
	isolateXDG(t)
	stubUser(t, sampleUser(), client.ErrNotFound)

	code, stdout, stderr := runCmd(t, []string{cmdUser, "42"})
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

func TestUserNilResult(t *testing.T) {
	isolateXDG(t)
	stubUser(t, nil, nil)

	code, stdout, stderr := runCmd(t, []string{cmdUser, "42"})
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

func TestUserMarshalError(t *testing.T) {
	isolateXDG(t)
	stubUser(t, sampleUser(), nil)
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdUser, "42"})
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

func TestUserStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubUser(t, sampleUser(), nil)

	if code := app.Run([]string{cmdUser, "42"}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestUserTextStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubUser(t, sampleUser(), nil)

	if code := app.Run([]string{flagText, cmdUser, "42"}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestUserMissingCookieFile(t *testing.T) {
	_, state := isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{cmdUser, "42"})
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

func TestUserHelpSkipsAccount(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("user --help opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdUser, flagHelp})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Usage: nsk user") {
		t.Fatalf("stdout=%q", stdout)
	}

	if strings.Contains(stdout, "{") {
		t.Fatalf("stdout leaked JSON: %q", stdout)
	}
}

func TestUserVersionSkipsAccount(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("user --version opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdUser, flagVersion})
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

func TestUserInvalidIDNoHTTP(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("invalid user id opened Account")

		return nil, errNoAccount
	}))

	for _, raw := range []string{"alice", "12.3", "", "0"} {
		code, stdout, stderr := runCmd(t, []string{cmdUser, raw})
		if code != 1 {
			t.Fatalf("id=%q code=%d stderr=%q", raw, code, stderr)
		}

		if !strings.Contains(stderr, "无效用户 id") {
			t.Fatalf("id=%q stderr=%q", raw, stderr)
		}

		if stdout != "" {
			t.Fatalf("id=%q stdout=%q", raw, stdout)
		}
	}
}

func TestUserNegativeIDNoHTTP(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("negative user id opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdUser, "-1"})
	if code != 2 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "" {
		t.Fatalf("stdout=%q", stdout)
	}

	if !strings.Contains(stderr, "unknown flag") {
		t.Fatalf("stderr=%q", stderr)
	}

	code, stdout, stderr = runCmd(t, []string{cmdUser, "--", "-1"})
	if code != 1 {
		t.Fatalf("-- -1 code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stderr, "无效用户 id") {
		t.Fatalf("-- -1 stderr=%q", stderr)
	}

	if stdout != "" {
		t.Fatalf("-- -1 stdout=%q", stdout)
	}
}

func TestUserMissingID(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("user without id opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdUser})
	if code != 2 && code != 1 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "" && strings.Contains(stdout, "{") {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr == "" && !strings.Contains(stdout, "Usage:") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}
