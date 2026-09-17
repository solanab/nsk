package app

import (
	"io"
	"os"

	"github.com/alecthomas/kong"

	"github.com/solanab/nsk/internal/client"
	"github.com/solanab/nsk/internal/server"
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

// StubWriteFile replaces markdown -o writes.
func StubWriteFile(fn func(string, []byte, os.FileMode) error) func() {
	return swapSeam(&writeFile, fn)
}

// StubLoadCategories replaces the local Categories table loader.
func StubLoadCategories(fn func() ([]client.Category, error)) func() {
	return swapSeam(&loadCategories, fn)
}

// WrapAccount is the Forum constructor error-wrap test seam.
//
//nolint:ireturn // test seam for account interface
func WrapAccount(forum *client.Client, err error) (Account, error) {
	return wrapAccount(forum, err)
}

// StubOpenRemote replaces remote Forum construction.
func StubOpenRemote(fn func(string, string) (Account, error)) func() {
	return swapSeam(&openRemote, fn)
}

// StubOpenForum replaces local Forum construction for nsk server.
func StubOpenForum(fn func(string) (client.Forum, error)) func() {
	return swapSeam(&openForum, fn)
}

// StubServeHTTP replaces Server.ListenAndServe.
func StubServeHTTP(fn func(*server.Server) error) func() {
	return swapSeam(&serveHTTP, fn)
}

// PingRemote is the [client] Ping+wrap test seam.
//
//nolint:ireturn // test seam for account interface
func PingRemote(baseURL, token string) (Account, error) {
	return pingRemote(baseURL, token)
}

// OpenLocalForum is the nsk server Forum opener test seam.
//
//nolint:ireturn // test seam for Forum
func OpenLocalForum(cookieFile string) (client.Forum, error) {
	return openLocalForum(cookieFile)
}

// WrapForum is the Forum constructor error-wrap test seam.
//
//nolint:ireturn // test seam for Forum
func WrapForum(forum *client.Client, err error) (client.Forum, error) {
	return wrapForum(forum, err)
}

// ServeListen is the ListenAndServe wrapper test seam.
func ServeListen(srv *server.Server) error {
	return serveListen(srv)
}

// ImplementedCommandNames is the SiteStructure.commands name list.
func ImplementedCommandNames() []string {
	specs := implementedCommands()

	names := make([]string, len(specs))
	for i, spec := range specs {
		names[i] = spec.Name
	}

	return names
}
