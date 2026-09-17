package client_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/solanab/nsk/internal/client"
)

const (
	jsonNotifyOK    = "notify-ok.json"
	jsonNotifyEmpty = "notify-empty.json"
	jsonNotifyGuest = "notify-guest.json"
	notifyTitle     = "奥特曼咋还不重置啊，出了这么大封号乌龙事件"
	notifyURL       = "https://www.nodeseek.com/post-763505-2#11"
)

func sampleNotify() []client.Notification {
	return []client.Notification{
		{
			PostID: 763505,
			Page:   2,
			Floor:  11,
			URL:    notifyURL,
			Text:   notifyTitle,
		},
	}
}

func TestParseNotificationsOK(t *testing.T) {
	t.Parallel()

	got, err := client.ParseNotifications(jsonFixture(t, jsonNotifyOK))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(sampleNotify(), got); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseNotificationsEmpty(t *testing.T) {
	t.Parallel()

	got, err := client.ParseNotifications(jsonFixture(t, jsonNotifyEmpty))
	if err != nil {
		t.Fatal(err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("got = %#v", got)
	}
}

func TestParseNotificationsGuest(t *testing.T) {
	t.Parallel()

	got, err := client.ParseNotifications(jsonFixture(t, jsonNotifyGuest))
	if !errors.Is(err, client.ErrAuthRequired) {
		t.Fatalf("err = %v", err)
	}

	if got != nil {
		t.Fatalf("got = %#v", got)
	}
}

func TestParseNotificationsShapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		err  error
	}{
		{name: "invalid-json", body: bodyNotJSON, err: client.ErrNotFound},
		{name: "array-top", body: `[]`, err: client.ErrNotFound},
		{name: "success-0", body: `{"success":0,"data":[]}`, err: client.ErrAuthRequired},
		{name: "success-false", body: `{"success":false}`, err: client.ErrAuthRequired},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := client.ParseNotifications([]byte(test.body))
			if !errors.Is(err, test.err) {
				t.Fatalf("err = %v", err)
			}

			if got != nil {
				t.Fatalf("got = %#v", got)
			}
		})
	}
}

func TestParseNotificationsSuccessNumber(t *testing.T) {
	t.Parallel()

	body := []byte(`{"success":1,"data":[{"post_id":7,"floor_id":0,"title":"op"}]}`)

	got, err := client.ParseNotifications(body)
	if err != nil {
		t.Fatal(err)
	}

	want := []client.Notification{
		{
			PostID: 7,
			Page:   1,
			Floor:  0,
			URL:    "https://www.nodeseek.com/post-7-1#0",
			Text:   "op",
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseNotificationsSkipsInvalid(t *testing.T) {
	t.Parallel()

	body := []byte(`{"success":true,"data":[` +
		`{"post_id":0,"floor_id":1,"title":"x"},` +
		`{"post_id":9,"floor_id":-1,"title":"y"},` +
		`{"post_id":8,"floor_id":11,"title":"ok"}]}`)

	got, err := client.ParseNotifications(body)
	if err != nil {
		t.Fatal(err)
	}

	want := []client.Notification{
		{
			PostID: 8,
			Page:   2,
			Floor:  11,
			URL:    "https://www.nodeseek.com/post-8-2#11",
			Text:   "ok",
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestParseNotificationsMissingData(t *testing.T) {
	t.Parallel()

	got, err := client.ParseNotifications([]byte(`{"success":true}`))
	if err != nil {
		t.Fatal(err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("got = %#v", got)
	}
}

func TestParseNotificationsMissingSuccess(t *testing.T) {
	t.Parallel()

	got, err := client.ParseNotifications([]byte(`{"data":[]}`))
	if err != nil {
		t.Fatal(err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("got = %#v", got)
	}
}
