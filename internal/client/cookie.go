package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

var (
	readFile      = os.ReadFile        //nolint:gochecknoglobals // cookie file test seam
	writeFile     = os.WriteFile       //nolint:gochecknoglobals // cookie file test seam
	chmodFile     = os.Chmod           //nolint:gochecknoglobals // cookie mode test seam
	statFile      = os.Stat            //nolint:gochecknoglobals // cookie stat test seam
	absPath       = filepath.Abs       //nolint:gochecknoglobals // cookie path test seam
	marshalIndent = json.MarshalIndent //nolint:gochecknoglobals // cookie JSON test seam
)

func forumURL() *url.URL {
	return httpsURL(cookieHost)
}

func httpsURL(host string) *url.URL {
	parsed, err := url.Parse("https://" + host + "/")
	if err != nil {
		parsed = new(url.URL)
		parsed.Scheme = "https"
		parsed.Host = host
		parsed.Path = "/"
	}

	return parsed
}

func readCookieFile(path string) ([]*http.Cookie, string, error) {
	abs, err := loadSeam(&absPath)(path)
	if err != nil {
		return nil, "", fmt.Errorf("cookie 路径: %w", err)
	}

	if _, err := loadSeam(&statFile)(abs); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, "", missingCookie(abs)
		}

		return nil, "", fmt.Errorf("读取 cookie 文件: %w", err)
	}

	data, err := loadSeam(&readFile)(abs)
	if err != nil {
		return nil, "", fmt.Errorf("读取 cookie 文件: %w", err)
	}

	cookies, err := parseCookieBytes(data)
	if err != nil {
		return nil, "", err
	}

	return cookies, abs, nil
}

func missingCookie(path string) error {
	return fmt.Errorf("%w: %s\n%s", errCookieMissing, path, cookieHint)
}

func parseCookieBytes(data []byte) ([]*http.Cookie, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, errInvalidCookie
	}

	if bytes.HasPrefix(trimmed, []byte("pjwt=")) {
		return parsePJWTLine(trimmed)
	}

	return parseCookieEditor(trimmed)
}

func parsePJWTLine(trimmed []byte) ([]*http.Cookie, error) {
	value := strings.TrimPrefix(string(trimmed), "pjwt=")
	if strings.ContainsAny(value, "\r\n") {
		return nil, errInvalidCookie
	}

	cookie := new(http.Cookie)
	cookie.Name = "pjwt"
	cookie.Value = value
	cookie.Path = "/"
	cookie.Domain = cookieDomain

	return []*http.Cookie{cookie}, nil
}

func parseCookieEditor(data []byte) ([]*http.Cookie, error) {
	var items []cookieItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidCookie, err)
	}

	cookies := make([]*http.Cookie, 0, len(items))
	for _, item := range items {
		if item.Name == "" || cloudflareCookie(item.Name) {
			continue
		}

		cookies = append(cookies, item.toCookie())
	}

	return cookies, nil
}

func cloudflareCookie(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), "cf_")
}

func applyCookies(jar http.CookieJar, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		jar.SetCookies(setCookieURL(cookie), []*http.Cookie{cookie})
	}
}

func setCookieURL(cookie *http.Cookie) *url.URL {
	switch cookie.Domain {
	case "", cookieDomain, cookieHost:
		return forumURL()
	default:
		host := strings.TrimPrefix(cookie.Domain, ".")
		if host == "" {
			return forumURL()
		}

		return httpsURL(host)
	}
}

func (c *Client) saveCookies() error {
	payload, err := c.ExportCookiesJSON()
	if err != nil {
		return err
	}

	if err := loadSeam(&writeFile)(c.cookieFile, []byte(payload+"\n"), cookieFileMode); err != nil {
		return fmt.Errorf("写入 cookie 文件: %w", err)
	}

	if err := loadSeam(&chmodFile)(c.cookieFile, cookieFileMode); err != nil {
		return fmt.Errorf("设置 cookie 文件权限: %w", err)
	}

	return nil
}

// ExportCookiesJSON dumps this process jar as Cookie-Editor JSON.
func (c *Client) ExportCookiesJSON() (string, error) {
	items := toCookieEditorItems(c.allCookies(), cookieHost)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Name == items[j].Name {
			return items[i].Domain < items[j].Domain
		}

		return items[i].Name < items[j].Name
	})

	data, err := loadSeam(&marshalIndent)(items, "", "  ")
	if err != nil {
		return "", fmt.Errorf("编码 cookie JSON: %w", err)
	}

	return string(data), nil
}

// ExportPJWT prints the pjwt cookie as a pjwt= line.
func (c *Client) ExportPJWT() (string, error) {
	for _, cookie := range c.allCookies() {
		if cookie.Name == "pjwt" {
			return "pjwt=" + cookie.Value, nil
		}
	}

	return "", errNoPJWT
}

func (c *Client) allCookies() []*http.Cookie {
	if c.jar == nil {
		return nil
	}

	var cookies []*http.Cookie
	for _, list := range c.jar.GetAllCookies() {
		cookies = append(cookies, list...)
	}

	if len(cookies) > 0 {
		return cookies
	}

	return c.jar.Cookies(forumURL())
}
