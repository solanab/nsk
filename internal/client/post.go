package client

import "strconv"

// PageForFloor maps a 0-based floor number onto a 1-based post page.
// Floor 0 (OP) is page 1; page = floor/FloorsPerPage + 1.
func PageForFloor(floor int) int {
	if floor < 0 {
		return 0
	}

	return floor/FloorsPerPage + 1
}

// GetPost fetches one post page. page 0 is 1.
func (c *Client) GetPost(postID, page int) (*PostDetail, error) {
	page = normalizePage(page)

	body, err := c.fetch(postURL(postID, page), "请求帖子")
	if err != nil {
		return nil, err
	}

	return parsePostDetail(body, postID, page)
}

// GetPostAll concatenates floors from page 1 until an empty page or MaxPostPages.
// A mid-loop error fails the whole call; v1 does not retry 429.
func (c *Client) GetPostAll(postID int) (*PostDetail, error) {
	var out *PostDetail

	for page := 1; page <= MaxPostPages; page++ {
		detail, err := c.GetPost(postID, page)
		if err != nil {
			return nil, err
		}

		next, stop := mergePostPage(out, detail)
		out = next

		if stop {
			return out, nil
		}
	}

	return out, nil
}

func mergePostPage(out, detail *PostDetail) (*PostDetail, bool) {
	if len(detail.Floors) == 0 {
		if out == nil {
			return detail, true
		}

		return out, true
	}

	if out == nil {
		first := *detail
		first.Floors = append([]Floor{}, detail.Floors...)

		return &first, false
	}

	out.Floors = append(out.Floors, detail.Floors...)
	out.Pages = detail.Pages

	return out, false
}

func postURL(postID, page int) string {
	return Site + "/post-" + strconv.Itoa(postID) + "-" + strconv.Itoa(page)
}
