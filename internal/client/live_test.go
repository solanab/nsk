package client_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solanab/nsk/internal/client"
)

func TestLiveChrome124Spike(t *testing.T) {
	if os.Getenv("NSK_LIVE") != "1" {
		t.Skip("NSK_LIVE=1 才打真实站点")
	}

	path := os.Getenv("NSK_LIVE_COOKIE")
	if path == "" {
		t.Skip("未设置 NSK_LIVE_COOKIE，不编造 pjwt")
	}

	data, err := os.ReadFile(path) //nolint:gosec // operator-supplied cookie path
	if err != nil {
		t.Skip("NSK_LIVE_COOKIE 不可读，不编造 pjwt")
	}

	tmp := filepath.Join(t.TempDir(), "cookie.json")
	if err := os.WriteFile(tmp, data, 0o600); err != nil { //nolint:gosec // temp copy, fixed name
		t.Fatal(err)
	}

	forum, err := client.New(tmp)
	if err != nil {
		t.Fatal(err)
	}

	info, err := forum.WhoAmI()
	if err != nil {
		t.Fatal(err)
	}

	if info.ID == 0 || info.Name == "" {
		t.Fatalf("whoami: %+v", info)
	}
}
