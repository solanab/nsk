package client_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

func listDoer(t *testing.T, listName string) client.DoerFunc {
	t.Helper()

	return listStatusDoer(t, listName, http.StatusOK, nil)
}

func listStatusDoer(t *testing.T, listName string, status int, extra http.Header) client.DoerFunc {
	t.Helper()

	home := fixture(t, htmlUser)
	listBody := fixture(t, listName)

	return func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, home, nil), nil
		}

		return htmlResponse(status, listBody, extra), nil
	}
}

type listProbe struct {
	uri  string
	pjwt bool
	ua   string
}

func (probe *listProbe) wrap(inner client.DoerFunc) client.DoerFunc {
	return func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path != "/" {
			probe.uri = req.URL.RequestURI()
			probe.pjwt = requestHasPJWT(req)
			probe.ua = headerVal(req.Header, "user-agent")
		}

		return inner(req)
	}
}

func TestLatestPostsPage1(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(listDoer(t, htmlListPage)))

	list, err := forum.LatestPosts(1)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/page-1" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if !probe.pjwt {
		t.Fatal("list request missing pjwt cookie")
	}

	if probe.ua != wantUA {
		t.Fatalf("ua = %q", probe.ua)
	}

	if len(list.Posts) != client.ListPerPage {
		t.Fatalf("posts=%d", len(list.Posts))
	}
}

func TestLatestPostsPage0UsesPage1(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(listDoer(t, htmlListEmpty)))

	list, err := forum.LatestPosts(0)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/page-1" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if list.Page != 1 || len(list.Posts) != 0 {
		t.Fatalf("page=%d posts=%d", list.Page, len(list.Posts))
	}
}

func TestLatestPostsPage2(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(listDoer(t, htmlListEmpty)))

	list, err := forum.LatestPosts(2)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/page-2" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if list.Page != 2 {
		t.Fatalf("page=%d", list.Page)
	}
}

func TestCategoryPostsKnownSlug(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(listDoer(t, htmlListDedup)))

	list, err := forum.CategoryPosts(slugTech, 1)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/categories/"+slugTech+"?page=1" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if !probe.pjwt {
		t.Fatal("list request missing pjwt cookie")
	}

	if list.Page != 1 || len(list.Posts) != 2 {
		t.Fatalf("page=%d posts=%d", list.Page, len(list.Posts))
	}
}

func TestCategoryPostsPage0(t *testing.T) {
	t.Parallel()

	probe := new(listProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(listDoer(t, htmlListEmpty)))

	_, err := forum.CategoryPosts("meaningless", 0)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/categories/meaningless?page=1" {
		t.Fatalf("uri = %q", probe.uri)
	}
}

func TestCategoryPostsUnknownSlugNoHTTP(t *testing.T) {
	t.Parallel()

	_, err := client.Unstarted().CategoryPosts("no-such-board", 1)
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestLatestPostsCloudflare(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), listDoer(t, "cf-challenge.html"))

	_, err := forum.LatestPosts(1)
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func TestLatestPostsCloudflareHeader(t *testing.T) {
	t.Parallel()

	extra := http.Header{}
	extra.Set(headerCF, cfChallenge)
	forum := openOK(t, pjwtFile(t), listStatusDoer(t, htmlListEmpty, http.StatusOK, extra))

	_, err := forum.LatestPosts(1)
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func TestLatestPostsNotFound(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), listStatusDoer(t, htmlListEmpty, http.StatusNotFound, nil))

	_, err := forum.LatestPosts(1)
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestLatestPostsForbidden(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), listStatusDoer(t, htmlListEmpty, http.StatusForbidden, nil))

	_, err := forum.LatestPosts(1)
	if !errors.Is(err, client.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}

func TestLatestPostsRateLimited(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), listStatusDoer(t, htmlListEmpty, http.StatusTooManyRequests, nil))

	_, err := forum.LatestPosts(1)
	if !errors.Is(err, client.ErrRateLimited) {
		t.Fatalf("err = %v", err)
	}
}

func TestLatestPostsHTTP500(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), listStatusDoer(t, htmlListEmpty, http.StatusInternalServerError, nil))

	err := mustListErr(t, forum)
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("err = %v", err)
	}
}

func TestLatestPostsDoError(t *testing.T) {
	t.Parallel()

	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, fixture(t, htmlUser), nil), nil
		}

		return nil, errBoom
	})
	forum := openOK(t, pjwtFile(t), doer)

	err := mustListErr(t, forum)
	if err == nil || !strings.Contains(err.Error(), "请求列表") {
		t.Fatalf("err = %v", err)
	}
}

func TestLatestPostsParseError(t *testing.T) {
	forum := openOK(t, pjwtFile(t), listDoer(t, htmlListEmpty))
	t.Cleanup(client.StubParseHTML(func(io.Reader) (*goquery.Document, error) {
		return nil, errHTML
	}))

	err := mustListErr(t, forum)
	if err == nil || !strings.Contains(err.Error(), "解析列表") {
		t.Fatalf("err = %v", err)
	}
}

func mustListErr(t *testing.T, forum *client.Client) error {
	t.Helper()

	_, err := forum.LatestPosts(1)
	if err == nil {
		t.Fatal("expected error")
	}

	return err //nolint:wrapcheck // test helper returns the Forum error unchanged
}
