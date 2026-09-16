package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/solanab/nsk/internal/client"
)

var (
	errInvalidPostID = errors.New("无效帖子 id")
	errInvalidFloor  = errors.New("无效楼层")
	errPageAndAll    = errors.New("--page 与 --all 不能同时使用")
	errFloorAndAll   = errors.New("id/floor 与 --all 不能同时使用")
	errFloorAndPage  = errors.New("id/floor 与 --page 不能同时使用")
)

type outputPath struct {
	set  bool
	path string
}

func (path *outputPath) Decode(ctx *kong.DecodeContext) error {
	path.set = true

	token := ctx.Scan.Peek()
	if token.IsValue() {
		path.path = ctx.Scan.Pop().String()
	}

	return nil
}

func (cmd *postCmd) Run(root *cliRoot, env *runEnv) error {
	postID, floor, hasFloor, err := parsePostRef(cmd.ID)
	if err != nil {
		return err
	}

	if err := cmd.checkFlags(hasFloor); err != nil {
		return err
	}

	acc, err := accountFrom(root)
	if err != nil {
		return err
	}

	return cmd.emitLoaded(root, env, acc, postID, floor, hasFloor)
}

func (cmd *postCmd) emitLoaded(
	root *cliRoot,
	env *runEnv,
	acc account,
	postID, floor int,
	hasFloor bool,
) error {
	detail, err := cmd.load(acc, postID, floor, hasFloor)
	if err != nil {
		return err
	}

	if detail == nil {
		return fmt.Errorf("%w", client.ErrNotFound)
	}

	return cmd.emit(root, env, detail)
}

func (cmd *postCmd) checkFlags(hasFloor bool) error {
	if cmd.All && cmd.Page != 0 {
		return errPageAndAll
	}

	if hasFloor && cmd.All {
		return errFloorAndAll
	}

	if hasFloor && cmd.Page != 0 {
		return errFloorAndPage
	}

	return nil
}

func parsePostRef(raw string) (int, int, bool, error) {
	idStr, floorStr, hasFloor := strings.Cut(raw, "/")

	postID, err := strconv.Atoi(idStr)
	if err != nil || postID <= 0 {
		return 0, 0, false, errInvalidPostID
	}

	if !hasFloor {
		return postID, 0, false, nil
	}

	floor, err := strconv.Atoi(floorStr)
	if err != nil || floor < 0 {
		return 0, 0, false, errInvalidFloor
	}

	return postID, floor, true, nil
}

func (cmd *postCmd) load(acc account, postID, floor int, hasFloor bool) (*client.PostDetail, error) {
	if cmd.All {
		detail, err := acc.GetPostAll(postID)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return detail, nil
	}

	page := cmd.Page
	if hasFloor {
		page = client.PageForFloor(floor)
	}

	detail, err := acc.GetPost(postID, page)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return detail, nil
}

func (cmd *postCmd) emit(root *cliRoot, env *runEnv, detail *client.PostDetail) error {
	markdown := client.FormatPost(detail)

	if !cmd.Output.set {
		if root.Text {
			return writeText(env.stdout, markdown)
		}

		return writeJSON(env.stdout, detail)
	}

	path := cmd.Output.path
	if path == "" {
		path = "post-" + strconv.Itoa(detail.ID) + ".md"
	}

	return writeSaved(root, env, path, markdown, detail)
}

func writeSaved(root *cliRoot, env *runEnv, path, markdown string, detail *client.PostDetail) error {
	if err := prepareSaved(root, path, markdown, detail); err != nil {
		return err
	}

	if root.Text {
		line := "saved " + path + " id=" + strconv.Itoa(detail.ID) + " title=" + detail.Title

		return writeText(env.stdout, line)
	}

	return writeJSON(env.stdout, client.SavedView{Saved: path, ID: detail.ID, Title: detail.Title})
}

func prepareSaved(root *cliRoot, path, markdown string, detail *client.PostDetail) error {
	if !root.Text {
		if _, err := loadSeam(&marshalJSON)(client.SavedView{
			Saved: path, ID: detail.ID, Title: detail.Title,
		}); err != nil {
			return fmt.Errorf("编码 JSON: %w", err)
		}
	}

	return writeMarkdown(path, markdown)
}

func writeMarkdown(path, markdown string) error {
	if err := loadSeam(&writeFile)(path, []byte(markdown), markdownFileMode); err != nil {
		return fmt.Errorf("写入 Markdown: %w", err)
	}

	return nil
}
