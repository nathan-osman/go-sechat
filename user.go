package sechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

var ErrInvalidJavaScript = errors.New("invalid JavaScript (no access to room)")

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

// mapToUser creates a user from an astMap. Interestingly, the fastest way to
// do this is encode and then decode JSON.
func mapToUser(m astMap) (*User, error) {
	buff := &bytes.Buffer{}
	if err := json.NewEncoder(buff).Encode(m); err != nil {
		return nil, err
	}
	user := &User{}
	if err := json.NewDecoder(buff).Decode(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Users retrieves information for the specified users.
func (c *Conn) Users(users []int, room int) ([]*User, error) {
	usersStr := make([]string, len(users))
	for i, user := range users {
		usersStr[i] = strconv.Itoa(user)
	}
	res, err := c.postForm(
		"/user/info",
		&url.Values{
			"ids":    {strings.Join(usersStr, ", ")},
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

// UsersInRoom retrieves a list of users in the specified room.
func (c *Conn) UsersInRoom(room int) ([]*User, error) {
	program, err := c.parseJavaScript(
		fmt.Sprintf("https://chat.stackexchange.com/rooms/%d", room),
	)
	if err != nil {
		return nil, err
	}
	statements := c.findOnReadyStatements(program)
	for _, stm := range statements {
		call := c.parseFunctionCall(stm)
		if call == nil {
			continue
		}
		if call.Name != "CHAT.RoomUsers.initPresent" {
			continue
		}
		if len(call.Arguments) != 1 {
			return nil, ErrInvalidJavaScript
		}
		vals, ok := c.parseArray(call.Arguments[0])
		if !ok {
			return nil, ErrInvalidJavaScript
		}
		users := []*User{}
		for _, exp := range vals {
			user, err := mapToUser(c.parseMap(exp))
			if err != nil {
				return nil, err
			}
			users = append(users, user)
		}
		return users, nil
	}
	return nil, ErrInvalidJavaScript
}
