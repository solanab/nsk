package client

func (c *Client) warmup() error {
	body, err := c.fetchHome()
	if err != nil {
		return err
	}

	user := parseUser(body)
	if user == nil {
		return ErrExpiredCookie
	}

	c.me = user

	return nil
}
