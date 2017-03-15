package sechat

import (
	"encoding/json"
	"time"
)

// wsRoom represents data from a specific chat room.
type wsRoom struct {
	Events []*Event `json:"e"`
}

// run is the main loop. It continually connects to the chat server, sleeping,
// and reconnecting upon error. It runs continually until stopped.
func (c *Conn) run(ch chan<- *Event) {
	defer close(c.closedCh)
	defer close(c.connectedCh)
	defer close(ch)
	defer c.log.Info("closing event channel")
	for {
		// Use the stored credentials to authenticate
		if err := c.auth(); err != nil {
			c.log.Error(err)
			goto retry
		}
		// Connect to to the websocket server
		if err := c.connectWebSocket(); err != nil {
			c.log.Error(err)
			goto retry
		}
		c.log.Info("connected to WebSocket")
		select {
		case c.connectedCh <- true:
		default:
		}
		// Event receiving loop
	loop:
		for {
			_, r, err := c.conn.NextReader()
			if err != nil {
				// Check to see if the error was caused by a shutdown or if it
				// is an actual error (in which case, leave the loop)
				select {
				case <-c.closeCh:
					return
				default:
					c.log.Error(err)
					break loop
				}
			}
			// Partially decode the message
			msg := map[string]json.RawMessage{}
			if err := json.NewDecoder(r).Decode(&msg); err != nil {
				continue
			}
			// Use a "set" to prevent duplicate events from being sent
			msgIDs := map[int]struct{}{}
			for _, v := range msg {
				room := &wsRoom{}
				if err := json.Unmarshal(v, &room); err != nil {
					continue
				}
				for _, e := range room.Events {
					if _, exists := msgIDs[e.ID]; !exists {
						e.precompute()
						ch <- e
						msgIDs[e.ID] = struct{}{}
					}
				}
			}
		}
	retry:
		select {
		case c.connectedCh <- false:
		default:
		}
		if _, ok := <-c.closeCh; !ok {
			return
		}
		c.log.Info("reconnecting in 30 seconds")
		select {
		case <-time.After(30 * time.Second):
		case <-c.closeCh:
			return
		}
	}
}
