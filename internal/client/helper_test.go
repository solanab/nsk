package client_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

const (
	wantUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	wantSecChUA = `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`
	pjwtToken   = "test-pjwt-token" //nolint:gosec // fixture token, not a secret
	nameAlice   = "alice"
	nameBob     = "bob"
	levelTwo    = "Lv.2"
	htmlUser    = "homepage-user.html"
	cfChallenge = "challenge"
	forumHost   = "www.nodeseek.com"
	pjwtName    = "pjwt"
	slugDaily   = "daily"
	slugTech    = "tech"
)

type editorCookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("testdata", "html", name)

	data, err := os.ReadFile(path) //nolint:gosec // testdata HTML fixture
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func writeCookie(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "cookie.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func htmlResponse(status int, body []byte, extra http.Header) *http.Response {
	header := http.Header{}
	header.Set("Content-Type", "text/html")

	for key, values := range extra {
		header[key] = append([]string{}, values...)
	}

	resp := new(http.Response)
	resp.StatusCode = status
	resp.Header = header
	resp.Body = io.NopCloser(bytes.NewReader(body))
	resp.ContentLength = int64(len(body))

	return resp
}

func fixtureDoer(t *testing.T, name string, status int, extra http.Header) client.DoerFunc {
	t.Helper()
	body := fixture(t, name)

	return func(*http.Request) (*http.Response, error) {
		return htmlResponse(status, body, extra), nil
	}
}

func userDoer(t *testing.T) client.DoerFunc {
	t.Helper()

	return fixtureDoer(t, htmlUser, http.StatusOK, nil)
}

func openOK(t *testing.T, path string, transport client.Doer) *client.Client {
	t.Helper()

	forum, err := client.NewWithDoer(path, transport)
	if err != nil {
		t.Fatal(err)
	}

	return forum
}

func openErr(t *testing.T, path string, transport client.Doer) error {
	t.Helper()

	_, err := client.NewWithDoer(path, transport)
	if err == nil {
		t.Fatal("expected error")
	}

	return err //nolint:wrapcheck // test helper returns the Forum error unchanged
}

func parseEditor(t *testing.T, raw string) []editorCookie {
	t.Helper()

	var items []editorCookie
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		t.Fatal(err)
	}

	return items
}

func cookieNames(items []editorCookie) map[string]string {
	out := make(map[string]string, len(items))
	for _, item := range items {
		out[item.Name] = item.Value
	}

	return out
}

func headerVal(header http.Header, key string) string {
	if value := header.Get(key); value != "" {
		return value
	}

	for name, values := range header {
		if len(values) > 0 && strings.EqualFold(name, key) {
			return values[0]
		}
	}

	return ""
}

func pjwtFile(t *testing.T) string {
	t.Helper()

	return writeCookie(t, "pjwt="+pjwtToken+"\n")
}

func requestHasPJWT(req *http.Request) bool {
	for _, cookie := range req.Cookies() {
		if cookie.Name == pjwtName && cookie.Value == pjwtToken {
			return true
		}
	}

	return false
}
