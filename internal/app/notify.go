package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/solanab/nsk/internal/client"
)

func (*notifyCmd) Run(root *cliRoot, env *runEnv) error {
	acc, err := accountFrom(root)
	if err != nil {
		return err
	}

	notes, err := acc.Notifications()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if notes == nil {
		notes = []client.Notification{}
	}

	if root.Text {
		return writeText(env.stdout, formatNotify(notes))
	}

	return writeJSON(env.stdout, notes)
}

func formatNotify(notes []client.Notification) string {
	lines := make([]string, len(notes))
	for i, note := range notes {
		lines[i] = formatNotifyLine(note)
	}

	return strings.Join(lines, "\n")
}

func formatNotifyLine(note client.Notification) string {
	line := "post_id=" + strconv.Itoa(note.PostID)
	if note.Page != 0 {
		line += " page=" + strconv.Itoa(note.Page)
	}

	if note.Floor != 0 {
		line += " floor=" + strconv.Itoa(note.Floor)
	}

	line += " url=" + note.URL
	if note.Text != "" {
		line += " text=" + note.Text
	}

	return line
}
