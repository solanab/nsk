// Package remote implements Forum over nsk server HTTP.
package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/solanab/nsk/internal/client"
)

const remoteTimeout = 60 * time.Second

var (
	errNotReady = errors.New("server 未就绪")
	errRemote   = errors.New("server")
	newRequest  = http.NewRequest //nolint:gochecknoglobals // request constructor test seam
)

// Client is the Forum implementation that talks to nsk server.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

var _ client.Forum = (*Client)(nil)

// New constructs a remote Forum client. Timeout is 60s.
func New(baseURL, token string) *Client {
	httpClient := new(http.Client)
	httpClient.Timeout = remoteTimeout

	remote := new(Client)
	remote.baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	remote.token = token
	remote.http = httpClient

	return remote
}

// Ping checks GET /api/v1/health for {ok:true,ready:true}.
func (c *Client) Ping() error {
	var health struct {
		OK    bool `json:"ok"`
		Ready bool `json:"ready"`
	}
	if err := c.doJSON("/api/v1/health", nil, &health); err != nil {
		return err
	}

	if !health.OK || !health.Ready {
		return errNotReady
	}

	return nil
}

// FormatPost renders Markdown locally. It does not touch the network.
func (*Client) FormatPost(detail *client.PostDetail) string {
	return client.FormatPost(detail)
}

func (c *Client) doJSON(path string, query url.Values, out any) error {
	parsed, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("无效 URL: %w", err)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	if query != nil {
		parsed.RawQuery = query.Encode()
	}

	req, err := newRequest(http.MethodGet, parsed.String(), nil)
	if err != nil {
		return fmt.Errorf("构造请求: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("请求: %w", err)
	}

	return decodeResponse(resp, out)
}

func decodeResponse(resp *http.Response, out any) error {
	body, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()

	if err != nil {
		return fmt.Errorf("读取响应: %w", err)
	}

	if closeErr != nil {
		return fmt.Errorf("关闭响应: %w", closeErr)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return decodeAPIError(resp.StatusCode, body)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("解析 JSON: %w", err)
	}

	return nil
}

func decodeAPIError(status int, body []byte) error {
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error != "" {
		return fmt.Errorf("%w: %s", errRemote, payload.Error)
	}

	return fmt.Errorf("%w: HTTP %d", errRemote, status)
}
