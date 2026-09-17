package server_test

import (
	"net/http/httptest"
	"testing"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/remote"
	"github.com/solanab/nsk/internal/server"
)

func TestRemoteRoundTrip(t *testing.T) {
	stub := newStub()
	stub.pages = map[int]*client.PostDetail{1: pageOne(), 2: pageTwo()}

	ts := httptest.NewServer(mustServer(t, server.Config{Forum: stub, Token: tokenSecret}).Handler())
	t.Cleanup(ts.Close)

	rem := remote.New(ts.URL, tokenSecret)
	if err := rem.Ping(); err != nil {
		t.Fatal(err)
	}

	mustWhoAmI(t, rem)
	mustUser(t, rem, stub)
	mustCats(t, rem)
	mustLists(t, rem, stub)
	assertRoundTripPostAll(t, rem)
}

func mustWhoAmI(t *testing.T, rem *remote.Client) {
	t.Helper()

	info, err := rem.WhoAmI()
	if err != nil || info.Name != nameAlice {
		t.Fatalf("whoami %#v %v", info, err)
	}
}

func mustUser(t *testing.T, rem *remote.Client, stub *stubForum) {
	t.Helper()

	if _, err := rem.GetUser(3); err != nil || stub.lastID != 3 {
		t.Fatalf("user %v", err)
	}
}

func mustCats(t *testing.T, rem *remote.Client) {
	t.Helper()

	cats, err := rem.Categories()
	if err != nil || len(cats) != 1 {
		t.Fatalf("cats %v", err)
	}
}

func mustLists(t *testing.T, rem *remote.Client, stub *stubForum) {
	t.Helper()

	if _, err := rem.LatestPosts(1); err != nil {
		t.Fatal(err)
	}

	if _, err := rem.CategoryPosts(slugTech, 1); err != nil || stub.lastSlug != slugTech {
		t.Fatalf("board %v", err)
	}

	if _, err := rem.Search(queryVPS, 1); err != nil || stub.lastQ != queryVPS {
		t.Fatalf("search %v", err)
	}

	notes, err := rem.Notifications()
	if err != nil || len(notes) != 1 {
		t.Fatalf("notes %v", err)
	}
}

func assertRoundTripPostAll(t *testing.T, rem *remote.Client) {
	t.Helper()

	all, err := rem.GetPostAll(9)
	if err != nil {
		t.Fatal(err)
	}

	if len(all.Floors) != 2 || all.Pages != 2 {
		t.Fatalf("all %#v", all)
	}

	if got := rem.FormatPost(all); got == "" {
		t.Fatal("format")
	}
}

func pageOne() *client.PostDetail {
	detail := new(client.PostDetail)
	detail.ID = 9
	detail.Title = titleHello
	detail.Page = 1
	detail.Pages = 2
	detail.Floors = []client.Floor{{Number: 0, Author: nameAlice, Markdown: "hi"}}

	return detail
}

func pageTwo() *client.PostDetail {
	detail := new(client.PostDetail)
	detail.ID = 9
	detail.Page = 2
	detail.Pages = 2
	detail.Floors = []client.Floor{{Number: 11, Author: "bob", Markdown: "re"}}

	return detail
}
