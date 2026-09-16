package app

import (
	"io"

	"github.com/alecthomas/kong"

	"github.com/solanab/nsk/internal/client"
)

// RunWithKongNew is the test seam for Kong constructor and Exit error paths.
func RunWithKongNew(
	ctor func(grammar any, options ...kong.Option) (*kong.Kong, error),
	args []string,
	stdout, stderr io.Writer,
) int {
	return run(ctor, args, stdout, stderr)
}

// Account is the CLI Forum+cookie test seam.
type Account = account

// StubOpenAccount replaces local Client construction.
func StubOpenAccount(fn func(string) (Account, error)) func() {
	return swapSeam(&openAccount, fn)
}

// StubMarshalJSON replaces JSON encoding.
func StubMarshalJSON(fn func(any) ([]byte, error)) func() {
	return swapSeam(&marshalJSON, fn)
}

// WrapAccount is the Forum constructor error-wrap test seam.
//
//nolint:ireturn // test seam for account interface
func WrapAccount(forum *client.Client, err error) (Account, error) {
	return wrapAccount(forum, err)
}
