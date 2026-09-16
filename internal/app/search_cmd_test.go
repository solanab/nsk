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

const queryVPS = "vps"

func sampleSearch() *client.SearchResult {
	return &client.SearchResult{
		Query: queryVPS,
		Page:  1,
		Posts: []client.PostSummary{
			{
				ID:     10,
				Title:  titleHello,
				Author: nameAlice,
				URL:    "https://www.nodeseek.com/post-99-1",
			},
		},
	}
}

func stubSearch(t *testing.T, result *client.SearchResult, err error) *fakeAccount {
	t.Helper()

	fake := new(fakeAccount)
	fake.search = result
	fake.searchErr = err
	stubAccount(t, fake)

	return fake
}

func TestSearchJSON(t *testing.T) {
	isolateXDG(t)
	stubSearch(t, sampleSearch(), nil)

	code, stdout, stderr := runCmd(t, []string{cmdSearch, queryVPS})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got client.SearchResult
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(sampleSearch(), &got); diff != "" {
		t.Fatal(diff)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestSearchText(t *testing.T) {
	isolateXDG(t)
	stubSearch(t, sampleSearch(), nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdSearch, queryVPS})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	want := "query=vps page=1\nid=10 title=" + titleHello + " author=alice\n"
	if stdout != want {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestSearchPageFlag(t *testing.T) {
	isolateXDG(t)
	fake := stubSearch(t, sampleSearch(), nil)

	code, _, stderr := runCmd(t, []string{cmdSearch, queryVPS, flagPage, "2"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotQuery != queryVPS || fake.gotPage != 2 {
		t.Fatalf("query=%q page=%d", fake.gotQuery, fake.gotPage)
	}
}

func TestSearchMissingPageIsZero(t *testing.T) {
	isolateXDG(t)
	fake := stubSearch(t, sampleSearch(), nil)

	code, _, stderr := runCmd(t, []string{cmdSearch, queryVPS})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotPage != 0 {
		t.Fatalf("page=%d", fake.gotPage)
	}
}

func TestSearchTrimQuery(t *testing.T) {
	isolateXDG(t)
	fake := stubSearch(t, sampleSearch(), nil)

	code, _, stderr := runCmd(t, []string{cmdSearch, "  vps  "})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotQuery != queryVPS {
		t.Fatalf("query=%q", fake.gotQuery)
	}
}

func TestSearchErrorNoJSON(t *testing.T) {
	isolateXDG(t)
	stubSearch(t, sampleSearch(), client.ErrExpiredCookie)

	code, stdout, stderr := runCmd(t, []string{cmdSearch, queryVPS})
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

func TestSearchNilResult(t *testing.T) {
	isolateXDG(t)
	stubSearch(t, nil, nil)

	code, stdout, stderr := runCmd(t, []string{cmdSearch, queryVPS})
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

func TestSearchMarshalError(t *testing.T) {
	isolateXDG(t)
	stubSearch(t, sampleSearch(), nil)
	t.Cleanup(app.StubMarshalJSON(func(any) ([]byte, error) { return nil, errNoAccount }))

	code, stdout, stderr := runCmd(t, []string{cmdSearch, queryVPS})
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

func TestSearchStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubSearch(t, sampleSearch(), nil)

	if code := app.Run([]string{cmdSearch, queryVPS}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestSearchTextStdoutWriteError(t *testing.T) {
	isolateXDG(t)
	stubSearch(t, sampleSearch(), nil)

	if code := app.Run([]string{flagText, cmdSearch, queryVPS}, &failWriter{remainingWrites: 0}, io.Discard); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestSearchMissingCookieFile(t *testing.T) {
	_, state := isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{cmdSearch, queryVPS})
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

func TestSearchHelpSkipsAccount(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("search --help opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdSearch, flagHelp})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Usage: nsk search") {
		t.Fatalf("stdout=%q", stdout)
	}

	if strings.Contains(stdout, "{") {
		t.Fatalf("stdout leaked JSON: %q", stdout)
	}
}

func TestSearchEmptyText(t *testing.T) {
	isolateXDG(t)

	empty := new(client.SearchResult)
	empty.Query = queryVPS
	empty.Page = 1
	empty.Posts = []client.PostSummary{}
	stubSearch(t, empty, nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdSearch, queryVPS})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != "query=vps page=1\n" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestSearchEmptyQueryNoHTTP(t *testing.T) {
	isolateXDG(t)
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("empty query opened Account")

		return nil, errNoAccount
	}))

	for _, query := range []string{"", "  ", "\t"} {
		code, stdout, stderr := runCmd(t, []string{cmdSearch, query})
		if code != 1 {
			t.Fatalf("query=%q code=%d stderr=%q", query, code, stderr)
		}

		if !strings.Contains(stderr, "搜索词为空") {
			t.Fatalf("query=%q stderr=%q", query, stderr)
		}

		if stdout != "" {
			t.Fatalf("query=%q stdout=%q", query, stdout)
		}
	}
}

func TestSearchMissingQuery(t *testing.T) {
	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		t.Fatal("search without query opened Account")

		return nil, errNoAccount
	}))

	code, stdout, stderr := runCmd(t, []string{cmdSearch})
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
