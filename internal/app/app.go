// Package app implements the nsk command surface.
package app

import (
	"fmt"
	"io"

	"github.com/alecthomas/kong"
)

const (
	exitFailure = 1
	exitUsage   = 2
)

var version = "0.0.0-dev"

type cliRoot struct {
	Version kong.VersionFlag `help:"Print version and exit."`
	Config  string           `help:"Path to config.toml."                   type:"path"`
	Text    bool             `help:"Human-readable output instead of JSON."`
}

type kongNewFunc func(grammar any, options ...kong.Option) (*kong.Kong, error)

// Run executes nsk and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	return run(kong.New, args, stdout, stderr)
}

func run(ctor kongNewFunc, args []string, stdout, stderr io.Writer) int {
	var (
		code int
		root cliRoot
		help kong.HelpOptions
	)

	help.Compact = true

	parser, err := ctor(&root,
		kong.Name("nsk"),
		kong.Description("NodeSeek agent CLI. One-shot commands; stdout is slim JSON."),
		kong.Writers(stdout, stderr),
		kong.Vars{"version": "nsk " + version},
		kong.Exit(func(value int) { code = value }),
		kong.ConfigureHelp(help),
	)
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "nsk: %v\n", err); writeErr != nil {
			return exitFailure
		}

		return exitFailure
	}

	if len(args) == 0 || isHelpArg(args[0]) {
		args = []string{"--help"}
	}

	_, err = parser.Parse(args)
	if err != nil {
		if code != 0 {
			return code
		}

		if _, writeErr := fmt.Fprintf(stderr, "nsk: %v\n", err); writeErr != nil {
			return exitFailure
		}

		return exitUsage
	}

	return code
}

func isHelpArg(value string) bool {
	return value == "help" || value == "--help" || value == "-h"
}
