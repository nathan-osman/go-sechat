package sechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
)

// wsAuthReply is returned by the server when obtaining the websocket URL.
type wsAuthReply struct {
	URL string `json:"url"`
}

// wsRoom represents data from a specific chat room.
type wsRoom struct {
	Events []*Event `json:"e"`
}

// run listens for new events and sends them on the channel.
func (c *Conn) run() {
	for {
		_, r, err := c.conn.NextReader()
		if err != nil {
			break
		}
		msg := map[string]json.RawMessage{}
		if err := json.NewDecoder(r).Decode(&msg); err != nil {
			continue
		}
		// Create a set of message IDs to avoid sending duplicates
		msgIDs := map[int]struct{}{}
		for _, v := range msg {
			room := &wsRoom{}
			if err := json.Unmarshal(v, &room); err != nil {
				continue
			}
			for _, e := range room.Events {
				if _, exists := msgIDs[e.ID]; !exists {
					c.Events <- e
					msgIDs[e.ID] = struct{}{}
				}
			}
		}
	}
	close(c.Events)
	close(c.closed)
}

// connectWebSocket attempts to establish the websocket connection to the chat
// server. An event loop is created in a separate goroutine to feed new events
// into the event channel.
func (c *Conn) connectWebSocket() error {
	res, err := c.postForm(
		"/ws-auth",
		&url.Values{"roomid": {strconv.Itoa(c.room)}},
	)
	if err != nil {
		return err
	}
	wsa := wsAuthReply{}
	if json.NewDecoder(res.Body).Decode(&wsa); err != nil {
		return err
	}
	// A custom dialer is used so that cookies are included
	dialer := &websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
		Jar:   c.client.Jar,
	}
	conn, _, err := dialer.Dial(
		fmt.Sprintf("%s?l=999999999999", wsa.URL),
		http.Header{"Origin": {"https://chat.stackexchange.com"}},
	)
	if err != nil {
		return err
	}
	c.conn = conn
	go c.run()
	return nil
}
