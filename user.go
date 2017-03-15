package sechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrInvalidUserResponse = errors.New("invalid user response")
	ErrInvalidJavaScript   = errors.New("invalid JavaScript (no access to room)")
)

// Room provides information about a room that a user is currently present in.
type Room struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LastPost int    `json:"last_post"`
	Activity int    `json:"activity"`
}

// Site describes a user's parent site.
type Site struct {
	Icon    string `json:"icon"`
	Caption string `json:"caption"`
}

// InviteTarget describes a site that another user may be invited to.
type InviteTarget struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// User provides information about a chat user, such as their display name,
// moderator status, etc. Not every method that returns an instance of this
// type populates all fields.
type User struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	EmailHash    string  `json:"email_hash"`
	Reputation   int     `json:"reputation"`
	IsModerator  bool    `json:"is_moderator"`
	IsOwner      bool    `json:"is_owner"`
	IsRegistered bool    `json:"is_registered"`
	LastPost     int     `json:"last_post"`
	LastSeen     int     `json:"last_seen"`
	Rooms        []*Room `json:"rooms"`
	UserMessage  string  `json:"user_message"`
	ProfileURL   string  `json:"profileUrl"`
	Site         *Site   `json:"site"`
	Host         string  `json:"host"`
	MayPairoff   bool    `json:"may_pairoff"`
	Issues       int     `json:"issues"`
}

// mapToUser creates a user from an astMap. Interestingly, the easiest way to
// do this is encode the map and then decode it as JSON.
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

// User retrieves extended information for a specific user.
func (c *Conn) User(user int) (*User, error) {
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
	u := &User{}
	if err := json.NewDecoder(res.Body).Decode(u); err != nil {
		return nil, err
	}
	return u, nil
}

// Users retrieves limited information for the specified users. Only the first
// few fields in the User struct are filled in.
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

// UsersInRoom retrieves a list of users in the specified room. Only the first
// few fields in the User struct are filled in.
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
