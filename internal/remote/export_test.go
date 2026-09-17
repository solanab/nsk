package remote

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPTimeout is the remote Forum HTTP timeout.
func HTTPTimeout(c *Client) time.Duration {
	return c.http.Timeout
}

// StubNewRequest replaces http.NewRequest.
func StubNewRequest(fn func(string, string, io.Reader) (*http.Request, error)) func() {
	orig := newRequest
	newRequest = fn

	return func() { newRequest = orig }
}

// SetTransport replaces the HTTP transport.
func SetTransport(c *Client, transport http.RoundTripper) {
	c.http.Transport = transport
}

// PageQuery is the page query-string helper.
func PageQuery(page int) url.Values {
	return pageQuery(page)
}
