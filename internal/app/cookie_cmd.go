package app

import "fmt"

func (cmd *cookieCmd) Run(root *cliRoot, env *runEnv) error {
	acc, err := accountFrom(root)
	if err != nil {
		return err
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
