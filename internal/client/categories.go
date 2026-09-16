package client

// Site is the NodeSeek origin without a trailing slash.
const Site = "https://www.nodeseek.com"

const (
	// ListPerPage is the pinned list-page size from page-1 HTML fixtures.
	ListPerPage = 49
	// FloorsPerPage is the pinned post-page size from post-703863-1.html (11 content-item).
	FloorsPerPage = 11
	// MaxPostPages is the GetPostAll hard cap.
	MaxPostPages = 50
)

const (
	// PathHome is the latest-list URL template.
	PathHome = "/page-{n}"
	// PathCategory is the category-list URL template.
	PathCategory = "/categories/{slug}?page={n}"
	// PathPost is the post-page URL template.
	PathPost = "/post-{id}-{page}"
	// PathFloor is the floor-anchor URL template.
	PathFloor = "/post-{id}-{page}#{floor}"
	// PathUser is the user-space URL template.
	PathUser = "/space/{id}"
)

func categoryTable() []Category {
	return []Category{
		{Slug: "daily", Name: "日常"},
		{Slug: "tech", Name: "技术"},
		{Slug: "info", Name: "情报"},
		{Slug: "review", Name: "测评"},
		{Slug: "trade", Name: "交易"},
		{Slug: "carpool", Name: "拼车"},
		{Slug: "promotion", Name: "推广"},
		{Slug: "life", Name: "生活"},
		{Slug: "dev", Name: "Dev"},
		{Slug: "photo-share", Name: "贴图"},
		{Slug: "expose", Name: "曝光"},
		{Slug: "inside", Name: "内版"},
		{Slug: "meaningless", Name: "无意义"},
		{Slug: "sandbox", Name: "沙盒"},
	}
}

// Categories returns the hardcoded board table. Local Forum never errors.
func Categories() ([]Category, error) {
	return categoryTable(), nil
}

// Categories returns the hardcoded board table. It does not touch the network.
func (*Client) Categories() ([]Category, error) {
	return Categories()
}
