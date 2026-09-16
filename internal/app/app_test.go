package app_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/alecthomas/kong"

	"github.com/solanab/nsk/internal/app"
)

var (
	errKongGrammar = errors.New("kong grammar")
	errKongForced  = errors.New("kong forced exit")
)

const unknownCommand = "serve"

type failWriter struct {
	remainingWrites int
}

func (writer *failWriter) Write(data []byte) (int, error) {
	if writer.remainingWrites == 0 {
		return 0, io.ErrClosedPipe
	}

	writer.remainingWrites--

	return len(data), nil
}

func TestRunHelp(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{"help"}, {flagHelp}, {"-h"}} {
		var stdout, stderr bytes.Buffer
		if code := app.Run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("Run(%v) returned %d, want 0; stderr=%q stdout=%q", args, code, stderr.String(), stdout.String())
		}

		got := stdout.String()
		if !strings.Contains(got, "nsk") {
			t.Fatalf("Run(%v) stdout = %q", args, got)
		}

		if strings.Contains(got, "{") || strings.Contains(got, `"categories"`) {
			t.Fatalf("Run(%v) printed structure JSON: %q", args, got)
		}
	}
}

func TestRunVersion(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := app.Run([]string{flagVersion}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run returned %d, want 0; stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}

	got := stdout.String()
	if !strings.Contains(got, "nsk") || !strings.Contains(got, "0.0.0-dev") {
		t.Fatalf("stdout = %q", got)
	}

	if strings.Contains(got, "{") || strings.Contains(got, `"categories"`) || strings.Contains(got, "site=") {
		t.Fatalf("stdout leaked structure: %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunHelpVersionSkipCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args []string
		want string
	}{
		{args: []string{flagVersion}, want: "nsk 0.0.0-dev"},
		{args: []string{flagText, flagVersion}, want: "nsk 0.0.0-dev"},
		{args: []string{cmdStructure, flagHelp}, want: "Usage: nsk structure"},
		{args: []string{cmdCats, flagHelp}, want: "Usage: nsk cats"},
		{args: []string{cmdList, flagHelp}, want: "Usage: nsk list"},
	}
	for _, test := range tests {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			if code := app.Run(test.args, &stdout, &stderr); code != 0 {
				t.Fatalf("Run(%v) returned %d; stderr=%q stdout=%q",
					test.args, code, stderr.String(), stdout.String())
			}

			got := stdout.String()
			if !strings.Contains(got, test.want) {
				t.Fatalf("Run(%v) stdout = %q want %q", test.args, got, test.want)
			}

			if strings.Contains(got, "{") || strings.Contains(got, `"categories"`) ||
				strings.Contains(got, "site=") {
				t.Fatalf("Run(%v) leaked command output: %q", test.args, got)
			}

			if stderr.Len() != 0 {
				t.Fatalf("Run(%v) stderr = %q", test.args, stderr.String())
			}
		})
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := app.Run([]string{unknownCommand}, &stdout, &stderr); code != 2 {
		t.Fatalf("Run returned %d, want 2; stderr=%q", code, stderr.String())
	}

	if stderr.Len() == 0 {
		t.Fatal("expected stderr")
	}
}

func TestRunWriteErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		stdout io.Writer
		stderr io.Writer
	}{
		{
			name:   "unknown command",
			args:   []string{unknownCommand},
			stdout: io.Discard,
			stderr: &failWriter{remainingWrites: 0},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if code := app.Run(test.args, test.stdout, test.stderr); code != 1 {
				t.Fatalf("Run returned %d, want 1", code)
			}
		})
	}
}

func TestRunKongNewError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	if code := app.RunWithKongNew(failKongNew, nil, io.Discard, &stderr); code != 1 {
		t.Fatalf("Run returned %d, want 1; stderr=%q", code, stderr.String())
	}

	if !strings.Contains(stderr.String(), "nsk:") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunKongNewWriteError(t *testing.T) {
	t.Parallel()

	if code := app.RunWithKongNew(failKongNew, nil, io.Discard, &failWriter{remainingWrites: 0}); code != 1 {
		t.Fatalf("Run returned %d, want 1", code)
	}
}

func TestRunKongExitError(t *testing.T) {
	t.Parallel()

	ctor := func(grammar any, options ...kong.Option) (*kong.Kong, error) {
		options = append(options, kong.WithBeforeApply(func(parser *kong.Kong) error {
			parser.Exit(2)

			return errKongForced
		}))

		return kong.New(grammar, options...)
	}
	if code := app.RunWithKongNew(ctor, []string{cmdWhoami}, io.Discard, io.Discard); code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
}

func failKongNew(any, ...kong.Option) (*kong.Kong, error) {
	return nil, errKongGrammar
}
