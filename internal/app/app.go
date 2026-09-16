// Package app implements the nsk command surface.
package app

import (
	"io"

	"github.com/alecthomas/kong"
)

const (
	exitFailure = 1
	exitUsage   = 2
)

var version = "0.0.0-dev"

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
	env := &runEnv{stdout: stdout, stderr: stderr}

	parser, err := ctor(&root, kongOptions(stdout, stderr, env, &code, help)...)
	if err != nil {
		return writeErr(stderr, err, exitFailure)
	}

	switch {
	case len(args) == 0:
		args = []string{"structure"}
	case isHelpArg(args[0]):
		args = []string{"--help"}
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		if code != 0 {
			return code
		}

		if helpOrVersion(args) {
			return 0
		}

		return writeErr(stderr, err, exitUsage)
	}

	if helpOrVersion(args) {
		return code
	}

	return runContext(kctx, code, stderr)
}

func kongOptions(stdout, stderr io.Writer, env *runEnv, code *int, help kong.HelpOptions) []kong.Option {
	return []kong.Option{
		kong.Name("nsk"),
		kong.Description("NodeSeek agent CLI. One-shot commands; stdout is slim JSON."),
		kong.Writers(stdout, stderr),
		kong.Vars{"version": "nsk " + version},
		kong.Exit(func(value int) { *code = value }),
		kong.ConfigureHelp(help),
		kong.Bind(env),
	}
}

func runContext(kctx *kong.Context, code int, stderr io.Writer) int {
	if err := kctx.Run(); err != nil {
		return writeErr(stderr, err, exitFailure)
	}

	return code
}

func isHelpArg(value string) bool {
	return value == "help" || value == "--help" || value == "-h"
}

func helpOrVersion(args []string) bool {
	for _, arg := range args {
		if isHelpArg(arg) || arg == "--version" {
			return true
		}
	}

	return false
}
