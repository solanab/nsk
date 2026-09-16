package app

import (
	"fmt"
	"strconv"

	"github.com/solanab/nsk/internal/client"
)

func (*whoamiCmd) Run(root *cliRoot, env *runEnv) error {
	acc, err := accountFrom(root)
	if err != nil {
		return err
	}

	info, err := acc.WhoAmI()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if info == nil {
		return fmt.Errorf("%w", client.ErrExpiredCookie)
	}

	if root.Text {
		return writeText(env.stdout, formatUser(info))
	}

	return writeJSON(env.stdout, info)
}

func formatUser(info *client.UserInfo) string {
	out := "id=" + strconv.Itoa(info.ID) + " name=" + info.Name
	if info.Chicken != 0 {
		out += " chicken=" + strconv.Itoa(info.Chicken)
	}

	if info.Level != "" {
		out += " level=" + info.Level
	}

	return out
}
