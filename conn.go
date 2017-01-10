package sechat

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"

	"github.com/gorilla/websocket"
)

const (
	AccessReadWrite = "read-write"
	AccessReadOnly  = "read-only"
	AccessRequest   = "request"
)

var ErrRoomID = errors.New("unable to find room ID")

// roomRegexp matches a "room" URL.
var roomRegexp = regexp.MustCompile(`^(?:https?://chat.stackexchange.com)?/rooms(?:/info)?/(\d+)`)

// Conn represents a connection to the Stack Exchange chat network. HTTP
// requests are used to trigger actions and websockets are used for event
// notifications.
type Conn struct {
	Events   chan *Event
	closed   chan bool
	client   *http.Client
	conn     *websocket.Conn
	email    string
	password string
	fkey     string
	room     int
}

// New creates a new Conn instance in the disconnected state.
func New(email, password string, room int) (*Conn, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &Conn{
		Events: make(chan *Event),
		closed: make(chan bool),
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if roomRegexp.MatchString(req.URL.String()) {
					return http.ErrUseLastResponse
				}
				return nil
			},
			Jar: jar,
		},
		email:    email,
		password: password,
		room:     room,
	}
	return c, nil
}

// Connect establishes a connection with the chat server.
func (c *Conn) Connect() error {
	if err := c.auth(); err != nil {
		return err
	}
	if err := c.connectWebSocket(); err != nil {
		return err
	}
	return nil
}

// Join listens for events in the specified room in addition to those already
// joined.
func (c *Conn) Join(room int) error {
	_, err := c.postForm(
		"/events",
		&url.Values{fmt.Sprintf("r%d", room): {"999999999999"}},
	)
	return err
}

// Leave stops listening for events in the specified room.
func (c *Conn) Leave(room int) error {
	_, err := c.postForm(
		fmt.Sprintf("/chats/leave/%d", room),
		&url.Values{"quiet": {"true"}},
	)
	return err
}

// newRoom eliminates the redundant code in NewRoom and NewRoomWithUser.
func (c *Conn) newRoom(path string, data *url.Values) (int, error) {
	res, err := c.postForm(path, data)
	if err != nil {
		return 0, err
	}
	m := roomRegexp.FindStringSubmatch(res.Header.Get("Location"))
	if m == nil {
		return 0, ErrRoomID
	}
	return c.atoi(m[1]), nil
}

// NewRoom creates a new room with the specified parameters. defaultAccess
// should normally be set to AccessReadWrite.
func (c *Conn) NewRoom(name, description, host, defaultAccess string) (int, error) {
	return c.newRoom(
		"/rooms/save",
		&url.Values{
			"name":          {name},
			"description":   {description},
			"host":          {host},
			"defaultAccess": {defaultAccess},
			"noDupeCheck":   {"true"},
		},
	)
}

// NewRoomWithUser creates a new room with the specified name and invites the
// specifed user to the new room. The ID of the new room is returned.
func (c *Conn) NewRoomWithUser(user int, name string) (int, error) {
	return c.newRoom(
		"/rooms/pairoff",
		&url.Values{
			"withUserId": {strconv.Itoa(user)},
			"name":       {name},
		},
	)
}

// Invite sends an invitation to a user inviting them to join a room.
func (c *Conn) Invite(user, room int) error {
	_, err := c.postForm(
		"/users/invite",
		&url.Values{
			"UserId": {strconv.Itoa(user)},
			"RoomId": {strconv.Itoa(room)},
		},
	)
	return err
}

// Send posts the specified message to the specified room.
func (c *Conn) Send(room int, text string) error {
	_, err := c.postForm(
		fmt.Sprintf("/chats/%d/messages/new", room),
		&url.Values{"text": {text}},
	)
	return err
}

// Star stars the specified message.
func (c *Conn) Star(message int) error {
	_, err := c.postForm(
		fmt.Sprintf("/messages/%d/star", message),
		&url.Values{},
	)
	return err
}

// Close disconnects the websocket and shuts down the connection.
func (c *Conn) Close() {
	c.conn.Close()
	<-c.closed
}
