package sechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
)

// connectWebSocket attempts to establish the websocket connection to the chat
// server and return the new connection.
func (c *Conn) connectWebSocket() error {
	res, err := c.postForm(
		"/ws-auth",
		&url.Values{"roomid": {strconv.Itoa(c.room)}},
	)
	if err != nil {
		return err
	}
	var v struct {
		URL string `json:"url"`
	}
	if json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}
	// A custom dialer is used so that cookies are included
	dialer := &websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
		Jar:   c.client.Jar,
	}
	conn, _, err := dialer.Dial(
		fmt.Sprintf("%s?l=999999999999", v.URL),
		http.Header{"Origin": {"https://chat.stackexchange.com"}},
	)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	c.conn = conn
	c.mutex.Unlock()
	return nil
}
