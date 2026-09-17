package remote_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/remote"
)

const tokenTok = "tok"

var (
	errClose = errors.New("close")
	errRead  = errors.New("read")
	errBoom  = errors.New("boom")
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type closeBody struct {
	io.Reader

	err error
}

func (body closeBody) Close() error {
	return body.err
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errRead
}

func encodeJSON(t *testing.T, writer http.ResponseWriter, value any) {
	t.Helper()

	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatal(err)
	}
}

func encodeHealth(t *testing.T, writer http.ResponseWriter, ok, ready bool) {
	t.Helper()

	encodeJSON(t, writer, map[string]any{"ok": ok, "ready": ready})
}

func encodeAPIError(t *testing.T, writer http.ResponseWriter, status int, msg string) {
	t.Helper()

	writer.WriteHeader(status)
	encodeJSON(t, writer, map[string]string{"error": msg})
}

func TestTimeoutIs60s(t *testing.T) {
	t.Parallel()

	if remote.HTTPTimeout(remote.New("http://127.0.0.1:9200", "")) != 60*time.Second {
		t.Fatal("remote timeout")
	}
}

func TestPingOK(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/v1/health" {
			t.Fatalf("path %s", req.URL.Path)
		}

		if req.Header.Get("Authorization") != "Bearer "+tokenTok {
			t.Fatalf("auth %q", req.Header.Get("Authorization"))
		}

		encodeHealth(t, writer, true, true)
	}))
	t.Cleanup(server.Close)

	if err := remote.New(server.URL, tokenTok).Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestPingNotReady(t *testing.T) {
	t.Parallel()

	cases := []struct{ ok, ready bool }{
		{true, false},
		{false, true},
	}
	for _, testCase := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			encodeHealth(t, writer, testCase.ok, testCase.ready)
		}))
		err := remote.New(server.URL, "").Ping()
		server.Close()

		if err == nil || !strings.Contains(err.Error(), "未就绪") {
			t.Fatalf("%v: %v", testCase, err)
		}
	}
}

func TestPingErrors(t *testing.T) {
	t.Parallel()

	if err := remote.New("http://127.0.0.1:1", "").Ping(); err == nil {
		t.Fatal("expected connection error")
	}

	if err := remote.New("http://[", "").Ping(); err == nil {
		t.Fatal("expected parse error")
	}

	badJSON := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if _, err := writer.Write([]byte("not-json")); err != nil {
			t.Fatal(err)
		}
	}))
	t.Cleanup(badJSON.Close)

	if err := remote.New(badJSON.URL, "").Ping(); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestDoJSONAPIError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/whoami", func(writer http.ResponseWriter, _ *http.Request) {
		encodeAPIError(t, writer, http.StatusBadGateway, "session gone")
	})
	mux.HandleFunc("/api/v1/users/{id}", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)

		if _, err := writer.Write([]byte("not-json")); err != nil {
			t.Fatal(err)
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	forum := remote.New(server.URL, "")
	if _, err := forum.WhoAmI(); err == nil || !strings.Contains(err.Error(), "session gone") {
		t.Fatalf("api error: %v", err)
	}

	if _, err := forum.GetUser(1); err == nil || !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("status error: %v", err)
	}
}

func TestDoJSONReadClose(t *testing.T) {
	t.Parallel()

	forum := remote.New("http://example.com", "")
	remote.SetTransport(forum, roundTripFunc(func(*http.Request) (*http.Response, error) {
		resp := new(http.Response)
		resp.StatusCode = http.StatusOK
		resp.Body = io.NopCloser(errReader{})

		return resp, nil
	}))

	if err := forum.Ping(); err == nil || !strings.Contains(err.Error(), "读取响应") {
		t.Fatalf("%v", err)
	}

	remote.SetTransport(forum, roundTripFunc(func(*http.Request) (*http.Response, error) {
		resp := new(http.Response)
		resp.StatusCode = http.StatusOK
		resp.Body = closeBody{Reader: strings.NewReader(`{"ok":true,"ready":true}`), err: errClose}

		return resp, nil
	}))

	if err := forum.Ping(); err == nil || !strings.Contains(err.Error(), "关闭响应") {
		t.Fatalf("%v", err)
	}
}

func TestNewRequestError(t *testing.T) {
	t.Cleanup(remote.StubNewRequest(func(string, string, io.Reader) (*http.Request, error) {
		return nil, errBoom
	}))

	if err := remote.New("http://example.com", "").Ping(); err == nil || !strings.Contains(err.Error(), "构造请求") {
		t.Fatalf("%v", err)
	}
}

func TestFormatPostLocal(t *testing.T) {
	t.Parallel()

	detail := new(client.PostDetail)
	detail.Title = "t"

	if got := remote.New("http://127.0.0.1:9200", "").FormatPost(detail); !strings.Contains(got, "t") {
		t.Fatalf("%q", got)
	}
}

func TestPageQueryZero(t *testing.T) {
	t.Parallel()

	if remote.PageQuery(0) != nil {
		t.Fatal("expected nil")
	}
}
