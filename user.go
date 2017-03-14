package sechat

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// User provides information about a chat user, such as their display name,
// moderator status, etc.
type User struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	EmailHash   string `json:"email_hash"`
	Reputation  int    `json:"reputation"`
	IsModerator bool   `json:"is_moderator"`
	IsOwner     bool   `json:"is_owner"`
	LastPost    int    `json:"last_post"`
	LastSeen    int    `json:"last_seen"`
}

// User retrieves information for a specific user.
func (c *Conn) User(user int, room int) ([]*User, error) {
	res, err := c.postForm(
		"/user/info",
		&url.Values{
			"ids":    {strconv.Itoa(user)},
			"roomid": {strconv.Itoa(room)},
		},
	)
	if err != nil {
		return nil, err
	}
	var v struct {
		Users []*User `json:"users"`
	}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v.Users, nil
}
