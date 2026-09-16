package client

// Forum is the NodeSeek operation surface. After ticket #3 it has WhoAmI and Categories.
type Forum interface {
	WhoAmI() (*UserInfo, error)
	Categories() ([]Category, error)
}

var _ Forum = (*Client)(nil)
