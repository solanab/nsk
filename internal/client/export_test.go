package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/bandwidth"
	"golang.org/x/net/proxy"
)

// Doer is the HTTP test seam.
type Doer = doer

// DoerFunc adapts a function to Doer.
type DoerFunc func(req *http.Request) (*http.Response, error)

// Do runs the function.
func (fn DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type jarClient struct {
	inner Doer
	jar   tls_client.CookieJar
}

func (client *jarClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL != nil {
		for _, cookie := range client.jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	resp, err := client.inner.Do(req)
	if err != nil {
		return nil, fmt.Errorf("测试传输: %w", err)
	}

	if resp != nil && req.URL != nil {
		client.jar.SetCookies(req.URL, resp.Cookies())
	}

	return resp, nil
}

// NewWithDoer builds a Client that uses d for HTTP.
func NewWithDoer(cookieFile string, transport Doer) (*Client, error) {
	cookies, path, err := readCookieFile(cookieFile)
	if err != nil {
		return nil, err
	}

	jar := tls_client.NewCookieJar()

	return finishNew(path, cookies, &jarClient{inner: transport, jar: jar}, jar)
}

// Unstarted is a Client that never warmed up.
func Unstarted() *Client {
	forum := new(Client)
	forum.jar = tls_client.NewCookieJar()
	forum.headers = chromeHeaders()

	return forum
}

// NilJar is a Client with no cookie jar.
func NilJar() *Client {
	return new(Client)
}

// NewChromeClient constructs the production TLS stack without warmup.
func NewChromeClient() error {
	_, err := newChromeClient()

	return err
}

// ChromeTransport wraps a Doer the same way production TLS does.
func ChromeTransport(inner Doer) *chromeTransport {
	return &chromeTransport{inner: inner}
}

// ApplyCookies is the jar import test seam.
func ApplyCookies(forum *Client, cookies []*http.Cookie) {
	applyCookies(forum.jar, cookies)
}

// EditorDomainPath projects Domain/Path after Cookie-Editor conversion.
func EditorDomainPath(cookie *http.Cookie) (string, string) {
	var cookies []*http.Cookie
	if cookie != nil {
		cookies = []*http.Cookie{cookie}
	}

	items := toCookieEditorItems(cookies, cookieHost)
	if len(items) == 0 {
		return "", ""
	}

	return items[0].Domain, items[0].Path
}

// HTTPSURL is the URL helper test seam.
func HTTPSURL(host string) *url.URL {
	return httpsURL(host)
}

// SetCookieURL is the jar SetCookies URL test seam.
func SetCookieURL(cookie *http.Cookie) *url.URL {
	return setCookieURL(cookie)
}

// StubConstruct replaces client.New.
func StubConstruct(fn func(string) (*Client, error)) func() {
	return swapSeam(&construct, fn)
}

// StubNewHTTPClient replaces tls-client construction.
func StubNewHTTPClient(
	fn func(tls_client.Logger, ...tls_client.HttpClientOption) (tls_client.HttpClient, error),
) func() {
	return swapSeam(&newHTTPClient, fn)
}

// StubReadFile replaces cookie file reads.
func StubReadFile(fn func(string) ([]byte, error)) func() {
	return swapSeam(&readFile, fn)
}

// StubWriteFile replaces cookie file writes.
func StubWriteFile(fn func(string, []byte, os.FileMode) error) func() {
	return swapSeam(&writeFile, fn)
}

// StubChmodFile replaces cookie chmod.
func StubChmodFile(fn func(string, os.FileMode) error) func() {
	return swapSeam(&chmodFile, fn)
}

// StubStatFile replaces cookie stat.
func StubStatFile(fn func(string) (os.FileInfo, error)) func() {
	return swapSeam(&statFile, fn)
}

// StubAbsPath replaces filepath.Abs.
func StubAbsPath(fn func(string) (string, error)) func() {
	return swapSeam(&absPath, fn)
}

// StubNewRequest replaces HTTP request construction.
func StubNewRequest(fn func(context.Context, string, string, io.Reader) (*http.Request, error)) func() {
	return swapSeam(&newRequest, fn)
}

// ParsePostList is the list HTML parse test seam.
func ParsePostList(html []byte, page int) (*PostList, error) {
	return parsePostList(html, page)
}

// ParsePostDetail is the post HTML parse test seam.
func ParsePostDetail(html []byte, postID, page int) (*PostDetail, error) {
	return parsePostDetail(html, postID, page)
}

// StubParseHTML replaces goquery document parsing.
func StubParseHTML(fn func(io.Reader) (*goquery.Document, error)) func() {
	return swapSeam(&parseHTML, fn)
}

// StubMarshalIndent replaces cookie JSON encoding.
func StubMarshalIndent(fn func(any, string, string) ([]byte, error)) func() {
	return swapSeam(&marshalIndent, fn)
}

type fakeTLS struct {
	doer Doer
}

// FakeTLS is a tls-client.HttpClient that delegates Do to inner.
func FakeTLS(inner Doer) tls_client.HttpClient { //nolint:ireturn // test fake of an upstream interface
	return &fakeTLS{doer: inner}
}

func (fake *fakeTLS) Do(req *http.Request) (*http.Response, error) {
	resp, err := fake.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fake tls: %w", err)
	}

	return resp, nil
}

func (*fakeTLS) GetCookies(*url.URL) []*http.Cookie { return nil }

func (*fakeTLS) SetCookies(*url.URL, []*http.Cookie) {}

func (*fakeTLS) SetCookieJar(http.CookieJar) {}

func (*fakeTLS) GetCookieJar() http.CookieJar { return nil } //nolint:ireturn // unused fake method

func (*fakeTLS) SetProxy(string) error { return nil }

func (*fakeTLS) GetProxy() string { return "" }

func (*fakeTLS) SetFollowRedirect(bool) {}

func (*fakeTLS) GetFollowRedirect() bool { return false }

func (*fakeTLS) CloseIdleConnections() {}

func (*fakeTLS) Get(string) (*http.Response, error) { return nil, errNilResponse }

func (*fakeTLS) Head(string) (*http.Response, error) { return nil, errNilResponse }

func (*fakeTLS) Post(string, string, io.Reader) (*http.Response, error) {
	return nil, errNilResponse
}

func (*fakeTLS) GetBandwidthTracker() bandwidth.BandwidthTracker { //nolint:ireturn // unused fake method
	return nil
}

func (*fakeTLS) GetDialer() proxy.ContextDialer { //nolint:ireturn // unused fake method
	return nil
}

func (*fakeTLS) GetTLSDialer() tls_client.TLSDialerFunc { return nil }

func (*fakeTLS) AddPreRequestHook(tls_client.PreRequestHookFunc) {}

func (*fakeTLS) AddPostResponseHook(tls_client.PostResponseHookFunc) {}

func (*fakeTLS) ResetPreHooks() {}

func (*fakeTLS) ResetPostHooks() {}
