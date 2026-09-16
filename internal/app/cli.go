package app

import (
	"io"

	"github.com/alecthomas/kong"
)

type cliRoot struct {
	Version kong.VersionFlag `help:"Print version and exit."`
	Config  string           `help:"Path to config.toml."                   type:"path"`
	Text    bool             `help:"Human-readable output instead of JSON."`

	Whoami whoamiCmd `cmd:"" help:"Print the current Account."`
	Cookie cookieCmd `cmd:"" help:"Dump this process cookie jar."`
}

type whoamiCmd struct{}

type cookieCmd struct {
	Only bool `help:"Print pjwt= only."`
}

type runEnv struct {
	stdout io.Writer
	stderr io.Writer
}
