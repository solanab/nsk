package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/solanab/nsk/internal/client"
)

var loadCategories = client.Categories //nolint:gochecknoglobals // Categories() test seam

func (*structureCmd) Run(root *cliRoot, env *runEnv) error {
	structure, err := currentStructure()
	if err != nil {
		return err
	}

	if root.Text {
		return writeText(env.stdout, formatStructure(structure))
	}

	return writeJSON(env.stdout, structure)
}

func (*catsCmd) Run(root *cliRoot, env *runEnv) error {
	cats, err := loadSeam(&loadCategories)()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if root.Text {
		return writeText(env.stdout, formatCats(cats))
	}

	return writeJSON(env.stdout, cats)
}

func currentStructure() (*client.SiteStructure, error) {
	cats, err := loadSeam(&loadCategories)()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	structure := new(client.SiteStructure)
	structure.Site = client.Site
	structure.Categories = cats
	structure.Pagination.ListPerPage = client.ListPerPage
	structure.Pagination.FloorsPerPage = client.FloorsPerPage
	structure.Pagination.MaxPostPages = client.MaxPostPages
	structure.Paths.Home = client.PathHome
	structure.Paths.Category = client.PathCategory
	structure.Paths.Post = client.PathPost
	structure.Paths.Floor = client.PathFloor
	structure.Paths.User = client.PathUser
	structure.Commands = implementedCommands()

	return structure, nil
}

func implementedCommands() []client.CommandSpec {
	return []client.CommandSpec{
		{Name: "structure", Usage: "nsk [structure]", Stdout: "SiteStructure"},
		{Name: "cats", Usage: "nsk cats", Stdout: "[]Category"},
		{Name: "list", Usage: "nsk list [slug] [--page N]", Stdout: "PostList"},
		{Name: "post", Usage: "nsk post <id> [--page N|--all] [-o file]", Stdout: "PostDetail / SavedView"},
		{Name: "search", Usage: "nsk search <q> [--page N]", Stdout: "SearchResult"},
		{Name: "whoami", Usage: "nsk whoami", Stdout: "UserInfo"},
		{Name: "user", Usage: "nsk user <id>", Stdout: "UserInfo"},
		{Name: "notify", Usage: "nsk notify", Stdout: "[]Notification"},
		{Name: "cookie", Usage: "nsk cookie [--only]", Stdout: "JSON / pjwt="},
		{Name: "server", Usage: "nsk server [--addr HOST:PORT] [--token TOKEN]", Stdout: "stderr 日志"},
	}
}

func formatStructure(structure *client.SiteStructure) string {
	header := []string{
		"site=" + structure.Site,
		"list_per_page=" + strconv.Itoa(structure.Pagination.ListPerPage),
		"floors_per_page=" + strconv.Itoa(structure.Pagination.FloorsPerPage),
		"max_post_pages=" + strconv.Itoa(structure.Pagination.MaxPostPages),
		"home=" + structure.Paths.Home,
		"category=" + structure.Paths.Category,
		"post=" + structure.Paths.Post,
		"floor=" + structure.Paths.Floor,
		"user=" + structure.Paths.User,
	}
	lines := make([]string, 0, len(header)+len(structure.Categories)+len(structure.Commands))
	lines = append(lines, header...)
	lines = append(lines, catLines(structure.Categories)...)

	for _, spec := range structure.Commands {
		lines = append(lines, spec.Name+" usage="+spec.Usage+" stdout="+spec.Stdout)
	}

	return strings.Join(lines, "\n")
}

func formatCats(cats []client.Category) string {
	return strings.Join(catLines(cats), "\n")
}

func catLines(cats []client.Category) []string {
	lines := make([]string, len(cats))
	for i, cat := range cats {
		lines[i] = cat.Slug + " " + cat.Name
	}

	return lines
}
