package client_test

import (
	"io"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/client"
)

const (
	htmlListPage    = "list-page-1.html"
	htmlListDedup   = "list-page-dedup.html"
	htmlListEmpty   = "list-page-empty.html"
	htmlListPartial = "list-page-partial.html"
)

func TestParseListPage1(t *testing.T) {
	t.Parallel()

	list, err := client.ParsePostList(fixture(t, htmlListPage), 1)
	if err != nil {
		t.Fatal(err)
	}

	if list.Page != 1 || list.PerPage != client.ListPerPage {
		t.Fatalf("page=%d per_page=%d", list.Page, list.PerPage)
	}

	if len(list.Posts) != client.ListPerPage {
		t.Fatalf("posts=%d want %d", len(list.Posts), client.ListPerPage)
	}

	wantFirst := client.PostSummary{
		ID:       703692,
		Title:    "【问👀】115非VIP账号海外上传速度如何 🍊",
		Category: slugDaily,
		Author:   "橘子海",
		Replies:  4,
		URL:      "https://www.nodeseek.com/post-703692-1",
	}
	if diff := cmp.Diff(wantFirst, list.Posts[0]); diff != "" {
		t.Fatal(diff)
	}

	for _, post := range list.Posts {
		if post.CreatedAt != "" {
			t.Fatalf("created_at mapped last-comment time: id=%d created_at=%q", post.ID, post.CreatedAt)
		}
	}

	wantSecond := client.PostSummary{
		ID:       703691,
		Title:    "溢价200收netcup rs1000 6.25o硬盘翻倍",
		Category: "trade",
		Author:   "codeman9527",
		Replies:  3,
		URL:      "https://www.nodeseek.com/post-703691-1",
	}
	if diff := cmp.Diff(wantSecond, list.Posts[1]); diff != "" {
		t.Fatal(diff)
	}
}

func TestListPerPagePinnedByFixture(t *testing.T) {
	t.Parallel()

	if client.ListPerPage != 49 {
		t.Fatalf("ListPerPage=%d", client.ListPerPage)
	}

	list, err := client.ParsePostList(fixture(t, htmlListPage), 1)
	if err != nil {
		t.Fatal(err)
	}

	if list.PerPage != 49 || len(list.Posts) != 49 {
		t.Fatalf("per_page=%d posts=%d", list.PerPage, len(list.Posts))
	}

	seen := make(map[int]struct{}, len(list.Posts))
	for _, post := range list.Posts {
		if _, dup := seen[post.ID]; dup {
			t.Fatalf("duplicate id %d", post.ID)
		}

		seen[post.ID] = struct{}{}
	}
}

func TestParseListDedup(t *testing.T) {
	t.Parallel()

	list, err := client.ParsePostList(fixture(t, htmlListDedup), 1)
	if err != nil {
		t.Fatal(err)
	}

	want := []client.PostSummary{
		{
			ID:       10,
			Title:    "pinned",
			Category: slugTech,
			Author:   "alice",
			Replies:  1,
			URL:      "https://www.nodeseek.com/post-10-1",
		},
		{
			ID:       11,
			Title:    "fresh",
			Category: slugDaily,
			Author:   "bob",
			Replies:  2,
			URL:      "https://www.nodeseek.com/post-11-1",
		},
	}
	if diff := cmp.Diff(want, list.Posts); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseListEmpty(t *testing.T) {
	t.Parallel()

	list, err := client.ParsePostList(fixture(t, htmlListEmpty), 2)
	if err != nil {
		t.Fatal(err)
	}

	if list.Page != 2 || list.PerPage != client.ListPerPage {
		t.Fatalf("page=%d per_page=%d", list.Page, list.PerPage)
	}

	if list.Posts == nil || len(list.Posts) != 0 {
		t.Fatalf("posts=%#v", list.Posts)
	}
}

func TestParseListPartial(t *testing.T) {
	t.Parallel()

	list, err := client.ParsePostList(fixture(t, htmlListPartial), 1)
	if err != nil {
		t.Fatal(err)
	}

	want := []client.PostSummary{
		{ID: 12, Title: "bare", URL: "https://www.nodeseek.com/post-12-1"},
		{ID: 13, Title: "abs", Author: "carol", Replies: 7, URL: "https://www.nodeseek.com/post-13-1"},
	}
	if diff := cmp.Diff(want, list.Posts); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseListCategoryQueryAndTextReplies(t *testing.T) {
	t.Parallel()

	html := `<html><ul class="post-list"><li class="post-list-item">` +
		`<div class="post-title"><a href="/post-99-1">x</a></div>` +
		`<div class="post-info"><span title="n/a" class="info-item info-comments-count">8</span>` +
		`<a href="/categories/life?page=2#top" class="info-item post-category">生活</a>` +
		`</div></li>` +
		`<li class="post-list-item"><div class="post-title"><a href="/post-98-1">y</a></div>` +
		`<div class="post-info"><a href="/board/tech" class="info-item post-category">技术</a>` +
		`<span class="info-item info-comments-count">n/a</span></div>` +
		`</li></ul></html>`

	list, err := client.ParsePostList([]byte(html), 1)
	if err != nil {
		t.Fatal(err)
	}

	want := []client.PostSummary{
		{ID: 99, Title: "x", Category: "life", Replies: 8, URL: "https://www.nodeseek.com/post-99-1"},
		{ID: 98, Title: "y", URL: "https://www.nodeseek.com/post-98-1"},
	}
	if diff := cmp.Diff(want, list.Posts); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseListCreatedAtSkipsLastComment(t *testing.T) {
	t.Parallel()

	html := `<html><ul class="post-list">` +
		`<li class="post-list-item"><div class="post-title"><a href="/post-50-1">x</a></div>` +
		`<time datetime="2026-01-02T03:04:05.000Z">created</time>` +
		`<a href="/post-50-1#4" class="info-item info-last-comment-time">` +
		`<time datetime="2026-04-27T05:46:32.000Z">24s ago</time></a></li>` +
		`<li class="post-list-item"><div class="post-title"><a href="/post-51-1">y</a></div>` +
		`<a class="info-item info-last-comment-time">` +
		`<time datetime="2026-04-27T05:46:32.000Z">24s ago</time></a></li>` +
		`</ul></html>`

	list, err := client.ParsePostList([]byte(html), 1)
	if err != nil {
		t.Fatal(err)
	}

	want := []client.PostSummary{
		{ID: 50, Title: "x", CreatedAt: "2026-01-02T03:04:05.000Z", URL: "https://www.nodeseek.com/post-50-1"},
		{ID: 51, Title: "y", URL: "https://www.nodeseek.com/post-51-1"},
	}
	if diff := cmp.Diff(want, list.Posts); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseListSkipsBadHref(t *testing.T) {
	t.Parallel()

	html := `<html><ul class="post-list">` +
		`<li class="post-list-item"><div class="post-title"><a href="/thread-1">x</a></div></li>` +
		`<li class="post-list-item"><div class="post-title"><a href="/post-">x</a></div></li>` +
		`<li class="post-list-item"><div class="post-title"><a href="/post-7">x</a></div></li>` +
		`</ul></html>`

	list, err := client.ParsePostList([]byte(html), 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(list.Posts) != 0 {
		t.Fatalf("posts=%#v", list.Posts)
	}
}

func TestParseListHTMLError(t *testing.T) {
	t.Cleanup(client.StubParseHTML(func(io.Reader) (*goquery.Document, error) {
		return nil, errHTML
	}))

	_, err := client.ParsePostList([]byte("<html></html>"), 1)
	if err == nil || !strings.Contains(err.Error(), "解析列表") {
		t.Fatalf("err = %v", err)
	}
}
