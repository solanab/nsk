// Command nsk is the NodeSeek agent CLI.
package main

import (
	"os"

	"github.com/solanab/nsk/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
