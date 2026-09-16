package app

import (
	"io"

	"github.com/alecthomas/kong"
)

// RunWithKongNew is the test seam for Kong constructor and Exit error paths.
func RunWithKongNew(
	ctor func(grammar any, options ...kong.Option) (*kong.Kong, error),
	args []string,
	stdout, stderr io.Writer,
) int {
	return run(ctor, args, stdout, stderr)
}
