// Package server exposes a local Forum over HTTP. Cookie never leaves the process.
package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
)

var (
	listenHTTP = (*http.Server).ListenAndServe //nolint:gochecknoglobals // ListenAndServe test seam
	closeHTTP  = (*http.Server).Close          //nolint:gochecknoglobals // Close test seam
)

const (
	readHeaderTimeout = 5 * time.Second
	pathHealth        = "/api/v1/health"
	pathWhoAmI        = "/api/v1/whoami"
	pathMe            = "/api/v1/me"
	pathUser          = "/api/v1/users/{id}"
	pathPosts         = "/api/v1/posts"
	pathPost          = "/api/v1/posts/{id}"
	pathSearch        = "/api/v1/search"
	pathCategories    = "/api/v1/categories"
	pathNotifications = "/api/v1/notifications"
)

// Config is nsk server listen settings. Cookie stays in Forum.
type Config struct {
	Addr  string
	Token string
	Forum client.Forum
}

// Server exposes the local Forum over HTTP. Cookie never leaves this process.
type Server struct {
	addr  string
	token string
	forum client.Forum
	mu    sync.Mutex
	http  *http.Server
}

// New validates the listen address. Non-loopback requires a token.
func New(cfg Config) (*Server, error) {
	if cfg.Forum == nil {
		return nil, errMissingForum
	}

	addr, err := config.CanonicalListen(cfg.Addr, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	srv := new(Server)
	srv.addr = addr
	srv.token = cfg.Token
	srv.forum = cfg.Forum

	return srv, nil
}

// Addr returns the validated host:port.
func (srv *Server) Addr() string {
	return srv.addr
}

// Handler returns the authenticated mux for tests and ListenAndServe.
func (srv *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	srv.register(mux)

	return srv.withAuth(mux)
}

// ListenAndServe blocks on the configured address.
func (srv *Server) ListenAndServe() error {
	httpSrv := new(http.Server)
	httpSrv.Addr = srv.addr
	httpSrv.Handler = srv.Handler()
	httpSrv.ReadHeaderTimeout = readHeaderTimeout
	srv.http = httpSrv

	if err := currentListen()(httpSrv); err != nil {
		return fmt.Errorf("监听: %w", err)
	}

	return nil
}

// Close stops ListenAndServe. It is safe before the server starts.
func (srv *Server) Close() error {
	if srv.http == nil {
		return nil
	}

	if err := currentClose()(srv.http); err != nil {
		return fmt.Errorf("关闭: %w", err)
	}

	return nil
}

func (srv *Server) register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+pathHealth, srv.handleHealth)
	mux.HandleFunc("GET "+pathWhoAmI, srv.handleWhoAmI)
	mux.HandleFunc("GET "+pathMe, srv.handleWhoAmI)
	mux.HandleFunc("GET "+pathUser, srv.handleUser)
	mux.HandleFunc("GET "+pathPosts, srv.handlePosts)
	mux.HandleFunc("GET "+pathPost, srv.handlePost)
	mux.HandleFunc("GET "+pathSearch, srv.handleSearch)
	mux.HandleFunc("GET "+pathCategories, srv.handleCategories)
	mux.HandleFunc("GET "+pathNotifications, srv.handleNotifications)
}

func (srv *Server) withForum(load func(client.Forum) error) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if err := load(srv.forum); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
