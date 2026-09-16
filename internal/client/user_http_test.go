package client_test

import (
	"errors"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

func getInfoDoer(t *testing.T, name string) client.DoerFunc {
	t.Helper()

	return getInfoStatusDoer(t, jsonFixture(t, name), http.StatusOK, nil)
}

func getInfoStatusDoer(t *testing.T, body []byte, status int, extra http.Header) client.DoerFunc {
	t.Helper()

	home := fixture(t, htmlUser)

	return func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, home, nil), nil
		}

		return jsonResponse(status, body, extra), nil
	}
}

func TestGetUserOK(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(getInfoDoer(t, jsonGetInfoOK)))

	got, err := forum.GetUser(42)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/api/account/getInfo/42" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if !probe.pjwt {
		t.Fatal("getInfo request missing pjwt cookie")
	}

	if probe.ua != wantUA {
		t.Fatalf("ua = %q", probe.ua)
	}

	if got.ID != 42 || got.Name != nameAlice || got.Chicken != 7 || got.Level != levelTwo {
		t.Fatalf("%#v", got)
	}
}

func TestGetUserMissingDetail(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), getInfoDoer(t, jsonGetInfoNoDetail))

	got, err := forum.GetUser(99)
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}

	if got != nil {
		t.Fatalf("got = %#v", got)
	}
}

func TestGetUserEmptyObject(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), getInfoDoer(t, jsonGetInfoEmpty))

	_, err := forum.GetUser(99)
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetUserNotFoundStatus(t *testing.T) {
	t.Parallel()

	body := jsonFixture(t, jsonGetInfoEmpty)
	forum := openOK(t, pjwtFile(t), getInfoStatusDoer(t, body, http.StatusNotFound, nil))

	_, err := forum.GetUser(99)
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetUserCloudflare(t *testing.T) {
	t.Parallel()

	body := fixture(t, "cf-challenge.html")
	forum := openOK(t, pjwtFile(t), getInfoStatusDoer(t, body, http.StatusForbidden, nil))

	_, err := forum.GetUser(42)
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetUserCloudflareHeader(t *testing.T) {
	t.Parallel()

	extra := http.Header{"cf-mitigated": {cfChallenge}}
	body := jsonFixture(t, jsonGetInfoOK)
	forum := openOK(t, pjwtFile(t), getInfoStatusDoer(t, body, http.StatusOK, extra))

	_, err := forum.GetUser(42)
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetUserForbidden(t *testing.T) {
	t.Parallel()

	body := jsonFixture(t, jsonGetInfoEmpty)
	forum := openOK(t, pjwtFile(t), getInfoStatusDoer(t, body, http.StatusForbidden, nil))

	_, err := forum.GetUser(42)
	if !errors.Is(err, client.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetUserUnauthorized(t *testing.T) {
	t.Parallel()

	body := jsonFixture(t, jsonGetInfoEmpty)
	forum := openOK(t, pjwtFile(t), getInfoStatusDoer(t, body, http.StatusUnauthorized, nil))

	_, err := forum.GetUser(42)
	if err == nil || !strings.Contains(err.Error(), "unexpected HTTP status") {
		t.Fatalf("err = %v", err)
	}
}

func TestGetUserRateLimited(t *testing.T) {
	t.Parallel()

	body := jsonFixture(t, jsonGetInfoEmpty)
	forum := openOK(t, pjwtFile(t), getInfoStatusDoer(t, body, http.StatusTooManyRequests, nil))

	_, err := forum.GetUser(42)
	if !errors.Is(err, client.ErrRateLimited) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetUserDoError(t *testing.T) {
	t.Parallel()

	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, fixture(t, htmlUser), nil), nil
		}

		return nil, errBoom
	})
	forum := openOK(t, pjwtFile(t), doer)

	_, err := forum.GetUser(42)
	if err == nil || !strings.Contains(err.Error(), "请求用户") {
		t.Fatalf("err = %v", err)
	}
}

func TestGetUserInvalidIDNoHTTP(t *testing.T) {
	t.Parallel()

	for _, userID := range []int{0, -1} {
		_, err := client.Unstarted().GetUser(userID)
		if err == nil || !strings.Contains(err.Error(), "无效用户 id") {
			t.Fatalf("id=%d err=%v", userID, err)
		}
	}
}
