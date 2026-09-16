package client

import "errors"

// Forum and Account errors. User-facing text is Chinese.
var (
	ErrCloudflare    = errors.New("cloudflare 拦截")
	ErrRateLimited   = errors.New("rate limited")
	ErrExpiredCookie = errors.New("pjwt 无效或过期")
	ErrNotFound      = errors.New("not found")
	ErrAuthRequired  = errors.New("需要登录")
	ErrForbidden     = errors.New("forbidden")
)

var (
	errCookieMissing    = errors.New("找不到 cookie 文件")
	errInvalidCookie    = errors.New("无效 cookie 文件")
	errNoPJWT           = errors.New("cookie 中没有 pjwt")
	errNilResponse      = errors.New("空响应")
	errUnexpectedStatus = errors.New("unexpected HTTP status")
)

const cookieHint = "合法内容：Cookie-Editor JSON 数组，或单行 pjwt=..."
