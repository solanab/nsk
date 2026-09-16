package client

// UserInfo is the WhoAmI projection of window.__config__.user.
type UserInfo struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Chicken int    `json:"chicken,omitempty"`
	Level   string `json:"level,omitempty"`
}

// Category is a forum board.
type Category struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// PostSummary is one list-page row.
type PostSummary struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category,omitempty"`
	Author    string `json:"author"`
	Replies   int    `json:"replies"`
	CreatedAt string `json:"created_at,omitempty"` //nolint:tagliatelle // design.md JSON contract
	URL       string `json:"url"`
}

// PostList is one latest or category page.
//
// JSON tags are snake_case per docs/design.md PostList.
type PostList struct {
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"` //nolint:tagliatelle // design.md JSON contract
	Posts   []PostSummary `json:"posts"`
}

// Floor is one post reply. Number is the site's 0-based floor (#0 is OP).
type Floor struct {
	Number    int    `json:"floor"`
	Author    string `json:"author"`
	CreatedAt string `json:"created_at,omitempty"` //nolint:tagliatelle // design.md JSON contract
	Markdown  string `json:"markdown"`
	ReplyTo   *int   `json:"reply_to,omitempty"` //nolint:tagliatelle // design.md JSON contract
}

// PostDetail is one post page (or GetPostAll concatenation).
type PostDetail struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Category string  `json:"category,omitempty"`
	Author   string  `json:"author"`
	Page     int     `json:"page"`
	Pages    int     `json:"pages"`
	URL      string  `json:"url"`
	Floors   []Floor `json:"floors"`
}

// SavedView is JSON stdout after `nsk post -o` writes Markdown.
type SavedView struct {
	Saved string `json:"saved"`
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// SearchResult is one logged-in search page. It has no per_page field.
type SearchResult struct {
	Query string        `json:"query"`
	Page  int           `json:"page"`
	Posts []PostSummary `json:"posts"`
}

// SiteStructure is the Agent entry: empty argv or `nsk structure` stdout.
type SiteStructure struct {
	Site       string        `json:"site"`
	Categories []Category    `json:"categories"`
	Pagination Pagination    `json:"pagination"`
	Paths      SitePaths     `json:"paths"`
	Commands   []CommandSpec `json:"commands"`
}

// Pagination is the list/post page-size hypothesis until fixtures pin it.
//
// JSON tags are snake_case per docs/design.md SiteStructure.
type Pagination struct {
	ListPerPage   int `json:"list_per_page"`   //nolint:tagliatelle // design.md JSON contract
	FloorsPerPage int `json:"floors_per_page"` //nolint:tagliatelle // design.md JSON contract
	MaxPostPages  int `json:"max_post_pages"`  //nolint:tagliatelle // design.md JSON contract
}

// SitePaths holds URL templates for the forum.
type SitePaths struct {
	Home     string `json:"home"`
	Category string `json:"category"`
	Post     string `json:"post"`
	Floor    string `json:"floor"`
	User     string `json:"user"`
}

// CommandSpec describes one implemented CLI subcommand.
type CommandSpec struct {
	Name   string `json:"name"`
	Usage  string `json:"usage"`
	Stdout string `json:"stdout"`
}

type cookieItem struct {
	Domain         string   `json:"domain"`
	ExpirationDate *float64 `json:"expirationDate,omitempty"`
	HostOnly       bool     `json:"hostOnly"`
	HTTPOnly       bool     `json:"httpOnly"`
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	SameSite       string   `json:"sameSite"`
	Secure         bool     `json:"secure"`
	Session        bool     `json:"session"`
	StoreID        *string  `json:"storeId"`
	Value          string   `json:"value"`
}

type configJSON struct {
	User *userJSON `json:"user"`
}

type userJSON struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Chicken int    `json:"chicken"`
	Level   string `json:"level"`
}
