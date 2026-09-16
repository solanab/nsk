// Package client is the local Forum implementation.
package client

import (
	"fmt"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const (
	siteHome       = "https://www.nodeseek.com/"
	cookieDomain   = ".nodeseek.com"
	cookieHost     = "www.nodeseek.com"
	timeoutSeconds = 30
	requestTimeout = timeoutSeconds * time.Second
	cookieFileMode = 0o600
	chromeUA       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	secChUA         = `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`
	secChUAPlatform = `"macOS"`
	acceptHTML      = "text/html,application/xhtml+xml,application/xml;q=0.9," +
		"image/avif,image/webp,image/apng,*/*;q=0.8"
)

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the local Forum implementation (Chrome 124 TLS + cookie jar).
type Client struct {
	doer       doer
	jar        tls_client.CookieJar
	headers    http.Header
	cookieFile string
	me         *UserInfo
}

var (
	newHTTPClient = tls_client.NewHttpClient //nolint:gochecknoglobals // TLS constructor test seam
	construct     = newProduction            //nolint:gochecknoglobals // New() test seam
)

// New loads CookieFile, warms up GET /, and writes the jar on success.
func New(cookieFile string) (*Client, error) {
	return loadSeam(&construct)(cookieFile)
}

func newProduction(cookieFile string) (*Client, error) {
	cookies, path, err := readCookieFile(cookieFile)
	if err != nil {
		return nil, err
	}

	stack, err := newChromeClient()
	if err != nil {
		return nil, err
	}

	return finishNew(path, cookies, stack.transport, stack.jar)
}

func newChromeClient() (*chromeStack, error) {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(timeoutSeconds),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithCookieJar(jar),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	httpClient, err := loadSeam(&newHTTPClient)(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("创建 TLS 客户端失败: %w", err)
	}

	stack := new(chromeStack)
	stack.transport = &chromeTransport{inner: httpClient}
	stack.jar = jar

	return stack, nil
}

func finishNew(path string, cookies []*http.Cookie, transport doer, jar tls_client.CookieJar) (*Client, error) {
	applyCookies(jar, cookies)

	forum := new(Client)
	forum.doer = transport
	forum.jar = jar
	forum.headers = chromeHeaders()
	forum.cookieFile = path

	if err := forum.warmup(); err != nil {
		return nil, err
	}

	if err := forum.saveCookies(); err != nil {
		return nil, err
	}

	return forum, nil
}

// WhoAmI returns the warmup projection of the current Account.
func (c *Client) WhoAmI() (*UserInfo, error) {
	if c.me == nil {
		return nil, ErrExpiredCookie
	}

	info := *c.me

	return &info, nil
}
