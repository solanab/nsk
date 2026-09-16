package client

// Forum is the NodeSeek operation surface. After ticket #5 it has WhoAmI, Categories, list, and post.
type Forum interface {
	WhoAmI() (*UserInfo, error)
	Categories() ([]Category, error)
	LatestPosts(page int) (*PostList, error)
	CategoryPosts(slug string, page int) (*PostList, error)
	GetPost(postID, page int) (*PostDetail, error)
	GetPostAll(postID int) (*PostDetail, error)
	FormatPost(detail *PostDetail) string
}

var _ Forum = (*Client)(nil)
