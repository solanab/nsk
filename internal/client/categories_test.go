package client_test

import (
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
		{Slug: "daily", Name: "日常"},
		{Slug: "tech", Name: "技术"},
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

	if again[0].Slug != "daily" {
		t.Fatalf("Categories shared backing array: %#v", again[0])
	}
}
