package client

// Forum is the NodeSeek operation surface. After ticket #4 it has WhoAmI, Categories, and list.
type Forum interface {
	WhoAmI() (*UserInfo, error)
	Categories() ([]Category, error)
	LatestPosts(page int) (*PostList, error)
	CategoryPosts(slug string, page int) (*PostList, error)
}

var _ Forum = (*Client)(nil)
