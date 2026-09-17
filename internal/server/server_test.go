package server_test

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/server"
)

var errWrite = errors.New("write")

func TestNewRejects(t *testing.T) {
	t.Parallel()

	if _, err := server.New(server.Config{}); err == nil {
		t.Fatal("expected missing forum")
	}

	if _, err := server.New(server.Config{Forum: newStub(), Addr: "0.0.0.0:9200"}); err == nil {
		t.Fatal("expected unspecified bind")
	}
}

func TestNewDefaultAddr(t *testing.T) {
	t.Parallel()

	srv := mustServer(t, server.Config{Forum: newStub()})
	if srv.Addr() != "127.0.0.1:9200" {
		t.Fatalf("%s", srv.Addr())
	}
}

func TestHealthNoAuth(t *testing.T) {
	t.Parallel()

	srv := mustServer(t, server.Config{Forum: newStub(), Token: tokenSecret})

	rec := doGET(t, srv.Handler(), server.PathHealth, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("%d", rec.Code)
	}

	var got struct {
		OK    bool `json:"ok"`
		Ready bool `json:"ready"`
	}
	decodeJSON(t, rec, &got)

	if !got.OK || !got.Ready {
		t.Fatalf("%#v", got)
	}
}

func TestAuthBearer(t *testing.T) {
	t.Parallel()

	handler := mustServer(t, server.Config{Forum: newStub(), Token: tokenSecret}).Handler()

	denied := doGET(t, handler, server.PathWhoAmI, "")
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("missing token %d", denied.Code)
	}

	if errorMsg(t, denied) != "unauthorized" {
		t.Fatal("unauthorized body")
	}

	wrong := doGET(t, handler, server.PathWhoAmI, "nope")
	if wrong.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token %d", wrong.Code)
	}

	okRec := doGET(t, handler, server.PathWhoAmI, tokenSecret)
	if okRec.Code != http.StatusOK {
		t.Fatalf("ok %d", okRec.Code)
	}

	var info client.UserInfo
	decodeJSON(t, okRec, &info)

	if info.Name != nameAlice {
		t.Fatalf("%#v", info)
	}
}

func TestLoopbackSkipsAuth(t *testing.T) {
	t.Parallel()

	rec := doGET(t, mustServer(t, server.Config{Forum: newStub()}).Handler(), server.PathWhoAmI, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("%d", rec.Code)
	}
}

func TestCloseNilHTTP(t *testing.T) {
	t.Parallel()

	if err := mustServer(t, server.Config{Forum: newStub()}).Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCloseHTTPError(t *testing.T) {
	t.Cleanup(server.StubListenHTTP(func(*http.Server) error {
		return nil
	}))
	t.Cleanup(server.StubCloseHTTP(func(*http.Server) error {
		return errUpstream
	}))

	srv := mustServer(t, server.Config{Forum: newStub()})
	if err := srv.ListenAndServe(); err != nil {
		t.Fatal(err)
	}

	if err := srv.Close(); err == nil {
		t.Fatal("expected close error")
	}
}

func TestListenAndServeError(t *testing.T) {
	t.Cleanup(server.StubListenHTTP(func(*http.Server) error {
		return errUpstream
	}))

	if err := mustServer(t, server.Config{Forum: newStub()}).ListenAndServe(); err == nil {
		t.Fatal("expected listen error")
	}
}

func TestListenAndServeOK(t *testing.T) {
	t.Cleanup(server.StubListenHTTP(func(*http.Server) error {
		return nil
	}))

	if err := mustServer(t, server.Config{Forum: newStub()}).ListenAndServe(); err != nil {
		t.Fatal(err)
	}
}

func TestListenAndServeHealth(t *testing.T) {
	cfg := new(net.ListenConfig)

	listener, err := cfg.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	srv, err := server.New(server.Config{Forum: newStub(), Addr: addr})
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- srv.ListenAndServe()
	}()

	if err := waitHealth(t, "http://"+srv.Addr()+server.PathHealth); err != nil {
		t.Fatal(err)
	}

	if server.HeaderTimeout(srv) != 5*time.Second {
		t.Fatal("ReadHeaderTimeout")
	}

	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}

	<-done
}

func TestWriteJSONEncodeError(t *testing.T) {
	t.Parallel()

	rec := httptestRecorder()
	failing := server.FailWriter{ResponseWriter: rec, Err: errWrite}
	server.WriteJSONForTest(failing, http.StatusOK, map[string]bool{"ok": true})
}

func waitHealth(t *testing.T, rawURL string) error {
	t.Helper()

	var last error

	for range 50 {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, rawURL, nil)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			last = err

			time.Sleep(10 * time.Millisecond)

			continue
		}

		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("%w", err)
		}

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		last = errUpstream
	}

	return last
}

func httptestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}
