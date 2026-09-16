package app_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
)

func samplePost() *client.PostDetail {
	return &client.PostDetail{
		ID:       10,
		Title:    titleHello,
		Category: slugTech,
		Author:   nameAlice,
		Page:     1,
		Pages:    2,
		URL:      "https://www.nodeseek.com/post-10-1",
		Floors: []client.Floor{
			{Number: 0, Author: nameAlice, Markdown: "op body"},
			{Number: 1, Author: nameBob, Markdown: "reply"},
		},
	}
}

func stubPost(t *testing.T, detail *client.PostDetail, err error) *fakeAccount {
	t.Helper()

	fake := new(fakeAccount)
	fake.post = detail
	fake.postErr = err
	stubAccount(t, fake)

	return fake
}

func TestPostJSON(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)

	code, stdout, stderr := runCmd(t, []string{cmdPost, "10"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	var got client.PostDetail
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(samplePost(), &got); diff != "" {
		t.Fatal(diff)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestPostText(t *testing.T) {
	isolateXDG(t)
	stubPost(t, samplePost(), nil)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdPost, "10"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	want := client.FormatPost(samplePost()) + "\n"
	if stdout != want {
		t.Fatalf("stdout=%q want=%q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestPostPageFlag(t *testing.T) {
	isolateXDG(t)
	fake := stubPost(t, samplePost(), nil)

	code, _, stderr := runCmd(t, []string{cmdPost, "10", flagPage, "2"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotID != 10 || fake.gotPage != 2 || fake.gotAll {
		t.Fatalf("id=%d page=%d all=%v", fake.gotID, fake.gotPage, fake.gotAll)
	}
}

func TestPostAll(t *testing.T) {
	isolateXDG(t)
	fake := stubPost(t, samplePost(), nil)

	code, _, stderr := runCmd(t, []string{cmdPost, "10", flagAll})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if !fake.gotAll || fake.gotID != 10 {
		t.Fatalf("all=%v id=%d", fake.gotAll, fake.gotID)
	}
}

func TestPostFloorPage(t *testing.T) {
	isolateXDG(t)
	fake := stubPost(t, samplePost(), nil)

	code, _, stderr := runCmd(t, []string{cmdPost, "10/11"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotID != 10 || fake.gotPage != 2 || fake.gotAll {
		t.Fatalf("id=%d page=%d all=%v", fake.gotID, fake.gotPage, fake.gotAll)
	}
}

func TestPostFloorZero(t *testing.T) {
	isolateXDG(t)
	fake := stubPost(t, samplePost(), nil)

	code, _, stderr := runCmd(t, []string{cmdPost, "10/0"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if fake.gotPage != 1 {
		t.Fatalf("page=%d", fake.gotPage)
	}
}

func TestPostInvalidRefNoAccount(t *testing.T) {
	isolateXDG(t)

	opened := false

	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		opened = true

		return nil, errNoAccount
	}))

	for _, args := range [][]string{
		{cmdPost, "abc"},
		{cmdPost, "0"},
		{cmdPost, "10/"},
		{cmdPost, "10/-1"},
		{cmdPost, "10/x"},
		{cmdPost, "10/1/2"},
	} {
		opened = false

		code, stdout, stderr := runCmd(t, args)
		if code != 1 {
			t.Fatalf("%v code=%d", args, code)
		}

		if opened {
			t.Fatalf("%v opened Account", args)
		}

		if stdout != "" {
			t.Fatalf("%v stdout=%q", args, stdout)
		}

		if !strings.Contains(stderr, "无效帖子 id") && !strings.Contains(stderr, "无效楼层") {
			t.Fatalf("%v stderr=%q", args, stderr)
		}
	}
}

func TestPostFlagConflicts(t *testing.T) {
	isolateXDG(t)

	opened := false

	t.Cleanup(app.StubOpenAccount(func(string) (app.Account, error) {
		opened = true

		return nil, errNoAccount
	}))

	tests := []struct {
		args []string
		want string
	}{
		{args: []string{cmdPost, "10", flagPage, "2", flagAll}, want: "--page 与 --all"},
		{args: []string{cmdPost, "10/1", flagAll}, want: "id/floor 与 --all"},
		{args: []string{cmdPost, "10/1", flagPage, "2"}, want: "id/floor 与 --page"},
	}
	for _, test := range tests {
		opened = false

		code, stdout, stderr := runCmd(t, test.args)
		if code != 1 {
			t.Fatalf("%v code=%d", test.args, code)
		}

		if opened {
			t.Fatalf("%v opened Account", test.args)
		}

		if !strings.Contains(stderr, test.want) {
			t.Fatalf("%v stderr=%q", test.args, stderr)
		}

		if stdout != "" {
			t.Fatalf("%v stdout=%q", test.args, stdout)
		}
	}
}
