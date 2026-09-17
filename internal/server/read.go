package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/solanab/nsk/internal/client"
)

func (srv *Server) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, healthBody{OK: true, Ready: true})
}

func (srv *Server) handleWhoAmI(writer http.ResponseWriter, _ *http.Request) {
	srv.replyForum(writer, func(forum client.Forum) (any, error) {
		info, err := forum.WhoAmI()

		return info, wrapErr(err)
	})
}

func (srv *Server) handleUser(writer http.ResponseWriter, req *http.Request) {
	srv.replyByID(writer, req, func(forum client.Forum, userID int) (any, error) {
		info, err := forum.GetUser(userID)

		return info, wrapErr(err)
	})
}

func (srv *Server) handleCategories(writer http.ResponseWriter, _ *http.Request) {
	srv.replyForum(writer, func(forum client.Forum) (any, error) {
		cats, err := forum.Categories()

		return cats, wrapErr(err)
	})
}

func (srv *Server) handleNotifications(writer http.ResponseWriter, _ *http.Request) {
	srv.replyForum(writer, func(forum client.Forum) (any, error) {
		notes, err := forum.Notifications()
		if notes == nil {
			notes = []client.Notification{}
		}

		return notes, wrapErr(err)
	})
}

func (srv *Server) handleSearch(writer http.ResponseWriter, req *http.Request) {
	query := strings.TrimSpace(req.URL.Query().Get("q"))
	if query == "" {
		writeError(writer, http.StatusBadRequest, "搜索词为空")

		return
	}

	srv.replyForum(writer, func(forum client.Forum) (any, error) {
		result, err := forum.Search(query, queryInt(req, "page"))

		return result, wrapErr(err)
	})
}

func (srv *Server) handlePost(writer http.ResponseWriter, req *http.Request) {
	srv.replyByID(writer, req, func(forum client.Forum, postID int) (any, error) {
		detail, err := forum.GetPost(postID, queryInt(req, "page"))

		return detail, wrapErr(err)
	})
}

func (srv *Server) handlePosts(writer http.ResponseWriter, req *http.Request) {
	filter := strings.TrimSpace(req.URL.Query().Get("filter"))
	page := queryInt(req, "page")
	slug := strings.TrimSpace(req.URL.Query().Get("slug"))

	var list *client.PostList

	err := srv.withForum(func(forum client.Forum) error {
		loaded, inner := loadPosts(forum, filter, slug, page)
		list = loaded

		return inner
	})
	if err != nil {
		if errors.Is(err, errBadFilter) {
			writeError(writer, http.StatusBadRequest, err.Error())

			return
		}

		writeForumError(writer, err)

		return
	}

	writeJSON(writer, http.StatusOK, list)
}

func loadPosts(forum client.Forum, filter, slug string, page int) (*client.PostList, error) {
	switch filter {
	case "", "latest":
		list, err := forum.LatestPosts(page)

		return list, wrapErr(err)
	case "category":
		if slug == "" {
			return nil, errBadFilter
		}

		list, err := forum.CategoryPosts(slug, page)

		return list, wrapErr(err)
	default:
		return nil, errBadFilter
	}
}
