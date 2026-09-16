package app

import (
	"errors"
	"fmt"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
)

var errServerLater = errors.New("nsk server 在后续票才可用")

type account interface {
	WhoAmI() (*client.UserInfo, error)
	GetUser(id int) (*client.UserInfo, error)
	ExportCookiesJSON() (string, error)
	ExportPJWT() (string, error)
	LatestPosts(page int) (*client.PostList, error)
	CategoryPosts(slug string, page int) (*client.PostList, error)
	GetPost(postID, page int) (*client.PostDetail, error)
	GetPostAll(postID int) (*client.PostDetail, error)
	Search(query string, page int) (*client.SearchResult, error)
}

var openAccount = openLocalClient //nolint:gochecknoglobals // Forum opener test seam

func openLocalClient(cookieFile string) (account, error) { //nolint:ireturn // CLI Account seam
	return wrapAccount(client.New(cookieFile))
}

func wrapAccount(forum *client.Client, err error) (account, error) { //nolint:ireturn // CLI Account seam
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return forum, nil
}

func accountFrom(root *cliRoot) (account, error) { //nolint:ireturn // CLI Account seam
	cfg, err := config.LoadPath(root.Config)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if cfg.HasClient() {
		return nil, errServerLater
	}

	acc, err := loadSeam(&openAccount)(cfg.CookieFile())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return acc, nil
}
