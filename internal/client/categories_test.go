package client_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/client"
)

func TestCategoriesTable(t *testing.T) {
	t.Parallel()

	got, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	want := []client.Category{
		{Slug: slugDaily, Name: "日常"},
		{Slug: slugTech, Name: "技术"},
		{Slug: "info", Name: "情报"},
		{Slug: "review", Name: "测评"},
		{Slug: "trade", Name: "交易"},
		{Slug: "carpool", Name: "拼车"},
		{Slug: "promotion", Name: "推广"},
		{Slug: "life", Name: "生活"},
		{Slug: "dev", Name: "Dev"},
		{Slug: "photo-share", Name: "贴图"},
		{Slug: "expose", Name: "曝光"},
		{Slug: "inside", Name: "内版"},
		{Slug: "meaningless", Name: "无意义"},
		{Slug: "sandbox", Name: "沙盒"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestClientCategories(t *testing.T) {
	t.Parallel()

	want, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.Unstarted().Categories()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestCategoriesCopy(t *testing.T) {
	t.Parallel()

	got, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	got[0].Slug = "mutated"

	again, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	if again[0].Slug != slugDaily {
		t.Fatalf("Categories shared backing array: %#v", again[0])
	}
}

func TestCategoriesMatchListFixtureSidebar(t *testing.T) {
	t.Parallel()

	html := string(fixture(t, htmlListPage))
	start := strings.Index(html, `<nav class="category-menu">`)

	end := strings.Index(html, `</nav>`)
	if start < 0 || end < 0 || end < start {
		t.Fatal("missing category nav")
	}

	nav := html[start:end]
	slugs := sidebarSlugs(nav)

	cats, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	got := make([]string, len(cats))
	for i, cat := range cats {
		got[i] = cat.Slug
	}

	if diff := cmp.Diff(slugs, got); diff != "" {
		t.Fatal(diff)
	}
}

func sidebarSlugs(nav string) []string {
	const marker = `href="/categories/`

	var slugs []string

	rest := nav
	for {
		idx := strings.Index(rest, marker)
		if idx < 0 {
			return slugs
		}

		rest = rest[idx+len(marker):]

		end := strings.Index(rest, `"`)
		if end < 0 {
			return slugs
		}

		slugs = append(slugs, rest[:end])
		rest = rest[end:]
	}
}
