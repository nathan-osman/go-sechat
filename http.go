package sechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// post is a utility method for sending a POST request with form data. The fkey
// is automatically added to the form data sent. If JSON is returned, the
// response can be read into a struct if the v parameter is not nil.
func (c *Conn) postForm(path string, data *url.Values, v interface{}) error {
	data.Set("fkey", c.fkey)
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://chat.stackexchange.com%s", path),
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	if v != nil {
		if err := json.NewDecoder(res.Body).Decode(v); err != nil {
			return err
		}
	}
	return nil
}
