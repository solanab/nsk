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

func sampleList() *client.PostList {
	return &client.PostList{
		Page:    1,
		PerPage: client.ListPerPage,
		Posts: []client.PostSummary{
			{
				ID:       10,
				Title:    "hello",
				Category: slugTech,
				Author:   nameAlice,
				Replies:  2,
				URL:      "https://www.nodeseek.com/post-10-1",
			},
		},
	}
}

func stubList(t *testing.T, list *client.PostList, err error) *fakeAccount {
	t.Helper()

	fake := new(fakeAccount)
	fake.list = list
	fake.listErr = err
	stubAccount(t, fake)

	return fake
}

func TestListJSON(t *testing.T) {
	isolateXDG(t)
	stubList(t, sampleList(), nil)

	code, stdout, stderr := runCmd(t, []string{cmdList})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got client.PostList
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(sampleList(), &got); diff != "" {
		t.Fatal(diff)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestListText(t *testing.T) {
	isolateXDG(t)
	stubList(t, sampleList(), nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdList})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "id=10 title=hello author=alice category=tech\n" {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestListPageFlag(t *testing.T) {
	isolateXDG(t)
	fake := stubList(t, sampleList(), nil)

	code, _, stderr := runCmd(t, []string{cmdList, "--page", "2"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotPage != 2 || fake.gotSlug != "" {
		t.Fatalf("page=%d slug=%q", fake.gotPage, fake.gotSlug)
	}
}

func TestListSlugAndPage(t *testing.T) {
	isolateXDG(t)
	fake := stubList(t, sampleList(), nil)

	code, _, stderr := runCmd(t, []string{cmdList, slugTech, "--page", "3"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotSlug != slugTech || fake.gotPage != 3 {
		t.Fatalf("slug=%q page=%d", fake.gotSlug, fake.gotPage)
	}
}

func TestListMissingPageIsZero(t *testing.T) {
	isolateXDG(t)
	fake := stubList(t, sampleList(), nil)

	code, _, stderr := runCmd(t, []string{cmdList})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotPage != 0 {
		t.Fatalf("page=%d", fake.gotPage)
	}
}

func TestListErrorNoJSON(t *testing.T) {
	for _, args := range [][]string{{cmdList}, {cmdList, slugTech}} {
		isolateXDG(t)
		stubList(t, sampleList(), client.ErrNotFound)

		code, stdout, stderr := runCmd(t, args)
		if code != 1 {
			t.Fatalf("%v code=%d", args, code)
		}

		if !strings.Contains(stderr, "not found") {
			t.Fatalf("%v stderr=%q", args, stderr)
		}

		if stdout != "" {
			t.Fatalf("%v stdout=%q", args, stdout)
		}
	}
}

func TestListNilResult(t *testing.T) {
	isolateXDG(t)
	stubList(t, nil, nil)

	code, stdout, stderr := runCmd(t, []string{cmdList})
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

func TestListMarshalError(t *testing.T) {
	isolateXDG(t)
	stubList(t, sampleList(), nil)
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdList})
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

func TestListStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubList(t, sampleList(), nil)

	if code := app.Run([]string{cmdList}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestListTextStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubList(t, sampleList(), nil)

	if code := app.Run([]string{flagText, cmdList}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestListMissingCookieFile(t *testing.T) {
	_, state := isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{cmdList})
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

func TestListHelpSkipsAccount(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("list --help opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdList, flagHelp})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Usage: nsk list") {
		t.Fatalf("stdout=%q", stdout)
	}

	if strings.Contains(stdout, "{") {
		t.Fatalf("stdout leaked JSON: %q", stdout)
	}
}

func TestListEmptyText(t *testing.T) {
	isolateXDG(t)

	empty := new(client.PostList)
	empty.Page = 1
	empty.PerPage = client.ListPerPage
	empty.Posts = []client.PostSummary{}
	stubList(t, empty, nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdList})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "\n" {
		t.Fatalf("stdout=%q", stdout)
	}
}
