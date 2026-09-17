package remote_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/remote"
)

const (
	slugTech = "tech"
	queryVPS = "vps"
)

func TestWhoAmIAndUser(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/whoami", func(writer http.ResponseWriter, req *http.Request) {
		assertBearer(t, req, tokenTok)
		writeUser(t, writer, 1, "alice")
	})
	mux.HandleFunc("/api/v1/users/{id}", func(writer http.ResponseWriter, req *http.Request) {
		if req.PathValue("id") != "7" {
			t.Fatalf("id %s", req.PathValue("id"))
		}

		writeUser(t, writer, 7, "bob")
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	forum := remote.New(server.URL, tokenTok)

	info, err := forum.WhoAmI()
	if err != nil || info.Name != "alice" {
		t.Fatalf("whoami %#v %v", info, err)
	}

	user, err := forum.GetUser(7)
	if err != nil || user.ID != 7 {
		t.Fatalf("user %#v %v", user, err)
	}
}

func TestListsSearchNotify(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/categories", func(writer http.ResponseWriter, _ *http.Request) {
		encodeJSON(t, writer, []client.Category{{Slug: slugTech, Name: "技术"}})
	})
	mux.HandleFunc("/api/v1/posts", func(writer http.ResponseWriter, req *http.Request) {
		writeList(t, writer, req)
	})
	mux.HandleFunc("/api/v1/posts/{id}", func(writer http.ResponseWriter, req *http.Request) {
		writePost(t, writer, req)
	})
	mux.HandleFunc("/api/v1/search", func(writer http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("q") != queryVPS || req.URL.Query().Get("page") != "2" {
			t.Fatalf("search %s", req.URL.RawQuery)
		}

		result := new(client.SearchResult)
		result.Query = queryVPS
		result.Page = 2
		encodeJSON(t, writer, result)
	})
	mux.HandleFunc("/api/v1/notifications", func(writer http.ResponseWriter, _ *http.Request) {
		if _, err := writer.Write([]byte("null")); err != nil {
			t.Fatal(err)
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	forum := remote.New(server.URL, tokenTok)
	assertRemoteLists(t, forum)
	assertRemoteSearchNotify(t, forum)
}

func assertRemoteLists(t *testing.T, forum *remote.Client) {
	t.Helper()

	cats, err := forum.Categories()
	if err != nil || len(cats) != 1 || cats[0].Slug != slugTech {
		t.Fatalf("cats %#v %v", cats, err)
	}

	latest, err := forum.LatestPosts(2)
	if err != nil || latest.Page != 2 {
		t.Fatalf("latest %#v %v", latest, err)
	}

	board, err := forum.CategoryPosts(slugTech, 0)
	if err != nil || board.Page != 1 {
		t.Fatalf("category %#v %v", board, err)
	}

	post, err := forum.GetPost(9, 3)
	if err != nil || post.Page != 3 {
		t.Fatalf("post %#v %v", post, err)
	}
}

func assertRemoteSearchNotify(t *testing.T, forum *remote.Client) {
	t.Helper()

	search, err := forum.Search(queryVPS, 2)
	if err != nil || search.Query != queryVPS {
		t.Fatalf("search %#v %v", search, err)
	}

	notes, err := forum.Notifications()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff([]client.Notification{}, notes); diff != "" {
		t.Fatal(diff)
	}
}

func TestGetPostAll(t *testing.T) {
	t.Parallel()

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		hits++
		page := queryPage(t, req)
		detail := new(client.PostDetail)
		detail.ID = 9
		detail.Page = page
		detail.Pages = 2

		if page <= 2 {
			detail.Floors = []client.Floor{{Number: page, Author: "a", Markdown: "x"}}
		}

		encodeJSON(t, writer, detail)
	}))
	t.Cleanup(server.Close)

	got, err := remote.New(server.URL, "").GetPostAll(9)
	if err != nil {
		t.Fatal(err)
	}

	if hits != 3 || len(got.Floors) != 2 || got.Pages != 2 {
		t.Fatalf("hits=%d floors=%d pages=%d", hits, len(got.Floors), got.Pages)
	}
}

func TestCategoriesAndNotifyErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		encodeAPIError(t, writer, http.StatusBadGateway, "down")
	}))
	t.Cleanup(server.Close)

	forum := remote.New(server.URL, "")
	if _, err := forum.Categories(); err == nil {
		t.Fatal("expected categories error")
	}

	if _, err := forum.Notifications(); err == nil {
		t.Fatal("expected notifications error")
	}
}

func TestGetPostAllError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		encodeAPIError(t, writer, http.StatusNotFound, "not found")
	}))
	t.Cleanup(server.Close)

	if _, err := remote.New(server.URL, "").GetPostAll(1); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPostAllMaxPages(t *testing.T) {
	t.Parallel()

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		hits++
		page := queryPage(t, req)
		detail := new(client.PostDetail)
		detail.ID = 1
		detail.Page = page
		detail.Pages = client.MaxPostPages
		detail.Floors = []client.Floor{{Number: page, Markdown: "x"}}
		encodeJSON(t, writer, detail)
	}))
	t.Cleanup(server.Close)

	got, err := remote.New(server.URL, "").GetPostAll(1)
	if err != nil {
		t.Fatal(err)
	}

	if hits != client.MaxPostPages || len(got.Floors) != client.MaxPostPages {
		t.Fatalf("hits=%d floors=%d", hits, len(got.Floors))
	}
}

func TestGetPostAllEmptyFirstPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		encodeJSON(t, writer, new(client.PostDetail))
	}))
	t.Cleanup(server.Close)

	got, err := remote.New(server.URL, "").GetPostAll(1)
	if err != nil || got == nil || len(got.Floors) != 0 {
		t.Fatalf("%#v %v", got, err)
	}
}

func writeUser(t *testing.T, writer http.ResponseWriter, id int, name string) {
	t.Helper()

	info := new(client.UserInfo)
	info.ID = id
	info.Name = name
	encodeJSON(t, writer, info)
}

func writeList(t *testing.T, writer http.ResponseWriter, req *http.Request) {
	t.Helper()

	filter := req.URL.Query().Get("filter")
	page := req.URL.Query().Get("page")
	slug := req.URL.Query().Get("slug")
	list := new(client.PostList)

	switch {
	case filter == "latest" && page == "2":
		list.Page = 2
	case filter == "category" && slug == slugTech && page == "":
		list.Page = 1
	default:
		t.Fatalf("posts query %s", req.URL.RawQuery)
	}

	encodeJSON(t, writer, list)
}

func writePost(t *testing.T, writer http.ResponseWriter, req *http.Request) {
	t.Helper()

	if req.PathValue("id") != "9" || req.URL.Query().Get("page") != "3" {
		t.Fatalf("post %s %s", req.PathValue("id"), req.URL.RawQuery)
	}

	detail := new(client.PostDetail)
	detail.ID = 9
	detail.Page = 3
	encodeJSON(t, writer, detail)
}

func assertBearer(t *testing.T, req *http.Request, token string) {
	t.Helper()

	if got := req.Header.Get("Authorization"); got != "Bearer "+token {
		t.Fatalf("auth %q", got)
	}
}

func queryPage(t *testing.T, req *http.Request) int {
	t.Helper()

	page, err := strconv.Atoi(req.URL.Query().Get("page"))
	if err != nil {
		t.Fatal(err)
	}

	return page
}
