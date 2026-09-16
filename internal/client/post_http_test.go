package client_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

func postDoer(t *testing.T, name string) client.DoerFunc {
	t.Helper()

	return postStatusDoer(t, name, http.StatusOK, nil)
}

func postStatusDoer(t *testing.T, name string, status int, extra http.Header) client.DoerFunc {
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

type postProbe struct {
	uri  string
	pjwt bool
	ua   string
}

func (probe *postProbe) wrap(inner client.DoerFunc) client.DoerFunc {
	return func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path != "/" {
			probe.uri = req.URL.RequestURI()
			probe.pjwt = requestHasPJWT(req)
			probe.ua = headerVal(req.Header, "user-agent")
		}

		return inner(req)
	}
}

func TestGetPostPage1(t *testing.T) {
	t.Parallel()

	probe := new(postProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(postDoer(t, htmlPostPage)))

	detail, err := forum.GetPost(postID703863, 1)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/post-703863-1" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if !probe.pjwt {
		t.Fatal("post request missing pjwt cookie")
	}

	if probe.ua != wantUA {
		t.Fatalf("ua = %q", probe.ua)
	}

	if detail.Title != postTitle703863 || len(detail.Floors) != 11 {
		t.Fatalf("title=%q floors=%d", detail.Title, len(detail.Floors))
	}
}

func TestGetPostPage0UsesPage1(t *testing.T) {
	t.Parallel()

	probe := new(postProbe)
	forum := openOK(t, pjwtFile(t), probe.wrap(postDoer(t, htmlPostEmpty)))

	detail, err := forum.GetPost(9, 0)
	if err != nil {
		t.Fatal(err)
	}

	if probe.uri != "/post-9-1" {
		t.Fatalf("uri = %q", probe.uri)
	}

	if detail.Page != 1 {
		t.Fatalf("page=%d", detail.Page)
	}
}

func TestGetPostCloudflare(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), postDoer(t, "cf-challenge.html"))

	_, err := forum.GetPost(1, 1)
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetPostNotFound(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), postStatusDoer(t, htmlPostEmpty, http.StatusNotFound, nil))

	_, err := forum.GetPost(1, 1)
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetPostRateLimited(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), postStatusDoer(t, htmlPostEmpty, http.StatusTooManyRequests, nil))

	_, err := forum.GetPost(1, 1)
	if !errors.Is(err, client.ErrRateLimited) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetPostDoError(t *testing.T) {
	t.Parallel()

	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, fixture(t, htmlUser), nil), nil
		}

		return nil, errBoom
	})
	forum := openOK(t, pjwtFile(t), doer)

	_, err := forum.GetPost(1, 1)
	if err == nil || !strings.Contains(err.Error(), "请求帖子") {
		t.Fatalf("err = %v", err)
	}
}

func TestGetPostParseError(t *testing.T) {
	forum := openOK(t, pjwtFile(t), postDoer(t, htmlPostEmpty))
	t.Cleanup(client.StubParseHTML(func(io.Reader) (*goquery.Document, error) {
		return nil, errHTML
	}))

	_, err := forum.GetPost(1, 1)
	if err == nil || !strings.Contains(err.Error(), "解析帖子") {
		t.Fatalf("err = %v", err)
	}
}

func TestGetPostAllConcat(t *testing.T) {
	t.Parallel()

	home := fixture(t, htmlUser)
	pages := map[string][]byte{
		"/post-7-1": fixture(t, htmlPostMagic),
		"/post-7-2": fixture(t, htmlPostPage2),
		"/post-7-3": fixture(t, htmlPostEmpty),
	}

	var paths []string

	forum := openOK(t, pjwtFile(t), mapPostDoer(t, home, pages, &paths))

	detail, err := forum.GetPostAll(7)
	if err != nil {
		t.Fatal(err)
	}

	assertConcatPost(t, detail, paths)
}

func mapPostDoer(t *testing.T, home []byte, pages map[string][]byte, paths *[]string) client.DoerFunc {
	t.Helper()

	return func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, home, nil), nil
		}

		*paths = append(*paths, req.URL.Path)
		if body, ok := pages[req.URL.Path]; ok {
			return htmlResponse(http.StatusOK, body, nil), nil
		}

		return htmlResponse(http.StatusNotFound, nil, nil), nil
	}
}

func assertConcatPost(t *testing.T, detail *client.PostDetail, paths []string) {
	t.Helper()

	if strings.Join(paths, ",") != "/post-7-1,/post-7-2,/post-7-3" {
		t.Fatalf("paths=%v", paths)
	}

	if len(detail.Floors) != 2 || detail.Floors[0].Number != 0 || detail.Floors[1].Number != 11 {
		t.Fatalf("floors=%#v", detail.Floors)
	}

	if detail.Page != 1 || detail.Title != "magic post" {
		t.Fatalf("page=%d title=%q", detail.Page, detail.Title)
	}
}

func TestGetPostAllEmptyFirst(t *testing.T) {
	t.Parallel()

	forum := openOK(t, pjwtFile(t), postDoer(t, htmlPostEmpty))

	detail, err := forum.GetPostAll(9)
	if err != nil {
		t.Fatal(err)
	}

	if len(detail.Floors) != 0 {
		t.Fatalf("floors=%#v", detail.Floors)
	}
}

func TestGetPostAllErrorAborts(t *testing.T) {
	t.Parallel()

	home := fixture(t, htmlUser)
	page1 := fixture(t, htmlPostMagic)

	var hits int

	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, home, nil), nil
		}

		hits++

		if req.URL.Path == "/post-7-1" {
			return htmlResponse(http.StatusOK, page1, nil), nil
		}

		return htmlResponse(http.StatusTooManyRequests, nil, nil), nil
	})
	forum := openOK(t, pjwtFile(t), doer)

	detail, err := forum.GetPostAll(7)
	if !errors.Is(err, client.ErrRateLimited) {
		t.Fatalf("err = %v", err)
	}

	if detail != nil {
		t.Fatalf("partial = %#v", detail)
	}

	if hits != 2 {
		t.Fatalf("hits=%d (retried 429?)", hits)
	}
}

func TestGetPostAllCapsAtMax(t *testing.T) {
	t.Parallel()

	home := fixture(t, htmlUser)

	var maxPage, hits int

	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL != nil && req.URL.Path == "/" {
			return htmlResponse(http.StatusOK, home, nil), nil
		}

		hits++

		var page int
		if _, err := fmt.Sscanf(req.URL.Path, "/post-9-%d", &page); err != nil {
			t.Fatalf("path=%q", req.URL.Path)
		}

		if page > maxPage {
			maxPage = page
		}

		body := fmt.Sprintf(
			`<html><h1><a class="post-title-link">t</a></h1>`+
				`<div class="content-item"><a class="author-name">a</a>`+
				`<a class="floor-link">#%d</a>`+
				`<article class="post-content"><p>x</p></article></div></html>`,
			page-1,
		)

		return htmlResponse(http.StatusOK, []byte(body), nil), nil
	})
	forum := openOK(t, pjwtFile(t), doer)

	detail, err := forum.GetPostAll(9)
	if err != nil {
		t.Fatal(err)
	}

	if maxPage != client.MaxPostPages || hits != client.MaxPostPages {
		t.Fatalf("maxPage=%d hits=%d", maxPage, hits)
	}

	if len(detail.Floors) != client.MaxPostPages {
		t.Fatalf("floors=%d", len(detail.Floors))
	}
}
