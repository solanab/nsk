package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/solanab/nsk/internal/client"
)

func (cmd *listCmd) Run(root *cliRoot, env *runEnv) error {
	acc, err := accountFrom(root)
	if err != nil {
		return err
	}

	list, err := cmd.load(acc)
	if err != nil {
		return err
	}

	if list == nil {
		return fmt.Errorf("%w", client.ErrNotFound)
	}

	if root.Text {
		return writeText(env.stdout, formatPostList(list))
	}

	return writeJSON(env.stdout, list)
}

func (cmd *listCmd) load(acc account) (*client.PostList, error) {
	if cmd.Slug == "" {
		list, err := acc.LatestPosts(cmd.Page)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return list, nil
	}

	list, err := acc.CategoryPosts(cmd.Slug, cmd.Page)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return list, nil
}

func formatPostList(list *client.PostList) string {
	lines := make([]string, len(list.Posts))
	for i, post := range list.Posts {
		lines[i] = "id=" + strconv.Itoa(post.ID) +
			" title=" + post.Title +
			" author=" + post.Author +
			" category=" + post.Category
	}

	return strings.Join(lines, "\n")
}
