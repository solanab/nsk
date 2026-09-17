package server_test

import (
	"net/http"
	"testing"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/server"
)

func TestMeAndUser(t *testing.T) {
	t.Parallel()

	stub := newStub()
	handler := mustServer(t, server.Config{Forum: stub}).Handler()

	var me client.UserInfo
	decodeJSON(t, doGET(t, handler, server.PathMe, ""), &me)

	if me.Name != nameAlice {
		t.Fatal("me")
	}

	var user client.UserInfo
	decodeJSON(t, doGET(t, handler, pathUsers42, ""), &user)

	if user.ID != 1 || stub.lastID != 42 {
		t.Fatal("user")
	}
}

func TestListsAndPost(t *testing.T) {
	t.Parallel()

	stub := newStub()
	handler := mustServer(t, server.Config{Forum: stub}).Handler()

	var cats []client.Category
	decodeJSON(t, doGET(t, handler, server.PathCategories, ""), &cats)

	if len(cats) != 1 {
		t.Fatal("cats")
	}

	var latest client.PostList
	decodeJSON(t, doGET(t, handler, server.PathPosts+"?filter=latest&page=2", ""), &latest)

	if latest.Page != 1 || stub.lastPage != 2 {
		t.Fatal("latest")
	}

	empty := doGET(t, handler, server.PathPosts, "")
	if empty.Code != http.StatusOK {
		t.Fatalf("empty filter %d", empty.Code)
	}

	board := doGET(t, handler, server.PathPosts+"?filter=category&slug="+slugTech+"&page=3", "")
	if board.Code != http.StatusOK || stub.lastSlug != slugTech || stub.lastPage != 3 {
		t.Fatalf("category %s %d", stub.lastSlug, stub.lastPage)
	}

	var post client.PostDetail
	decodeJSON(t, doGET(t, handler, pathPostNine, ""), &post)

	if post.ID != 9 || stub.lastPage != 2 {
		t.Fatal("post")
	}
}

func TestSearchAndNotify(t *testing.T) {
	t.Parallel()

	stub := newStub()
	handler := mustServer(t, server.Config{Forum: stub}).Handler()

	var result client.SearchResult
	decodeJSON(t, doGET(t, handler, server.PathSearch+"?q="+queryVPS+"&page=4", ""), &result)

	if result.Query != queryVPS || stub.lastPage != 4 {
		t.Fatal("search")
	}

	var notes []client.Notification
	decodeJSON(t, doGET(t, handler, server.PathNotifications, ""), &notes)

	if len(notes) != 1 {
		t.Fatal("notes")
	}
}

func TestReadValidation(t *testing.T) {
	t.Parallel()

	handler := mustServer(t, server.Config{Forum: newStub()}).Handler()

	cases := []struct {
		path string
		code int
		msg  string
	}{
		{"/api/v1/users/abc", http.StatusBadRequest, errInvalidID},
		{"/api/v1/users/0", http.StatusBadRequest, errInvalidID},
		{"/api/v1/posts/x", http.StatusBadRequest, errInvalidID},
		{server.PathSearch, http.StatusBadRequest, "搜索词为空"},
		{server.PathPosts + "?filter=other", http.StatusBadRequest, "无效 filter"},
		{server.PathPosts + "?filter=category", http.StatusBadRequest, "无效 filter"},
		{"/api/v1/posts/9/all", http.StatusNotFound, ""},
	}
	for _, testCase := range cases {
		rec := doGET(t, handler, testCase.path, "")
		if rec.Code != testCase.code {
			t.Fatalf("%s: %d", testCase.path, rec.Code)
		}

		if testCase.msg != "" && errorMsg(t, rec) != testCase.msg {
			t.Fatalf("%s body", testCase.path)
		}
	}
}

func TestForumStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		code int
	}{
		{client.ErrNotFound, http.StatusNotFound},
		{client.ErrAuthRequired, http.StatusUnauthorized},
		{client.ErrExpiredCookie, http.StatusUnauthorized},
		{client.ErrForbidden, http.StatusForbidden},
		{client.ErrRateLimited, http.StatusTooManyRequests},
		{errUpstream, http.StatusBadGateway},
	}
	for _, testCase := range cases {
		stub := newStub()
		stub.err = testCase.err
		rec := doGET(t, mustServer(t, server.Config{Forum: stub}).Handler(), server.PathWhoAmI, "")

		if rec.Code != testCase.code {
			t.Fatalf("%v: %d", testCase.err, rec.Code)
		}
	}
}

func TestNilNotifications(t *testing.T) {
	t.Parallel()

	stub := newStub()
	stub.notes = nil
	rec := doGET(t, mustServer(t, server.Config{Forum: stub}).Handler(), server.PathNotifications, "")

	var notes []client.Notification
	decodeJSON(t, rec, &notes)

	if notes == nil || len(notes) != 0 {
		t.Fatalf("%#v", notes)
	}
}

func TestQueryIntGarbage(t *testing.T) {
	t.Parallel()

	stub := newStub()
	doGET(t, mustServer(t, server.Config{Forum: stub}).Handler(), server.PathPosts+"?page=nope", "")

	if stub.lastPage != 0 {
		t.Fatalf("page %d", stub.lastPage)
	}
}

func TestPostsForumError(t *testing.T) {
	t.Parallel()

	stub := newStub()
	stub.err = client.ErrNotFound
	rec := doGET(t, mustServer(t, server.Config{Forum: stub}).Handler(), server.PathPosts+"?filter=latest", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("%d", rec.Code)
	}
}
