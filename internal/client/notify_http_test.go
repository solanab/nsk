package client_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

func notifyDoer(t *testing.T, name string) client.DoerFunc {
	t.Helper()

	return notifyStatusDoer(t, jsonFixture(t, name), http.StatusOK, nil)
}

func notifyStatusDoer(t *testing.T, body []byte, status int, extra http.Header) client.DoerFunc {
	t.Helper()

	home := fixture(t, htmlUser)

	return func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, home, nil), nil
		}

		return jsonResponse(status, body, extra), nil
	}
}

func TestNotificationsOK(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(notifyDoer(t, jsonNotifyOK)))

	got, err := forum.Notifications()
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/api/notification/at-me/list" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if !probe.pjwt {
		t.Fatal("notify request missing pjwt cookie")
	}

	if probe.ua != wantUA {
		t.Fatalf("ua = %q", probe.ua)
	}

	if diff := cmp.Diff(sampleNotify(), got); diff != "" {
		t.Fatal(diff)
	}
}

func TestNotificationsEmpty(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), notifyDoer(t, jsonNotifyEmpty))

	got, err := forum.Notifications()
	if err != nil {
		t.Fatal(err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("got = %#v", got)
	}
}

func TestNotificationsGuest(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), notifyDoer(t, jsonNotifyGuest))

	got, err := forum.Notifications()
	if !errors.Is(err, client.ErrAuthRequired) {
		t.Fatalf("err = %v", err)
	}

	if got != nil {
		t.Fatalf("got = %#v", got)
	}
}

func TestNotificationsNotFoundStatus(t *testing.T) {
	t.Parallel()

	body := jsonFixture(t, jsonNotifyEmpty)
	forum := openOK(t, pjwtFile(t), notifyStatusDoer(t, body, http.StatusNotFound, nil))

	_, err := forum.Notifications()
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestNotificationsCloudflare(t *testing.T) {
	t.Parallel()

	body := fixture(t, "cf-challenge.html")
	forum := openOK(t, pjwtFile(t), notifyStatusDoer(t, body, http.StatusForbidden, nil))

	_, err := forum.Notifications()
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func TestNotificationsCloudflareHeader(t *testing.T) {
	t.Parallel()

	extra := http.Header{headerCF: {cfChallenge}}
	body := jsonFixture(t, jsonNotifyOK)
	forum := openOK(t, pjwtFile(t), notifyStatusDoer(t, body, http.StatusOK, extra))

	_, err := forum.Notifications()
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func TestNotificationsForbidden(t *testing.T) {
	t.Parallel()

	body := jsonFixture(t, jsonNotifyEmpty)
	forum := openOK(t, pjwtFile(t), notifyStatusDoer(t, body, http.StatusForbidden, nil))

	_, err := forum.Notifications()
	if !errors.Is(err, client.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}

func TestNotificationsUnauthorized(t *testing.T) {
	t.Parallel()

	body := jsonFixture(t, jsonNotifyGuest)
	forum := openOK(t, pjwtFile(t), notifyStatusDoer(t, body, http.StatusUnauthorized, nil))

	_, err := forum.Notifications()
	if err == nil || !strings.Contains(err.Error(), "unexpected HTTP status") {
		t.Fatalf("err = %v", err)
	}
}

func TestNotificationsRateLimited(t *testing.T) {
	t.Parallel()

	body := jsonFixture(t, jsonNotifyEmpty)
	forum := openOK(t, pjwtFile(t), notifyStatusDoer(t, body, http.StatusTooManyRequests, nil))

	_, err := forum.Notifications()
	if !errors.Is(err, client.ErrRateLimited) {
		t.Fatalf("err = %v", err)
	}
}

func TestNotificationsDoError(t *testing.T) {
	t.Parallel()

	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, fixture(t, htmlUser), nil), nil
		}

		return nil, errBoom
	})
	forum := openOK(t, pjwtFile(t), doer)

	_, err := forum.Notifications()
	if err == nil || !strings.Contains(err.Error(), "请求通知") {
		t.Fatalf("err = %v", err)
	}
}
