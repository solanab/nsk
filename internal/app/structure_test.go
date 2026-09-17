package app_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/app"
	"github.com/solanab/nsk/internal/client"
)

const (
	cmdStructure = "structure"
	cmdCats      = "cats"
)

func implementedCommandNames() []string {
	return app.ImplementedCommandNames()
}

func forbiddenCommandNames() []string {
	return []string{"reply", "server"}
}

func TestHelpCommandsMatchImplemented(t *testing.T) {
	t.Parallel()

	code, stdout, stderr := runCmd(t, []string{flagHelp})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	got := helpCommandNames(t, stdout)
	if diff := cmp.Diff(implementedCommandNames(), got); diff != "" {
		t.Fatal(diff)
	}

	for _, name := range got {
		for _, bad := range forbiddenCommandNames() {
			if name == bad {
				t.Fatalf("help listed unimplemented %q", name)
			}
		}
	}
}

func TestEmptyArgvMatchesStructure(t *testing.T) {
	isolateXDG(t)

	codeEmpty, outEmpty, errEmpty := runCmd(t, nil)
	codeSlice, outSlice, errSlice := runCmd(t, []string{})
	codeNamed, outNamed, errNamed := runCmd(t, []string{cmdStructure})

	if codeEmpty != 0 || codeSlice != 0 || codeNamed != 0 {
		t.Fatalf("codes empty=%d slice=%d named=%d stderr=%q %q %q",
			codeEmpty, codeSlice, codeNamed, errEmpty, errSlice, errNamed)
	}

	if outEmpty != outNamed || outSlice != outNamed {
		t.Fatalf("empty=%q slice=%q named=%q", outEmpty, outSlice, outNamed)
	}

	got := decodeStructure(t, outNamed)
	assertSiteStructure(t, got)

	if errEmpty != "" || errSlice != "" || errNamed != "" {
		t.Fatalf("stderr empty=%q slice=%q named=%q", errEmpty, errSlice, errNamed)
	}
}

func TestCatsJSON(t *testing.T) {
	isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{cmdCats})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	assertCatsJSON(t, stdout)

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestStructureText(t *testing.T) {
	isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdStructure})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	assertStructureText(t, stdout)

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestTextWithoutCommandMatchesStructure(t *testing.T) {
	isolateXDG(t)

	codeBare, outBare, errBare := runCmd(t, []string{flagText})

	codeNamed, outNamed, errNamed := runCmd(t, []string{flagText, cmdStructure})
	if codeBare != 0 || codeNamed != 0 {
		t.Fatalf("codes bare=%d named=%d stderr=%q %q", codeBare, codeNamed, errBare, errNamed)
	}

	if outBare != outNamed {
		t.Fatalf("bare=%q named=%q", outBare, outNamed)
	}
}

func TestCatsText(t *testing.T) {
	isolateXDG(t)

	code, stdout, stderr := runCmd(t, []string{flagText, cmdCats})
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}

	if stdout != wantCatsText() {
		t.Fatalf("stdout=%q", stdout)
	}

	if stderr != "" {
		t.Fatalf("stderr=%q", stderr)
	}
}

func decodeStructure(t *testing.T, stdout string) client.SiteStructure {
	t.Helper()

	var got client.SiteStructure
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json %v stdout=%q", err, stdout)
	}

	return got
}

func assertSiteStructure(t *testing.T, got client.SiteStructure) {
	t.Helper()
	assertSiteMeta(t, got)
	assertStructureCommands(t, got.Commands)
}

func assertSiteMeta(t *testing.T, got client.SiteStructure) {
	t.Helper()

	if got.Site != "https://www.nodeseek.com" {
		t.Fatalf("site=%q", got.Site)
	}

	wantCats, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCats, got.Categories); diff != "" {
		t.Fatal(diff)
	}

	wantPages := client.Pagination{
		ListPerPage: client.ListPerPage, FloorsPerPage: client.FloorsPerPage, MaxPostPages: 50,
	}
	if diff := cmp.Diff(wantPages, got.Pagination); diff != "" {
		t.Fatal(diff)
	}

	wantPaths := client.SitePaths{
		Home:     "/page-{n}",
		Category: "/categories/{slug}?page={n}",
		Post:     "/post-{id}-{page}",
		Floor:    "/post-{id}-{page}#{floor}",
		User:     "/space/{id}",
	}
	if diff := cmp.Diff(wantPaths, got.Paths); diff != "" {
		t.Fatal(diff)
	}
}

func assertStructureCommands(t *testing.T, specs []client.CommandSpec) {
	t.Helper()

	names := commandNames(specs)
	if diff := cmp.Diff(implementedCommandNames(), names); diff != "" {
		t.Fatal(diff)
	}

	for _, name := range names {
		for _, bad := range forbiddenCommandNames() {
			if name == bad {
				t.Fatalf("commands listed unimplemented %q", name)
			}
		}
	}

	for _, spec := range specs {
		if spec.Usage == "" || spec.Stdout == "" {
			t.Fatalf("empty command spec %#v", spec)
		}
	}
}

func assertCatsJSON(t *testing.T, stdout string) {
	t.Helper()

	var got []client.Category
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json %v stdout=%q", err, stdout)
	}

	want, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func assertStructureText(t *testing.T, stdout string) {
	t.Helper()

	if strings.HasPrefix(stdout, "{") {
		t.Fatalf("JSON on --text: %q", stdout)
	}

	for _, line := range []string{
		"site=https://www.nodeseek.com",
		"list_per_page=49",
		"floors_per_page=11",
		"max_post_pages=50",
		"home=/page-{n}",
		"category=/categories/{slug}?page={n}",
		"post=/post-{id}-{page}",
		"floor=/post-{id}-{page}#{floor}",
		"user=/space/{id}",
		"daily 日常",
		"sandbox 沙盒",
		"structure usage=nsk [structure] stdout=SiteStructure",
		"cats usage=nsk cats stdout=[]Category",
		"list usage=nsk list [slug] [--page N] stdout=PostList",
		"post usage=nsk post <id> [--page N|--all] [-o file] stdout=PostDetail / SavedView",
		"search usage=nsk search <q> [--page N] stdout=SearchResult",
		"whoami usage=nsk whoami stdout=UserInfo",
		"user usage=nsk user <id> stdout=UserInfo",
		"notify usage=nsk notify stdout=[]Notification",
		"cookie usage=nsk cookie [--only] stdout=JSON / pjwt=",
	} {
		if !strings.Contains(stdout, line) {
			t.Fatalf("missing %q in %q", line, stdout)
		}
	}
}

func helpCommandNames(t *testing.T, help string) []string {
	t.Helper()

	const header = "\nCommands:\n"

	_, section, found := strings.Cut(help, header)
	if !found {
		t.Fatalf("no Commands section: %q", help)
	}

	var names []string

	for line := range strings.SplitSeq(section, "\n") {
		if strings.TrimSpace(line) == "" {
			break
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		names = append(names, fields[0])
	}

	if len(names) == 0 {
		t.Fatalf("no commands in help: %q", help)
	}

	return names
}

func commandNames(specs []client.CommandSpec) []string {
	names := make([]string, len(specs))
	for i, spec := range specs {
		names[i] = spec.Name
	}

	return names
}

func wantCatsText() string {
	return "daily 日常\n" +
		"tech 技术\n" +
		"info 情报\n" +
		"review 测评\n" +
		"trade 交易\n" +
		"carpool 拼车\n" +
		"promotion 推广\n" +
		"life 生活\n" +
		"dev Dev\n" +
		"photo-share 贴图\n" +
		"expose 曝光\n" +
		"inside 内版\n" +
		"meaningless 无意义\n" +
		"sandbox 沙盒\n"
}
