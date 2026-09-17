package client_test

import (
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"

	"github.com/solanab/nsk/internal/client"
)

func TestNewSuccessUsesTLSStack(t *testing.T) {
	path := pjwtFile(t)
	t.Cleanup(client.StubNewHTTPClient(func(
		_ tls_client.Logger,
		_ ...tls_client.HttpClientOption,
	) (tls_client.HttpClient, error) {
		return client.FakeTLS(userDoer(t)), nil
	}))

	forum, err := client.New(path)
	if err != nil {
		t.Fatal(err)
	}

	info, err := forum.WhoAmI()
	if err != nil {
		t.Fatal(err)
	}

	if info.Name != nameAlice {
		t.Fatalf("%#v", info)
	}
}

func TestEmptyDomainAndPathExport(t *testing.T) {
	t.Parallel()

	empty := new(http.Cookie)
	empty.Name = "n"
	empty.Value = "v"

	domain, path := client.EditorDomainPath(empty)
	if domain != forumHost || path != "/" {
		t.Fatalf("domain=%q path=%q", domain, path)
	}

	none, nonePath := client.EditorDomainPath(nil)
	if none != "" || nonePath != "" {
		t.Fatalf("nil cookie %q %q", none, nonePath)
	}

	forum := client.Unstarted()
	first := new(http.Cookie)
	first.Name = "dup"
	first.Value = "a"
	first.Domain = "example.com"

	second := new(http.Cookie)
	second.Name = "dup"
	second.Value = "b"
	second.Domain = "other.org"

	www := new(http.Cookie)
	www.Name = "host"
	www.Value = "h"
	www.Domain = forumHost

	dot := new(http.Cookie)
	dot.Name = "dot"
	dot.Value = "c"
	dot.Domain = "."

	client.ApplyCookies(forum, []*http.Cookie{first, second, www, dot})

	raw, err := forum.ExportCookiesJSON()
	if err != nil {
		t.Fatal(err)
	}

	got := cookieNames(parseEditor(t, raw))
	if got["dup"] == "" || got["dot"] != "c" || got["host"] != "h" {
		t.Fatalf("%q", raw)
	}
}

func TestLowercaseCFHeader(t *testing.T) {
	t.Parallel()

	extra := http.Header{}
	extra[headerCF] = []string{cfChallenge}

	err := openErr(t, pjwtFile(t), fixtureDoer(t, htmlUser, http.StatusOK, extra))
	if err == nil || err.Error() != client.ErrCloudflare.Error() {
		t.Fatalf("err = %v", err)
	}
}

func TestPJWTSetCookiesUsesForumURL(t *testing.T) {
	t.Parallel()

	cookie := new(http.Cookie)
	cookie.Name = pjwtName
	cookie.Domain = ".nodeseek.com"

	got := client.SetCookieURL(cookie)
	if got == nil || got.String() != "https://"+forumHost+"/" {
		t.Fatalf("url = %v", got)
	}

	bare := new(http.Cookie)
	bare.Name = pjwtName

	got = client.SetCookieURL(bare)
	if got == nil || got.Host != forumHost {
		t.Fatalf("empty domain url = %v", got)
	}
}

func TestHTTPSURLParse(t *testing.T) {
	t.Parallel()

	parsed := client.HTTPSURL("www.nodeseek.com")
	if parsed == nil || parsed.Host != "www.nodeseek.com" {
		t.Fatalf("%#v", parsed)
	}

	fallback := client.HTTPSURL(":")
	if fallback == nil || fallback.Scheme != "https" {
		t.Fatalf("%#v", fallback)
	}
}
