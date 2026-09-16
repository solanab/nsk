package client

// Forum is the NodeSeek operation surface. After ticket #2 it only has WhoAmI.
type Forum interface {
	WhoAmI() (*UserInfo, error)
}

var _ Forum = (*Client)(nil)
