package client_test

import (
	"strings"
	"testing"

	"github.com/solanab/nsk/internal/client"
)

func TestFormatPostNil(t *testing.T) {
	t.Parallel()

	if got := client.FormatPost(nil); got != "" {
		t.Fatalf("got=%q", got)
	}

	if got := client.Unstarted().FormatPost(nil); got != "" {
		t.Fatalf("method got=%q", got)
	}
}

func TestFormatPostRendersTitleAndFloors(t *testing.T) {
	t.Parallel()

	detail := new(client.PostDetail)
	detail.Title = "hello"
	detail.Floors = []client.Floor{
		{Number: 0, Author: "alice", Markdown: "op body"},
		{Number: 1, Author: "bob", Markdown: "reply"},
	}

	got := client.Unstarted().FormatPost(detail)
	want := "# hello\n\n## #0 alice\n\nop body\n\n## #1 bob\n\nreply"

	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestFormatPostStripsANSI(t *testing.T) {
	t.Parallel()

	detail := new(client.PostDetail)
	detail.Title = "t"
	detail.Floors = []client.Floor{
		{Number: 0, Author: "a", Markdown: "speed \x1b[31mred\x1b[0m done"},
	}

	got := client.FormatPost(detail)
	if strings.Contains(got, "\x1b") || strings.Contains(got, "[31m") {
		t.Fatalf("ansi left: %q", got)
	}

	if !strings.Contains(got, "speed red done") {
		t.Fatalf("got=%q", got)
	}
}

func TestFormatPostEmptyFloors(t *testing.T) {
	t.Parallel()

	detail := new(client.PostDetail)
	detail.Title = "only title"
	detail.Floors = []client.Floor{}

	if got := client.FormatPost(detail); got != "# only title" {
		t.Fatalf("got=%q", got)
	}

	if got := client.FormatPost(new(client.PostDetail)); got != "" {
		t.Fatalf("empty=%q", got)
	}

	detail = new(client.PostDetail)
	detail.Floors = []client.Floor{{Number: 3, Markdown: "solo"}}

	if got := client.FormatPost(detail); got != "## #3\n\nsolo" {
		t.Fatalf("got=%q", got)
	}
}
