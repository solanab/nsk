package client_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

func TestWarmupUserPresent(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), userDoer(t))

	info, err := forum.WhoAmI()
	if err != nil {
		t.Fatal(err)
	}

	if info.ID != 42 || info.Name != nameAlice || info.Chicken != 7 || info.Level != levelTwo {
		t.Fatalf("%#v", info)
	}
}

func assertUnchanged(t *testing.T, path string, orig []byte, err error, want error) {
	t.Helper()

	if !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}

	got, readErr := os.ReadFile(path) //nolint:gosec // test cookie file
	if readErr != nil {
		t.Fatal(readErr)
	}

	if string(got) != string(orig) {
		t.Fatalf("file changed: %q", got)
	}
}

func TestWarmupGuestDoesNotOverwrite(t *testing.T) {
	t.Parallel()

	orig := []byte("pjwt=" + pjwtToken + "\n")
	path := writeCookie(t, string(orig))
	err := openErr(t, path, fixtureDoer(t, "homepage-guest.html", http.StatusOK, nil))
	assertUnchanged(t, path, orig, err, client.ErrExpiredCookie)
}

func TestWarmupNoUserScript(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), fixtureDoer(t, "homepage-nouser.html", http.StatusOK, nil))
	if !errors.Is(err, client.ErrExpiredCookie) {
		t.Fatalf("err = %v", err)
	}
}

func TestWarmupBadConfig(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), fixtureDoer(t, "homepage-bad-config.html", http.StatusOK, nil))
	if !errors.Is(err, client.ErrExpiredCookie) {
		t.Fatalf("err = %v", err)
	}
}

func TestWarmupNoObject(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), fixtureDoer(t, "homepage-nobj.html", http.StatusOK, nil))
	if !errors.Is(err, client.ErrExpiredCookie) {
		t.Fatalf("err = %v", err)
	}
}

func TestWarmupInvalidJSON(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), fixtureDoer(t, "homepage-invalid-json.html", http.StatusOK, nil))
	if !errors.Is(err, client.ErrExpiredCookie) {
		t.Fatalf("err = %v", err)
	}
}

func TestWarmupEmptyUser(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), fixtureDoer(t, "homepage-empty-user.html", http.StatusOK, nil))
	if !errors.Is(err, client.ErrExpiredCookie) {
		t.Fatalf("err = %v", err)
	}
}

func TestWarmupEscapedName(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), fixtureDoer(t, "homepage-escaped.html", http.StatusOK, nil))

	info, err := forum.WhoAmI()
	if err != nil {
		t.Fatal(err)
	}

	if info.ID != 3 || info.Name != `a"b` {
		t.Fatalf("%#v", info)
	}
}

func TestWarmupMinUser(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), fixtureDoer(t, "homepage-min.html", http.StatusOK, nil))

	info, err := forum.WhoAmI()
	if err != nil {
		t.Fatal(err)
	}

	if info.ID != 1 || info.Name != nameBob || info.Chicken != 0 || info.Level != "" {
		t.Fatalf("%#v", info)
	}
}

func TestWarmupCloudflareBodyDoesNotOverwrite(t *testing.T) {
	t.Parallel()

	orig := []byte("pjwt=" + pjwtToken + "\n")
	path := writeCookie(t, string(orig))
	err := openErr(t, path, fixtureDoer(t, "cf-challenge.html", http.StatusForbidden, nil))
	assertUnchanged(t, path, orig, err, client.ErrCloudflare)
}

func TestWarmupCloudflareHeader(t *testing.T) {
	t.Parallel()

	extra := http.Header{}
	extra.Set(headerCF, cfChallenge)

	err := openErr(t, pjwtFile(t), fixtureDoer(t, htmlUser, http.StatusForbidden, extra))
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func statusDoer(status int, body []byte) client.DoerFunc {
	return func(*http.Request) (*http.Response, error) {
		return htmlResponse(status, body, nil), nil
	}
}

func TestWarmupForbidden(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), statusDoer(http.StatusForbidden, []byte(`{"error":"no"}`)))
	if !errors.Is(err, client.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}

func TestWarmupNotFound(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), statusDoer(http.StatusNotFound, []byte("missing")))
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestWarmupRateLimited(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), statusDoer(http.StatusTooManyRequests, []byte("slow")))
	if !errors.Is(err, client.ErrRateLimited) {
		t.Fatalf("err = %v", err)
	}
}

func TestWarmupHTTP500(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), statusDoer(http.StatusInternalServerError, []byte("oops")))
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("err = %v", err)
	}
}
