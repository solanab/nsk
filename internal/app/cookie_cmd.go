package app

import (
	"fmt"

	"github.com/solanab/nsk/internal/config"
)

func (cmd *cookieCmd) Run(root *cliRoot, env *runEnv) error {
	cfg, err := config.LoadPath(root.Config)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if cfg.HasClient() {
		return errCookieOnClient
	}

	acc, err := loadSeam(&openAccount)(cfg.CookieFile())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if cmd.Only {
		line, err := acc.ExportPJWT()
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return writeText(env.stdout, line)
	}

	payload, err := acc.ExportCookiesJSON()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return writeText(env.stdout, payload)
}
