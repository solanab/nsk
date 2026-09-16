package client_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/client"
)

const (
	jsonGetInfoOK       = "getInfo-ok.json"
	jsonGetInfoNoDetail = "getInfo-no-detail.json"
	jsonGetInfoEmpty    = "getInfo-empty.json"
)

func TestParseGetInfoOK(t *testing.T) {
	t.Parallel()

	got, err := client.ParseGetInfo(jsonFixture(t, jsonGetInfoOK))
	if err != nil {
		t.Fatal(err)
	}

	want := &client.UserInfo{ID: 42, Name: nameAlice, Chicken: 7, Level: levelTwo}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseGetInfoMissingDetail(t *testing.T) {
	t.Parallel()

	got, err := client.ParseGetInfo(jsonFixture(t, jsonGetInfoNoDetail))
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}

	if got != nil {
		t.Fatalf("got = %#v", got)
	}
}

func TestParseGetInfoEmptyObject(t *testing.T) {
	t.Parallel()

	got, err := client.ParseGetInfo(jsonFixture(t, jsonGetInfoEmpty))
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}

	if got != nil {
		t.Fatalf("got = %#v", got)
	}
}

func TestParseGetInfoDetailShapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "null", body: `{"detail":null}`},
		{name: "array", body: `{"detail":[]}`},
		{name: "string", body: `{"detail":"alice"}`},
		{name: "number", body: `{"detail":1}`},
		{name: "empty-detail", body: `{"detail":{}}`},
		{name: "bad-member-id", body: `{"detail":{"member_id":{},"member_name":"alice"}}`},
		{name: "not-json", body: `not-json`},
		{name: "array-top", body: `[]`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := client.ParseGetInfo([]byte(test.body))
			if !errors.Is(err, client.ErrNotFound) {
				t.Fatalf("err = %v", err)
			}

			if got != nil {
				t.Fatalf("got = %#v", got)
			}
		})
	}
}

func TestParseGetInfoMinNoOptional(t *testing.T) {
	t.Parallel()

	body := []byte(`{"detail":{"member_id":1,"member_name":"bob","rank":"Lv.9"}}`)

	got, err := client.ParseGetInfo(body)
	if err != nil {
		t.Fatal(err)
	}

	want := &client.UserInfo{ID: 1, Name: nameBob}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}
