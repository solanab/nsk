package client

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type getInfoResponse struct {
	Detail json.RawMessage `json:"detail"`
}

type getInfoDetail struct {
	MemberID   int    `json:"member_id"`   //nolint:tagliatelle // official getInfo field
	MemberName string `json:"member_name"` //nolint:tagliatelle // official getInfo field
	Chicken    int    `json:"chicken"`
	Level      string `json:"level"`
}

// GetUser fetches one user by numeric id. Non-positive ids do not hit HTTP.
func (c *Client) GetUser(userID int) (*UserInfo, error) {
	if userID <= 0 {
		return nil, errInvalidUserID
	}

	body, err := c.fetch(userInfoURL(userID), "请求用户")
	if err != nil {
		return nil, err
	}

	return parseGetInfo(body)
}

func userInfoURL(userID int) string {
	return Site + "/api/account/getInfo/" + strconv.Itoa(userID)
}

func parseGetInfo(body []byte) (*UserInfo, error) {
	var top getInfoResponse
	if err := json.Unmarshal(body, &top); err != nil {
		return nil, ErrNotFound
	}

	detail := bytes.TrimSpace(top.Detail)
	if len(detail) == 0 || detail[0] != '{' {
		return nil, ErrNotFound
	}

	var raw getInfoDetail
	if err := json.Unmarshal(detail, &raw); err != nil {
		return nil, ErrNotFound
	}

	info := projectUser(&userJSON{
		ID:      raw.MemberID,
		Name:    raw.MemberName,
		Chicken: raw.Chicken,
		Level:   raw.Level,
	})
	if info == nil {
		return nil, ErrNotFound
	}

	return info, nil
}
