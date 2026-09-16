package client

import (
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
)

func (item cookieItem) toCookie() *http.Cookie {
	path := item.Path
	if path == "" {
		path = "/"
	}

	domain := item.Domain
	if domain == "" {
		domain = cookieDomain
	}

	var expires time.Time
	if item.ExpirationDate != nil && *item.ExpirationDate > 0 {
		expires = time.Unix(int64(*item.ExpirationDate), 0)
	}

	cookie := new(http.Cookie)
	cookie.Name = item.Name
	cookie.Value = item.Value
	cookie.Path = path
	cookie.Domain = domain
	cookie.Expires = expires
	cookie.Secure = item.Secure
	cookie.HttpOnly = item.HTTPOnly
	cookie.SameSite = cookieSameSite(item.SameSite)

	return cookie
}

func cookieSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "no_restriction", "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

func toCookieEditorItems(cookies []*http.Cookie, defaultDomain string) []cookieItem {
	items := make([]cookieItem, 0, len(cookies))
	for _, cookie := range cookies {
		domain := cookie.Domain
		if domain == "" {
			domain = defaultDomain
		}

		path := cookie.Path
		if path == "" {
			path = "/"
		}

		isSession := cookie.Expires.IsZero()

		var expDate *float64

		if !isSession {
			sec := float64(cookie.Expires.Unix())
			expDate = &sec
		}

		items = append(items, cookieItem{
			Domain:         domain,
			ExpirationDate: expDate,
			HostOnly:       !strings.HasPrefix(domain, "."),
			HTTPOnly:       cookie.HttpOnly,
			Name:           cookie.Name,
			Path:           path,
			SameSite:       sameSiteName(cookie.SameSite),
			Secure:         cookie.Secure,
			Session:        isSession,
			StoreID:        nil,
			Value:          cookie.Value,
		})
	}

	return items
}

func sameSiteName(mode http.SameSite) string {
	switch mode {
	case http.SameSiteDefaultMode:
		return "unspecified"
	case http.SameSiteLaxMode:
		return "lax"
	case http.SameSiteStrictMode:
		return "strict"
	case http.SameSiteNoneMode:
		return "no_restriction"
	default:
		return "unspecified"
	}
}
