package client

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type notifyResponse struct {
	Success json.RawMessage `json:"success"`
	Data    []notifyItem    `json:"data"`
}

type notifyItem struct {
	PostID int    `json:"post_id"`  //nolint:tagliatelle // official at-me field
	Floor  int    `json:"floor_id"` //nolint:tagliatelle // official at-me field
	Title  string `json:"title"`
}

// Notifications fetches the current Account's at-me list. There is no page parameter.
func (c *Client) Notifications() ([]Notification, error) {
	body, err := c.fetch(notifyListURL(), "请求通知")
	if err != nil {
		return nil, err
	}

	return parseNotifications(body)
}

func notifyListURL() string {
	return Site + "/api/notification/at-me/list"
}

func parseNotifications(body []byte) ([]Notification, error) {
	var top notifyResponse
	if err := json.Unmarshal(body, &top); err != nil {
		return nil, ErrNotFound
	}

	if !notifySuccess(top.Success) {
		return nil, ErrAuthRequired
	}

	out := make([]Notification, 0, len(top.Data))
	for _, item := range top.Data {
		if note, ok := projectNotify(item); ok {
			out = append(out, note)
		}
	}

	return out, nil
}

func projectNotify(item notifyItem) (Notification, bool) {
	if item.PostID <= 0 || item.Floor < 0 {
		return Notification{}, false
	}

	page := PageForFloor(item.Floor)

	return Notification{
		PostID: item.PostID,
		Page:   page,
		Floor:  item.Floor,
		URL:    postURL(item.PostID, page) + "#" + strconv.Itoa(item.Floor),
		Text:   item.Title,
	}, true
}

func notifySuccess(raw json.RawMessage) bool {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return true
	}

	return bytes.Equal(raw, []byte("true")) || bytes.Equal(raw, []byte("1"))
}
