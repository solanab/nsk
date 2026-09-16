package client

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parsePostList(html []byte, page int) (*PostList, error) {
	doc, err := loadSeam(&parseHTML)(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("解析列表: %w", err)
	}

	list := new(PostList)
	list.Page = page
	list.PerPage = ListPerPage
	list.Posts = collectPosts(doc)

	return list, nil
}

func collectPosts(doc *goquery.Document) []PostSummary {
	seen := make(map[int]struct{})
	posts := make([]PostSummary, 0)

	doc.Find("ul.post-list > li.post-list-item").Each(func(_ int, sel *goquery.Selection) {
		posts = appendUnique(posts, seen, sel)
	})

	return posts
}

func appendUnique(posts []PostSummary, seen map[int]struct{}, sel *goquery.Selection) []PostSummary {
	post, ok := parsePostItem(sel)
	if !ok {
		return posts
	}

	if _, dup := seen[post.ID]; dup {
		return posts
	}

	seen[post.ID] = struct{}{}

	return append(posts, post)
}

func parsePostItem(sel *goquery.Selection) (PostSummary, bool) {
	href, ok := sel.Find("div.post-title > a").First().Attr("href")
	if !ok {
		return PostSummary{}, false
	}

	postID := postIDFromHref(href)
	if postID == 0 {
		return PostSummary{}, false
	}

	return PostSummary{
		ID:        postID,
		Title:     strings.TrimSpace(sel.Find("div.post-title > a").First().Text()),
		Category:  categorySlug(sel),
		Author:    strings.TrimSpace(sel.Find(".info-author a").First().Text()),
		Replies:   parseReplies(sel),
		CreatedAt: createdAt(sel),
		URL:       Site + "/post-" + strconv.Itoa(postID) + "-1",
	}, true
}

func createdAt(sel *goquery.Selection) string {
	stamp := ""

	sel.Find("time[datetime]").EachWithBreak(func(_ int, node *goquery.Selection) bool {
		if node.Closest(".info-last-comment-time").Length() > 0 {
			return true
		}

		stamp = strings.TrimSpace(node.AttrOr("datetime", ""))

		return false
	})

	return stamp
}

func postIDFromHref(href string) int {
	_, rest, found := strings.Cut(href, "/post-")
	if !found {
		return 0
	}

	idStr, _, found := strings.Cut(rest, "-")
	if !found || idStr == "" {
		return 0
	}

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		return 0
	}

	return postID
}

func categorySlug(sel *goquery.Selection) string {
	href, ok := sel.Find("a.post-category").First().Attr("href")
	if !ok {
		return ""
	}

	_, rest, found := strings.Cut(href, "/categories/")
	if !found {
		return ""
	}

	rest, _, _ = strings.Cut(rest, "?")
	rest, _, _ = strings.Cut(rest, "#")

	return strings.Trim(rest, "/")
}

func parseReplies(sel *goquery.Selection) int {
	node := sel.Find(".info-comments-count").First()
	if n, ok := commentsTitle(node); ok {
		return n
	}

	n, err := strconv.Atoi(strings.TrimSpace(node.Text()))
	if err != nil {
		return 0
	}

	return n
}

func commentsTitle(node *goquery.Selection) (int, bool) {
	title, ok := node.Attr("title")
	if !ok {
		return 0, false
	}

	n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(title), " comments"))
	if err != nil {
		return 0, false
	}

	return n, true
}
