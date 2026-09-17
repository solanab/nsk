package remote

import (
	"net/url"
	"strconv"

	"github.com/solanab/nsk/internal/client"
)

// WhoAmI implements Forum.
func (c *Client) WhoAmI() (*client.UserInfo, error) {
	return getJSON[client.UserInfo](c, "/api/v1/whoami", nil)
}

// GetUser implements Forum.
func (c *Client) GetUser(id int) (*client.UserInfo, error) {
	return getJSON[client.UserInfo](c, "/api/v1/users/"+strconv.Itoa(id), nil)
}

// Categories implements Forum.
func (c *Client) Categories() ([]client.Category, error) {
	var cats []client.Category
	if err := c.doJSON("/api/v1/categories", nil, &cats); err != nil {
		return nil, err
	}

	return cats, nil
}

// LatestPosts implements Forum.
func (c *Client) LatestPosts(page int) (*client.PostList, error) {
	return getJSON[client.PostList](c, "/api/v1/posts", postsQuery("latest", "", page))
}

// CategoryPosts implements Forum.
func (c *Client) CategoryPosts(slug string, page int) (*client.PostList, error) {
	return getJSON[client.PostList](c, "/api/v1/posts", postsQuery("category", slug, page))
}

// GetPost implements Forum.
func (c *Client) GetPost(postID, page int) (*client.PostDetail, error) {
	path := "/api/v1/posts/" + strconv.Itoa(postID)

	return getJSON[client.PostDetail](c, path, pageQuery(page))
}

// GetPostAll implements Forum by looping GET /api/v1/posts/{id}?page=.
func (c *Client) GetPostAll(postID int) (*client.PostDetail, error) {
	var out *client.PostDetail

	for page := 1; page <= client.MaxPostPages; page++ {
		detail, err := c.GetPost(postID, page)
		if err != nil {
			return nil, err
		}

		next, stop := client.MergePostPage(out, detail)
		out = next

		if stop {
			return out, nil
		}
	}

	return out, nil
}

// Search implements Forum.
func (c *Client) Search(query string, page int) (*client.SearchResult, error) {
	values := url.Values{}
	values.Set("q", query)

	if page > 0 {
		values.Set("page", strconv.Itoa(page))
	}

	return getJSON[client.SearchResult](c, "/api/v1/search", values)
}

// Notifications implements Forum.
func (c *Client) Notifications() ([]client.Notification, error) {
	var notes []client.Notification
	if err := c.doJSON("/api/v1/notifications", nil, &notes); err != nil {
		return nil, err
	}

	if notes == nil {
		notes = []client.Notification{}
	}

	return notes, nil
}

func getJSON[T any](c *Client, path string, query url.Values) (*T, error) {
	var out T
	if err := c.doJSON(path, query, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func postsQuery(filter, slug string, page int) url.Values {
	query := url.Values{}
	query.Set("filter", filter)

	if slug != "" {
		query.Set("slug", slug)
	}

	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}

	return query
}

func pageQuery(page int) url.Values {
	if page <= 0 {
		return nil
	}

	query := url.Values{}
	query.Set("page", strconv.Itoa(page))

	return query
}
