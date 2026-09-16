package app

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/solanab/nsk/internal/client"
)

var errInvalidUserID = errors.New("无效用户 id")

func (cmd *userCmd) Run(root *cliRoot, env *runEnv) error {
	userID, err := parseUserID(cmd.ID)
	if err != nil {
		return err
	}

	acc, err := accountFrom(root)
	if err != nil {
		return err
	}

	info, err := acc.GetUser(userID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if info == nil {
		return fmt.Errorf("%w", client.ErrNotFound)
	}

	if root.Text {
		return writeText(env.stdout, formatUser(info))
	}

	return writeJSON(env.stdout, info)
}

func parseUserID(raw string) (int, error) {
	userID, err := strconv.Atoi(raw)
	if err != nil || userID <= 0 {
		return 0, errInvalidUserID
	}

	return userID, nil
}
