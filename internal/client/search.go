package client

import (
	"net/url"
	"strconv"
	"strings"
)

// Search fetches one logged-in search page. page 0 is 1.
// Guest HTML (no __config__.user) is ErrExpiredCookie; it never falls back to the latest list.
func (c *Client) Search(query string, page int) (*SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errEmptyQuery
	}

	page = normalizePage(page)

	body, err := c.fetch(searchURL(query, page), "请求搜索")
	if err != nil {
		return nil, err
	}

	return parseSearch(body, query, page)
}

func parseSearch(html []byte, query string, page int) (*SearchResult, error) {
	if parseUser(html) == nil {
		return nil, ErrExpiredCookie
	}

	list, err := parsePostList(html, page)
	if err != nil {
		return nil, err
	}

	result := new(SearchResult)
	result.Query = query
	result.Page = page
	result.Posts = list.Posts

	return result, nil
}

// searchURL is GET /search?q=... ; page>1 adds &page=n. q is QueryEscape'd.
func searchURL(query string, page int) string {
	rawURL := Site + "/search?q=" + url.QueryEscape(query)
	if page > 1 {
		rawURL += "&page=" + strconv.Itoa(page)
	}

	return rawURL
}
