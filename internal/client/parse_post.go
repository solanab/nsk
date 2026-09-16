package client

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parsePostDetail(html []byte, postID, page int) (*PostDetail, error) {
	doc, err := loadSeam(&parseHTML)(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("解析帖子: %w", err)
	}

	floors := collectFloors(doc)
	detail := new(PostDetail)
	detail.ID = postID
	detail.Title = strings.TrimSpace(doc.Find("h1 a.post-title-link").First().Text())
	detail.Category = postCategory(doc)
	detail.Page = page
	detail.Pages = postPages(doc, page)
	detail.URL = postURL(postID, page)
	detail.Floors = floors
	detail.Author = firstAuthor(floors)

	return detail, nil
}

func collectFloors(doc *goquery.Document) []Floor {
	floors := make([]Floor, 0)

	doc.Find(".content-item").Each(func(_ int, sel *goquery.Selection) {
		floor, ok := parseFloor(sel)
		if ok {
			floors = append(floors, floor)
		}
	})

	return floors
}

func parseFloor(sel *goquery.Selection) (Floor, bool) {
	number, ok := floorNumber(sel)
	if !ok {
		return Floor{}, false
	}

	return Floor{
		Number:    number,
		Author:    strings.TrimSpace(sel.Find("a.author-name").First().Text()),
		CreatedAt: floorCreatedAt(sel),
		Markdown:  floorMarkdown(sel),
		ReplyTo:   floorReplyTo(sel),
	}, true
}

func floorNumber(sel *goquery.Selection) (int, bool) {
	text := strings.TrimSpace(sel.Find("a.floor-link").First().Text())
	text = strings.TrimPrefix(text, "#")

	if text == "" {
		return 0, false
	}

	number, err := strconv.Atoi(text)
	if err != nil {
		return 0, false
	}

	return number, true
}

func floorCreatedAt(sel *goquery.Selection) string {
	return strings.TrimSpace(sel.Find("span.date-created time[datetime]").First().AttrOr("datetime", ""))
}

func floorReplyTo(sel *goquery.Selection) *int {
	var reply *int

	sel.Find("article.post-content blockquote a[href]").EachWithBreak(func(_ int, node *goquery.Selection) bool {
		href := node.AttrOr("href", "")
		_, frag, found := strings.Cut(href, "#")

		if !found || frag == "" {
			return true
		}

		number, err := strconv.Atoi(frag)
		if err != nil {
			return true
		}

		reply = &number

		return false
	})

	return reply
}

func postCategory(doc *goquery.Document) string {
	href, ok := doc.Find(".content-category a[href^='/categories/']").First().Attr("href")
	if !ok {
		return ""
	}

	_, rest, _ := strings.Cut(href, "/categories/")
	rest, _, _ = strings.Cut(rest, "?")
	rest, _, _ = strings.Cut(rest, "#")

	return strings.Trim(rest, "/")
}

func postPages(doc *goquery.Document, page int) int {
	pages := max(page, 1)

	doc.Find(".nsk-pager [href]").Each(func(_ int, sel *goquery.Selection) {
		href := sel.AttrOr("href", "")
		if n := pageFromPostHref(href); n > pages {
			pages = n
		}
	})

	return pages
}

func pageFromPostHref(href string) int {
	_, rest, found := strings.Cut(href, "/post-")
	if !found {
		return 0
	}

	_, pageStr, found := strings.Cut(rest, "-")
	if !found {
		return 0
	}

	pageStr, _, _ = strings.Cut(pageStr, "#")
	pageStr, _, _ = strings.Cut(pageStr, "?")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return 0
	}

	return page
}

func firstAuthor(floors []Floor) string {
	if len(floors) == 0 {
		return ""
	}

	return floors[0].Author
}

func floorMarkdown(sel *goquery.Selection) string {
	return htmlToMarkdown(sel.Find("article.post-content").First())
}
