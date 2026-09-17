package app

import (
	"errors"
	"fmt"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
	"github.com/solanab/nsk/internal/remote"
)

var (
	errCookieOnClient = errors.New("client 机器不能导出 cookie")
	errServerOnClient = errors.New("client 机器不能跑 nsk server")
)

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
	Notifications() ([]client.Notification, error)
}

type remoteAccount struct {
	*remote.Client
}

func (*remoteAccount) ExportCookiesJSON() (string, error) {
	return "", errCookieOnClient
}

func (*remoteAccount) ExportPJWT() (string, error) {
	return "", errCookieOnClient
}

var (
	openAccount = openLocalClient //nolint:gochecknoglobals // Forum opener test seam
	openRemote  = pingRemote      //nolint:gochecknoglobals // remote opener test seam
	openForum   = openLocalForum  //nolint:gochecknoglobals // server Forum opener test seam
)

func openLocalClient(cookieFile string) (account, error) { //nolint:ireturn // CLI Account seam
	return wrapAccount(client.New(cookieFile))
}

func openLocalForum(cookieFile string) (client.Forum, error) { //nolint:ireturn // server Forum seam
	return wrapForum(client.New(cookieFile))
}

func wrapForum(forum *client.Client, err error) (client.Forum, error) { //nolint:ireturn // server Forum seam
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return forum, nil
}

func pingRemote(baseURL, token string) (account, error) { //nolint:ireturn // CLI Account seam
	rem := remote.New(baseURL, token)
	if err := rem.Ping(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	acc := new(remoteAccount)
	acc.Client = rem

	return acc, nil
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
		return loadSeam(&openRemote)(cfg.Client.URL, cfg.Client.Token)
	}

	acc, err := loadSeam(&openAccount)(cfg.CookieFile())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return acc, nil
}
