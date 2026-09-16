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
	htmlPostPage    = "post-703863-1.html"
	htmlPostEmpty   = "post-empty.html"
	htmlPostPage2   = "post-page-2.html"
	htmlPostANSI    = "post-ansi.html"
	htmlPostMagic   = "post-magic.html"
	htmlPostPartial = "post-partial.html"
	postTitle703863 = "绿云抢鸡竞赛又要开始了，一波传家宝又要来袭"
	postID703863    = 703863
)

func TestFloorsPerPagePinnedByFixture(t *testing.T) {
	t.Parallel()

	if client.FloorsPerPage != 11 {
		t.Fatalf("FloorsPerPage=%d", client.FloorsPerPage)
	}

	detail, err := client.ParsePostDetail(fixture(t, htmlPostPage), postID703863, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(detail.Floors) != 11 {
		t.Fatalf("floors=%d", len(detail.Floors))
	}
}

func TestParsePostPage1(t *testing.T) {
	t.Parallel()

	detail, err := client.ParsePostDetail(fixture(t, htmlPostPage), postID703863, 1)
	if err != nil {
		t.Fatal(err)
	}

	assertPostPage1Meta(t, detail)
	assertPostPage1Floors(t, detail)
}

func assertPostPage1Meta(t *testing.T, detail *client.PostDetail) {
	t.Helper()

	if detail.ID != postID703863 || detail.Page != 1 || detail.Pages != 4 {
		t.Fatalf("id=%d page=%d pages=%d", detail.ID, detail.Page, detail.Pages)
	}

	if detail.Title != postTitle703863 || detail.Category != slugDaily || detail.Author != "ipv4" {
		t.Fatalf("title=%q category=%q author=%q", detail.Title, detail.Category, detail.Author)
	}

	if detail.URL != "https://www.nodeseek.com/post-703863-1" {
		t.Fatalf("url=%q", detail.URL)
	}
}

func assertPostPage1Floors(t *testing.T, detail *client.PostDetail) {
	t.Helper()

	wantNums := []int{0, 4, 1, 2, 3, 5, 6, 7, 8, 9, 10}
	gotNums := make([]int, len(detail.Floors))

	for i, floor := range detail.Floors {
		gotNums[i] = floor.Number
	}

	if diff := cmp.Diff(wantNums, gotNums); diff != "" {
		t.Fatal(diff)
	}

	first := detail.Floors[0]
	if first.Author != "ipv4" || first.CreatedAt != "2026-04-27T07:57:00.000Z" {
		t.Fatalf("%#v", first)
	}

	if !strings.Contains(first.Markdown, "[Sticker]") {
		t.Fatalf("sticker missing: %q", first.Markdown)
	}

	if !strings.Contains(first.Markdown, "喊上你的五指小姐姐一起抢吧") {
		t.Fatalf("body missing: %q", first.Markdown)
	}

	if diff := cmp.Diff(new(2), detail.Floors[7].ReplyTo); diff != "" {
		t.Fatal(diff)
	}
}

func TestParsePostEmpty(t *testing.T) {
	t.Parallel()

	detail, err := client.ParsePostDetail(fixture(t, htmlPostEmpty), 9, 2)
	if err != nil {
		t.Fatal(err)
	}

	if detail.Title != "empty page" || detail.Page != 2 || detail.Pages != 2 {
		t.Fatalf("%#v", detail)
	}

	if detail.Floors == nil || len(detail.Floors) != 0 {
		t.Fatalf("floors=%#v", detail.Floors)
	}

	if detail.Author != "" {
		t.Fatalf("author=%q", detail.Author)
	}
}

func TestParsePostPartial(t *testing.T) {
	t.Parallel()

	detail, err := client.ParsePostDetail(fixture(t, htmlPostPartial), 6, 1)
	if err != nil {
		t.Fatal(err)
	}

	assertPartialFirst(t, detail)
	assertPartialRest(t, detail)
}

func assertPartialFirst(t *testing.T, detail *client.PostDetail) {
	t.Helper()

	if len(detail.Floors) != 3 {
		t.Fatalf("floors=%d", len(detail.Floors))
	}

	first := detail.Floors[0]
	if first.Number != 1 || first.CreatedAt != "2026-02-02T00:00:00.000Z" {
		t.Fatalf("%#v", first)
	}

	if diff := cmp.Diff(new(0), first.ReplyTo); diff != "" {
		t.Fatal(diff)
	}

	if !strings.Contains(first.Markdown, "[Sticker]") ||
		!strings.Contains(first.Markdown, "![pic](https://cdn.example/a.png)") {
		t.Fatalf("markdown=%q", first.Markdown)
	}
}

func assertPartialRest(t *testing.T, detail *client.PostDetail) {
	t.Helper()

	if detail.Floors[2].Number != 3 || detail.Floors[2].Markdown != "" {
		t.Fatalf("empty article: %#v", detail.Floors[2])
	}

	if strings.Contains(detail.Floors[1].Markdown, "\x1b") ||
		strings.Contains(detail.Floors[1].Markdown, "[31m") {
		t.Fatalf("ansi left in markdown: %q", detail.Floors[1].Markdown)
	}
}

func TestParsePostMagicTabs(t *testing.T) {
	t.Parallel()

	detail, err := client.ParsePostDetail(fixture(t, htmlPostMagic), 7, 1)
	if err != nil {
		t.Fatal(err)
	}

	if detail.Category != "review" {
		t.Fatalf("category=%q", detail.Category)
	}

	markdown := detail.Floors[0].Markdown
	if !strings.Contains(markdown, "```\n基本信息\nhello tab\n```") {
		t.Fatalf("tab1 missing: %q", markdown)
	}

	if !strings.Contains(markdown, "world") || !strings.Contains(markdown, "[Sticker]") {
		t.Fatalf("tab2 missing: %q", markdown)
	}

	if !strings.Contains(markdown, "after") {
		t.Fatalf("trailing text missing: %q", markdown)
	}
}

func TestParsePostANSIFixture(t *testing.T) {
	t.Parallel()

	detail, err := client.ParsePostDetail(fixture(t, htmlPostANSI), 8, 1)
	if err != nil {
		t.Fatal(err)
	}

	markdown := detail.Floors[0].Markdown
	if strings.Contains(markdown, "\x1b") || strings.Contains(markdown, "[31m") {
		t.Fatalf("ansi left: %q", markdown)
	}

	if !strings.Contains(markdown, "speed ") || !strings.Contains(markdown, "red") ||
		!strings.Contains(markdown, "done") {
		t.Fatalf("text missing: %q", markdown)
	}
}

func TestParsePostHTMLError(t *testing.T) {
	t.Cleanup(client.StubParseHTML(func(io.Reader) (*goquery.Document, error) {
		return nil, errHTML
	}))

	_, err := client.ParsePostDetail([]byte("<html></html>"), 1, 1)
	if err == nil || !strings.Contains(err.Error(), "解析帖子") {
		t.Fatalf("err = %v", err)
	}
}

func TestParsePostSkipsBadPagerAndReply(t *testing.T) {
	t.Parallel()

	html := `<html><h1><a class="post-title-link">t</a></h1>
<div class="content-category"><a href="/board/tech">x</a></div>
<div class="nsk-pager"><a href="/thread-1">1</a><a href="/post-5">2</a>
<a href="/post-5-x">3</a><span href="/post-5-3">3</span></div>
<div class="content-item"><a class="floor-link">#0</a>
<a class="author-name">a</a>
<article class="post-content"><blockquote><p>
<a href="/member?t=x">@x</a> <a href="/post-5-1#zz">bad</a>
</p></blockquote><p>body</p><a>plain</a><div>inner</div>
<pre><code>raw</code></pre><span>x</span><a href="/u">name</a></article></div>
<div class="content-item"><a class="floor-link">#nope</a></div>
</html>`

	detail, err := client.ParsePostDetail([]byte(html), 5, 1)
	if err != nil {
		t.Fatal(err)
	}

	if detail.Category != "" {
		t.Fatalf("category=%q", detail.Category)
	}

	if detail.Pages != 3 {
		t.Fatalf("pages=%d", detail.Pages)
	}

	if len(detail.Floors) != 1 || detail.Floors[0].ReplyTo != nil {
		t.Fatalf("%#v", detail.Floors)
	}

	markdown := detail.Floors[0].Markdown
	if !strings.Contains(markdown, "plain") || !strings.Contains(markdown, "inner") ||
		!strings.Contains(markdown, "```\nraw\n```") || !strings.Contains(markdown, "[name](/u)") {
		t.Fatalf("markdown=%q", markdown)
	}
}

func TestPageForFloor(t *testing.T) {
	t.Parallel()

	if got := client.PageForFloor(0); got != 1 {
		t.Fatalf("floor 0 -> %d", got)
	}

	if got := client.PageForFloor(10); got != 1 {
		t.Fatalf("floor 10 -> %d", got)
	}

	if got := client.PageForFloor(11); got != 2 {
		t.Fatalf("floor 11 -> %d", got)
	}

	if got := client.PageForFloor(-1); got != 0 {
		t.Fatalf("floor -1 -> %d", got)
	}
}
