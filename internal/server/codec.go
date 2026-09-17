package server

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/solanab/nsk/internal/client"
)

var (
	errMissingForum = errors.New("缺少 Forum")
	errBadFilter    = errors.New("无效 filter")
	errInvalidID    = errors.New("无效 id")
)

type errorBody struct {
	Error string `json:"error"`
}

type healthBody struct {
	OK    bool `json:"ok"`
	Ready bool `json:"ready"`
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)

	if err := json.NewEncoder(writer).Encode(value); err != nil {
		return
	}
}

func wrapErr(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w", err)
}

func (srv *Server) replyForum(writer http.ResponseWriter, load func(client.Forum) (any, error)) {
	var value any

	err := srv.withForum(func(forum client.Forum) error {
		loaded, inner := load(forum)
		value = loaded

		return inner
	})
	if err != nil {
		writeForumError(writer, err)

		return
	}

	writeJSON(writer, http.StatusOK, value)
}

func (srv *Server) replyByID(
	writer http.ResponseWriter,
	req *http.Request,
	load func(client.Forum, int) (any, error),
) {
	pathID, err := parsePathID(req, "id")
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())

		return
	}

	srv.replyForum(writer, func(forum client.Forum) (any, error) {
		return load(forum, pathID)
	})
}

func writeError(writer http.ResponseWriter, status int, msg string) {
	writeJSON(writer, status, errorBody{Error: msg})
}

func writeForumError(writer http.ResponseWriter, err error) {
	writeError(writer, forumStatus(err), err.Error())
}

func forumStatus(err error) int {
	switch {
	case errors.Is(err, client.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, client.ErrAuthRequired), errors.Is(err, client.ErrExpiredCookie):
		return http.StatusUnauthorized
	case errors.Is(err, client.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, client.ErrRateLimited):
		return http.StatusTooManyRequests
	default:
		return http.StatusBadGateway
	}
}

func (srv *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.URL.Path == pathHealth || srv.token == "" {
			next.ServeHTTP(writer, req)

			return
		}

		got := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(srv.token)) != 1 {
			writeError(writer, http.StatusUnauthorized, "unauthorized")

			return
		}

		next.ServeHTTP(writer, req)
	})
}

func parsePathID(req *http.Request, name string) (int, error) {
	raw := req.PathValue(name)

	pathID, err := strconv.Atoi(raw)
	if err != nil || pathID <= 0 {
		return 0, errInvalidID
	}

	return pathID, nil
}

func queryInt(req *http.Request, name string) int {
	raw := strings.TrimSpace(req.URL.Query().Get(name))
	if raw == "" {
		return 0
	}

	number, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}

	return number
}
