package client_test

import (
	"errors"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

func searchDoer(t *testing.T, name string) client.DoerFunc {
	t.Helper()

	return searchStatusDoer(t, name, http.StatusOK, nil)
}

func searchStatusDoer(t *testing.T, name string, status int, extra http.Header) client.DoerFunc {
	t.Helper()

	home := fixture(t, htmlUser)
	body := fixture(t, name)

	return func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, home, nil), nil
		}

		return htmlResponse(status, body, extra), nil
	}
}

func TestSearchPage1(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(searchDoer(t, htmlSearchPage)))

	got, err := forum.Search(queryVPS, 1)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/search?q="+queryVPS {
		t.Fatalf("uri = %q", probe.uri)
	}

	if !probe.pjwt {
		t.Fatal("search request missing pjwt cookie")
	}

	if probe.ua != wantUA {
		t.Fatalf("ua = %q", probe.ua)
	}

	if got.Query != queryVPS || got.Page != 1 || len(got.Posts) != 2 {
		t.Fatalf("query=%q page=%d posts=%d", got.Query, got.Page, len(got.Posts))
	}
}

func TestSearchPage0UsesPage1(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(searchDoer(t, htmlSearchEmpty)))

	got, err := forum.Search(queryVPS, 0)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/search?q="+queryVPS {
		t.Fatalf("uri = %q", probe.uri)
	}

	if got.Page != 1 || len(got.Posts) != 0 {
		t.Fatalf("page=%d posts=%d", got.Page, len(got.Posts))
	}
}

func TestSearchPage2(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(searchDoer(t, htmlSearchEmpty)))

	got, err := forum.Search(queryVPS, 2)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/search?q="+queryVPS+"&page=2" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if got.Page != 2 {
		t.Fatalf("page=%d", got.Page)
	}
}

func TestSearchQueryEncoded(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(searchDoer(t, htmlSearchEmpty)))

	_, err := forum.Search("hello world", 1)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/search?q=hello+world" {
		t.Fatalf("uri = %q", probe.uri)
	}
}

func TestSearchQueryAmpersand(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(searchDoer(t, htmlSearchEmpty)))

	_, err := forum.Search("a&b", 1)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/search?q=a%26b" {
		t.Fatalf("uri = %q", probe.uri)
	}
}

func TestSearchEmptyQueryNoHTTP(t *testing.T) {
	t.Parallel()

	_, err := client.Unstarted().Search("  ", 1)
	if err == nil || !strings.Contains(err.Error(), "搜索词为空") {
		t.Fatalf("err = %v", err)
	}
}

func TestSearchGuestNoListFallback(t *testing.T) {
	t.Parallel()

	var paths []string

	home := fixture(t, htmlUser)
	guest := fixture(t, htmlSearchGuest)
	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL != nil {
			paths = append(paths, req.URL.Path)
			if req.URL.Path == "/" {
				return htmlResponse(http.StatusOK, home, nil), nil
			}
		}

		return htmlResponse(http.StatusOK, guest, nil), nil
	})
	forum := openOK(t, pjwtFile(t), doer)

	got, err := forum.Search(queryVPS, 1)
	if !errors.Is(err, client.ErrExpiredCookie) {
		t.Fatalf("err = %v result=%#v", err, got)
	}

	if got != nil {
		t.Fatalf("guest search returned list: %#v", got)
	}

	for _, path := range paths {
		if strings.HasPrefix(path, "/page-") {
			t.Fatalf("search fell back to latest list: %q", paths)
		}
	}
}

func TestSearchCloudflare(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), searchDoer(t, "cf-challenge.html"))

	_, err := forum.Search(queryVPS, 1)
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func TestSearchNotFound(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), searchStatusDoer(t, htmlSearchEmpty, http.StatusNotFound, nil))

	_, err := forum.Search(queryVPS, 1)
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestSearchDoError(t *testing.T) {
	t.Parallel()

	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, fixture(t, htmlUser), nil), nil
		}

		return nil, errBoom
	})
	forum := openOK(t, pjwtFile(t), doer)

	_, err := forum.Search(queryVPS, 1)
	if err == nil || !strings.Contains(err.Error(), "请求搜索") {
		t.Fatalf("err = %v", err)
	}
}

func TestSearchTrimQuery(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(searchDoer(t, htmlSearchEmpty)))

	got, err := forum.Search("  vps  ", 1)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/search?q="+queryVPS {
		t.Fatalf("uri = %q", probe.uri)
	}

	if got.Query != queryVPS {
		t.Fatalf("query=%q", got.Query)
	}
}
