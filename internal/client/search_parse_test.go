package client_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/client"
)

const (
	htmlSearchPage  = "search-page-1.html"
	htmlSearchGuest = "search-guest.html"
	htmlSearchEmpty = "search-empty.html"
	queryVPS        = "vps"
)

func wantSearchPosts() []client.PostSummary {
	return []client.PostSummary{
		{
			ID:       10,
			Title:    "pinned",
			Category: slugTech,
			Author:   nameAlice,
			Replies:  1,
			URL:      "https://www.nodeseek.com/post-10-1",
		},
		{
			ID:       11,
			Title:    "fresh",
			Category: slugDaily,
			Author:   nameBob,
			Replies:  2,
			URL:      "https://www.nodeseek.com/post-11-1",
		},
	}
}

func TestParseSearchPage1(t *testing.T) {
	t.Parallel()

	got, err := client.ParseSearch(fixture(t, htmlSearchPage), queryVPS, 1)
	if err != nil {
		t.Fatal(err)
	}

	want := &client.SearchResult{Query: queryVPS, Page: 1, Posts: wantSearchPosts()}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseSearchGuest(t *testing.T) {
	t.Parallel()

	got, err := client.ParseSearch(fixture(t, htmlSearchGuest), queryVPS, 1)
	if !errors.Is(err, client.ErrExpiredCookie) {
		t.Fatalf("err = %v", err)
	}

	if got != nil {
		t.Fatalf("guest search returned list: %#v", got)
	}
}

func TestParseSearchEmpty(t *testing.T) {
	t.Parallel()

	got, err := client.ParseSearch(fixture(t, htmlSearchEmpty), queryVPS, 2)
	if err != nil {
		t.Fatal(err)
	}

	if got.Query != queryVPS || got.Page != 2 {
		t.Fatalf("query=%q page=%d", got.Query, got.Page)
	}

	if got.Posts == nil || len(got.Posts) != 0 {
		t.Fatalf("posts=%#v", got.Posts)
	}
}

func TestParseSearchHTMLError(t *testing.T) {
	calls := 0

	t.Cleanup(client.StubParseHTML(func(reader io.Reader) (*goquery.Document, error) {
		calls++
		if calls == 1 {
			return goquery.NewDocumentFromReader(reader)
		}

		return nil, errHTML
	}))

	_, err := client.ParseSearch(fixture(t, htmlSearchPage), queryVPS, 1)
	if err == nil || !strings.Contains(err.Error(), "解析列表") {
		t.Fatalf("err = %v", err)
	}
}
