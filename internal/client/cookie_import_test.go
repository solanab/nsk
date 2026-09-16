package client_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

const editorJSON = `[
  {
    "domain": ".nodeseek.com",
    "expirationDate": 2000000000,
    "hostOnly": false,
    "httpOnly": true,
    "name": "pjwt",
    "path": "/",
    "sameSite": "lax",
    "secure": true,
    "session": false,
    "value": "token-1"
  },
  {
    "domain": ".nodeseek.com",
    "name": "cf_clearance",
    "path": "/",
    "value": "should-drop"
  },
  {
    "domain": ".nodeseek.com",
    "name": "cf_bm",
    "path": "/",
    "value": "should-drop-too"
  },
  {
    "domain": ".nodeseek.com",
    "name": "theme",
    "path": "/",
    "sameSite": "strict",
    "value": "dark"
  },
  {
    "domain": ".nodeseek.com",
    "name": "none_ck",
    "path": "/",
    "sameSite": "none",
    "value": "x"
  },
  {
    "domain": ".nodeseek.com",
    "name": "nr",
    "path": "/",
    "sameSite": "no_restriction",
    "value": "y"
  },
  {
    "name": "bare",
    "value": "z"
  },
  {
    "name": "",
    "value": "skip-me"
  }
]`

func TestImportPJWTLine(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), userDoer(t))

	line, err := forum.ExportPJWT()
	if err != nil {
		t.Fatal(err)
	}

	if line != "pjwt="+pjwtToken {
		t.Fatalf("ExportPJWT = %q", line)
	}
}

func TestImportCookieEditorStripsCF(t *testing.T) {
	t.Parallel()

	forum := openOK(t, writeCookie(t, editorJSON), userDoer(t))

	raw, err := forum.ExportCookiesJSON()
	if err != nil {
		t.Fatal(err)
	}

	got := cookieNames(parseEditor(t, raw))
	if got["pjwt"] != "token-1" {
		t.Fatalf("pjwt = %q", got["pjwt"])
	}

	if got["theme"] != "dark" {
		t.Fatalf("theme = %q", got["theme"])
	}

	if got["cf_clearance"] == "should-drop" {
		t.Fatal("imported cf_clearance was not stripped")
	}

	if _, ok := got["cf_bm"]; ok {
		t.Fatal("imported cf_bm was not stripped")
	}

	if got["bare"] != "z" {
		t.Fatalf("bare = %q", got["bare"])
	}
}

func TestImportInvalidCookieFile(t *testing.T) {
	t.Parallel()

	err := openErr(t, writeCookie(t, "not-a-cookie"), userDoer(t))
	if !strings.Contains(err.Error(), "无效 cookie 文件") {
		t.Fatalf("err = %v", err)
	}
}

func TestImportEmptyCookieFile(t *testing.T) {
	t.Parallel()

	err := openErr(t, writeCookie(t, "   \n"), userDoer(t))
	if !strings.Contains(err.Error(), "无效 cookie 文件") {
		t.Fatalf("err = %v", err)
	}
}

func TestImportPJWTMultilineRejected(t *testing.T) {
	t.Parallel()

	err := openErr(t, writeCookie(t, "pjwt=one\ntwo"), userDoer(t))
	if !strings.Contains(err.Error(), "无效 cookie 文件") {
		t.Fatalf("err = %v", err)
	}
}

func TestMissingCookieFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.json")

	_, err := client.New(path)
	if err == nil {
		t.Fatal("expected missing cookie error")
	}

	if !strings.Contains(err.Error(), path) {
		t.Fatalf("path missing in %v", err)
	}

	if !strings.Contains(err.Error(), "Cookie-Editor JSON") || !strings.Contains(err.Error(), "pjwt=") {
		t.Fatalf("formats missing in %v", err)
	}
}

func TestWarmupWritesJSONNotPJWTLine(t *testing.T) {
	t.Parallel()

	path := pjwtFile(t)
	_ = openOK(t, path, userDoer(t))

	data, err := os.ReadFile(path) //nolint:gosec // test cookie file
	if err != nil {
		t.Fatal(err)
	}

	if !json.Valid(data) {
		t.Fatalf("cookie file is not JSON: %s", data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}
}

func TestWarmupKeepsCFFromResponse(t *testing.T) {
	t.Parallel()

	extra := http.Header{}
	extra.Add("Set-Cookie", "cf_clearance=from-warmup; Path=/; Domain=.nodeseek.com; Secure")

	forum := openOK(t, pjwtFile(t), fixtureDoer(t, htmlUser, http.StatusOK, extra))

	raw, err := forum.ExportCookiesJSON()
	if err != nil {
		t.Fatal(err)
	}

	got := cookieNames(parseEditor(t, raw))
	if got["cf_clearance"] != "from-warmup" {
		t.Fatalf("cf_clearance = %q", got["cf_clearance"])
	}

	if got["pjwt"] != pjwtToken {
		t.Fatalf("pjwt = %q", got["pjwt"])
	}
}
