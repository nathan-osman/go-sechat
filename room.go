package sechat

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Room provides information about a specific room.
type Room struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LastPost int    `json:"last_post"`
	Activity int    `json:"activity"`
}

// Rooms retrieves a list of all rooms the specified user is in.
func (c *Conn) Rooms(user int) ([]*Room, error) {
	req, err := c.newRequest(
		http.MethodGet,
		fmt.Sprintf("https://chat.stackexchange.com/users/thumbs/%d", user),
		nil,
	)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	var v struct {
		Rooms []*Room `json:"rooms"`
	}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v.Rooms, nil
}
