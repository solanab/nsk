package client

// UserInfo is the WhoAmI projection of window.__config__.user.
type UserInfo struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Chicken int    `json:"chicken,omitempty"`
	Level   string `json:"level,omitempty"`
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
