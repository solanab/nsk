package client

// Forum is the NodeSeek operation surface. After ticket #6 it adds Search.
type Forum interface {
	WhoAmI() (*UserInfo, error)
	Categories() ([]Category, error)
	LatestPosts(page int) (*PostList, error)
	CategoryPosts(slug string, page int) (*PostList, error)
	GetPost(postID, page int) (*PostDetail, error)
	GetPostAll(postID int) (*PostDetail, error)
	FormatPost(detail *PostDetail) string
	Search(query string, page int) (*SearchResult, error)
}

var _ Forum = (*Client)(nil)
