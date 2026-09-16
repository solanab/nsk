package client

import (
	"context"
	"fmt"
	"io"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

var newRequest = http.NewRequestWithContext //nolint:gochecknoglobals // request constructor test seam

type chromeStack struct {
	transport *chromeTransport
	jar       tls_client.CookieJar
}

type chromeTransport struct {
	inner doer
}

func (transport *chromeTransport) Do(req *http.Request) (*http.Response, error) {
	resp, err := transport.inner.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求: %w", err)
	}

	return resp, nil
}

func chromeHeaders() http.Header {
	return http.Header{
		"sec-ch-ua":                 {secChUA},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {secChUAPlatform},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {chromeUA},
		"accept":                    {acceptHTML},
		"accept-language":           {"zh-CN,zh;q=0.9"},
	}
}

func (c *Client) fetchHome() ([]byte, error) {
	status, headers, body, err := c.doHome()
	if err != nil {
		return nil, err
	}

	if challengePage(headers, body) {
		return nil, ErrCloudflare
	}

	if err := mapStatus(status); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) doHome() (int, http.Header, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := loadSeam(&newRequest)(ctx, http.MethodGet, siteHome, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("构造请求: %w", err)
	}

	req.Header = c.headers.Clone()

	resp, err := c.doer.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("请求首页: %w", err)
	}

	if resp == nil {
		return 0, nil, nil, errNilResponse
	}

	body, err := readClose(resp.Body)
	if err != nil {
		return 0, nil, nil, err
	}

	return resp.StatusCode, resp.Header, body, nil
}

func readClose(body io.ReadCloser) ([]byte, error) {
	data, readErr := io.ReadAll(body)
	closeErr := body.Close()

	if readErr != nil {
		return nil, fmt.Errorf("读取响应: %w", readErr)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("关闭响应: %w", closeErr)
	}

	return data, nil
}

func mapStatus(status int) error {
	switch status {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusForbidden:
		return ErrForbidden
	default:
		return fmt.Errorf("%w: %d", errUnexpectedStatus, status)
	}
}

func challengePage(headers http.Header, body []byte) bool {
	if headerNonEmpty(headers, "cf-mitigated") {
		return true
	}

	text := string(body)
	if strings.Contains(text, "challenge-platform") {
		return true
	}

	return strings.Contains(text, "Just a moment")
}

func headerNonEmpty(headers http.Header, key string) bool {
	if headers == nil {
		return false
	}

	if strings.TrimSpace(headers.Get(key)) != "" {
		return true
	}

	for name, values := range headers {
		if !strings.EqualFold(name, key) {
			continue
		}

		for _, value := range values {
			if strings.TrimSpace(value) != "" {
				return true
			}
		}
	}

	return false
}
