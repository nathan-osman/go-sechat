package sechat

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/gorilla/websocket"
)

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
	return &Conn{
		Events: make(chan *Event),
		closed: make(chan bool),
		client: &http.Client{
			Jar: jar,
		},
		email:    email,
		password: password,
		room:     room,
	}, nil
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
	return c.postForm(
		"/events",
		&url.Values{fmt.Sprintf("r%d", room): {"999999999999"}},
		nil,
	)
}

// Leave stops listening for events in the specified room.
func (c *Conn) Leave(room int) error {
	return c.postForm(
		fmt.Sprintf("/chats/leave/%d", room),
		&url.Values{"quiet": {"true"}},
		nil,
	)
}

// Send posts the specified message to the specified room.
func (c *Conn) Send(room int, text string) error {
	return c.postForm(
		fmt.Sprintf("/chats/%d/messages/new", room),
		&url.Values{"text": {text}},
		nil,
	)
}

// Close disconnects the websocket and shuts down the connection.
func (c *Conn) Close() {
	c.conn.Close()
	<-c.closed
}
