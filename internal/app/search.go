package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/solanab/nsk/internal/client"
)

var errEmptyQuery = errors.New("搜索词为空")

func (cmd *searchCmd) Run(root *cliRoot, env *runEnv) error {
	query := strings.TrimSpace(cmd.Query)
	if query == "" {
		return errEmptyQuery
	}

	acc, err := accountFrom(root)
	if err != nil {
		return err
	}

	result, err := acc.Search(query, cmd.Page)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if result == nil {
		return fmt.Errorf("%w", client.ErrNotFound)
	}

	if root.Text {
		return writeText(env.stdout, formatSearch(result))
	}

	return writeJSON(env.stdout, result)
}

func formatSearch(result *client.SearchResult) string {
	header := "query=" + result.Query + " page=" + strconv.Itoa(result.Page)
	if len(result.Posts) == 0 {
		return header
	}

	lines := make([]string, 0, 1+len(result.Posts))
	lines = append(lines, header)

	for _, post := range result.Posts {
		lines = append(lines, "id="+strconv.Itoa(post.ID)+
			" title="+post.Title+
			" author="+post.Author)
	}

	return strings.Join(lines, "\n")
}
