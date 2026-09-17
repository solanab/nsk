package server_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/server"
)

const (
	nameAlice    = "alice"
	tokenSecret  = "secret"
	slugTech     = "tech"
	queryVPS     = "vps"
	titleHello   = "hello"
	pathUsers42  = "/api/v1/users/42"
	pathPostNine = "/api/v1/posts/9?page=2"
	errInvalidID = "无效 id"
)

var errUpstream = errors.New("upstream")

func mustServer(t *testing.T, cfg server.Config) *server.Server {
	t.Helper()

	if cfg.Forum == nil {
		cfg.Forum = newStub()
	}

	srv, err := server.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	return srv
}

func doGET(t *testing.T, handler http.Handler, path, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()

	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatal(err)
	}
}

func errorMsg(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error string `json:"error"`
	}
	decodeJSON(t, rec, &body)

	return body.Error
}

func sampleUser() *client.UserInfo {
	info := new(client.UserInfo)
	info.ID = 1
	info.Name = nameAlice

	return info
}
