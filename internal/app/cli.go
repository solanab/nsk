package app

import (
	"io"

	"github.com/alecthomas/kong"
)

type cliRoot struct {
	Version kong.VersionFlag `help:"Print version and exit."`
	Config  string           `help:"Path to config.toml."                   type:"path"`
	Text    bool             `help:"Human-readable output instead of JSON."`

	Structure structureCmd `cmd:"" default:"1"                          help:"Print the forum structure map."`
	Cats      catsCmd      `cmd:"" help:"Print all categories."`
	Whoami    whoamiCmd    `cmd:"" help:"Print the current Account."`
	Cookie    cookieCmd    `cmd:"" help:"Dump this process cookie jar."`
}

type structureCmd struct{}

type catsCmd struct{}

type whoamiCmd struct{}

type cookieCmd struct {
	Only bool `help:"Print pjwt= only."`
}

type runEnv struct {
	stdout io.Writer
	stderr io.Writer
}
