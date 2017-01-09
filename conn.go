package sechat

import (
	"fmt"
)

// Conn respresents a connection to the Stack Exchange chat network suitable
// for sending requests. This includes posting messages, starring messages,
// etc.
type Conn struct {
	auth *Auth
}

// NewConn creates a new connection from the provided authentication object.
func NewConn(auth *Auth) *Conn {
	return &Conn{
		auth: auth,
	}
}

// Send posts the specified message to the specified room. The ID of the new
// message is returned.
func (c *Conn) Send(room int, text string) error {
	err := c.auth.post(
		fmt.Sprintf("/chats/%d/messages/new", room),
		paramMap{"text": text},
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}
