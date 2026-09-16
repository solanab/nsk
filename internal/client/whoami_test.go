package client_test

import (
	"errors"
	"testing"

	"github.com/solanab/nsk/internal/client"
)

func TestWhoAmIUnstarted(t *testing.T) {
	t.Parallel()

	_, err := client.Unstarted().WhoAmI()
	if !errors.Is(err, client.ErrExpiredCookie) {
		t.Fatalf("err = %v", err)
	}
}

func TestExportPJWTUnstarted(t *testing.T) {
	t.Parallel()

	_, err := client.Unstarted().ExportPJWT()
	if err == nil || err.Error() != "cookie 中没有 pjwt" {
		t.Fatalf("err = %v", err)
	}
}

func TestExportJSONUnstarted(t *testing.T) {
	t.Parallel()

	raw, err := client.Unstarted().ExportCookiesJSON()
	if err != nil {
		t.Fatal(err)
	}

	if raw != "[]" {
		t.Fatalf("raw = %q", raw)
	}
}

func TestWhoAmICopy(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), userDoer(t))

	info, err := forum.WhoAmI()
	if err != nil {
		t.Fatal(err)
	}

	info.Name = "mutated"

	again, err := forum.WhoAmI()
	if err != nil {
		t.Fatal(err)
	}

	if again.Name != nameAlice {
		t.Fatalf("WhoAmI mutated: %#v", again)
	}
}

func TestSentinelErrorsExist(t *testing.T) {
	t.Parallel()

	if client.ErrAuthRequired.Error() != "需要登录" {
		t.Fatalf("%v", client.ErrAuthRequired)
	}
}

func TestHTTPSURLFallback(t *testing.T) {
	t.Parallel()

	parsed := client.HTTPSURL("%")
	if parsed == nil || parsed.Scheme != "https" {
		t.Fatalf("%#v", parsed)
	}
}
