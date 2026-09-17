package app

import (
	"fmt"
	"strings"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/config"
	"github.com/solanab/nsk/internal/server"
)

var serveHTTP = serveListen //nolint:gochecknoglobals // ListenAndServe test seam

func serveListen(srv *server.Server) error {
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (cmd *serverCmd) Run(root *cliRoot, env *runEnv) error {
	cfg, err := config.LoadPath(root.Config)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if cfg.HasClient() {
		return errServerOnClient
	}

	forum, err := loadSeam(&openForum)(cfg.CookieFile())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	srv, err := newLocalServer(cfg, cmd, forum)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(env.stderr, "nsk server listening on http://%s\n", srv.Addr())
	if err != nil {
		return fmt.Errorf("写入 stderr: %w", err)
	}

	return loadSeam(&serveHTTP)(srv)
}

func newLocalServer(cfg *config.Config, cmd *serverCmd, forum client.Forum) (*server.Server, error) {
	addr := strings.TrimSpace(cmd.Addr)
	if addr == "" {
		addr = cfg.ListenAddr()
	}

	token := cmd.Token
	if token == "" {
		token = cfg.Server.Token
	}

	srv, err := server.New(server.Config{Addr: addr, Token: token, Forum: forum})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return srv, nil
}
