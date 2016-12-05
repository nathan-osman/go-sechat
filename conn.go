package sechat

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type paramMap map[string]string

// Conn respresents a connection to the Stack Exchange chat network suitable
// for sending requests. This includes posting messages, starring messages,
// etc.
type Conn struct {
	auth *Auth
}

// post eliminates a lot of common code used for sending POST requests.
func (c *Conn) post(path string, params paramMap) error {
	form := &url.Values{}
	form.Set("fkey", c.auth.state.Fkey)
	for key, value := range params {
		form.Set(key, value)
	}
	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://chat.stackexchange.com%s", path),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.auth.client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(response.Status)
	}
	return nil
}

// NewConn creates a new connection from the provided authentication object.
func NewConn(auth *Auth) *Conn {
	return &Conn{
		auth: auth,
	}
}

// Send posts the specified message to the specified room.
func (c *Conn) Send(room int, text string) error {
	return c.post(
		fmt.Sprintf("/chats/%d/messages/new", room),
		paramMap{"text": text},
	)
}
