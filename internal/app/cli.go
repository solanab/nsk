package app

import (
	"io"

	"github.com/alecthomas/kong"
)

type cliRoot struct {
	Version kong.VersionFlag `help:"Print version and exit."`
	Config  string           `help:"Path to config.toml."                   type:"path"`
	Text    bool             `help:"Human-readable output instead of JSON."`

	Structure structureCmd `cmd:"" default:"1"                            help:"Print the forum structure map."`
	Cats      catsCmd      `cmd:"" help:"Print all categories."`
	List      listCmd      `cmd:"" help:"List latest posts or a board."`
	Post      postCmd      `cmd:"" help:"Read a post page or all floors."`
	Search    searchCmd    `cmd:"" help:"Search posts."`
	Whoami    whoamiCmd    `cmd:"" help:"Print the current Account."`
	User      userCmd      `cmd:"" help:"Print a user by numeric id."`
	Notify    notifyCmd    `cmd:"" help:"Print at-me notifications."`
	Cookie    cookieCmd    `cmd:"" help:"Dump this process cookie jar."`
	Server    serverCmd    `cmd:"" help:"Hold Account and serve /api/v1."`
}

type structureCmd struct{}

type catsCmd struct{}

type listCmd struct {
	Slug string `arg:""                                    help:"Board slug. Omit for latest posts." optional:""`
	Page int    `help:"List page. Missing or 0 is page 1."`
}

type postCmd struct {
	ID     string     `arg:""                                      help:"Post id, or id/floor."`
	Page   int        `help:"Post page. Missing or 0 is page 1."`
	All    bool       `help:"Fetch all floors up to MaxPostPages."`
	Output outputPath `help:"Write Markdown."                      optional:""                  short:"o"`
}

type searchCmd struct {
	Query string `arg:""                                      help:"Search query." name:"q"`
	Page  int    `help:"Search page. Missing or 0 is page 1."`
}

type whoamiCmd struct{}

type userCmd struct {
	ID string `arg:"" help:"Numeric user id."`
}

type notifyCmd struct{}

type cookieCmd struct {
	Only bool `help:"Print pjwt= only."`
}

type serverCmd struct {
	Addr  string `help:"Listen host:port. Empty uses config or 127.0.0.1:9200."`
	Token string `help:"Bearer token. Required off loopback."`
}

type runEnv struct {
	stdout io.Writer
	stderr io.Writer
}
