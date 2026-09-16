package client

import "strconv"

// LatestPosts returns one page of the site-wide latest list. page 0 is 1.
func (c *Client) LatestPosts(page int) (*PostList, error) {
	page = normalizePage(page)

	return c.fetchList(Site+"/page-"+strconv.Itoa(page), page)
}

// CategoryPosts returns one page of a board. Unknown slug is ErrNotFound without HTTP.
func (c *Client) CategoryPosts(slug string, page int) (*PostList, error) {
	if !knownSlug(slug) {
		return nil, ErrNotFound
	}

	page = normalizePage(page)
	rawURL := Site + "/categories/" + slug + "?page=" + strconv.Itoa(page)

	return c.fetchList(rawURL, page)
}

func (c *Client) fetchList(rawURL string, page int) (*PostList, error) {
	body, err := c.fetch(rawURL, "请求列表")
	if err != nil {
		return nil, err
	}

	return parsePostList(body, page)
}

func normalizePage(page int) int {
	if page == 0 {
		return 1
	}

	return page
}

func knownSlug(slug string) bool {
	for _, cat := range categoryTable() {
		if cat.Slug == slug {
			return true
		}
	}

	return false
}
