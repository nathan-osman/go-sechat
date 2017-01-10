package sechat

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// atoi removes the error handling from Atoi() and ensures a value is always
// returned.
func (c *Conn) atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// postForm is a utility method for sending a POST request with form data. The
// fkey is automatically added to the form data sent.
func (c *Conn) postForm(path string, data *url.Values) (*http.Response, error) {
	data.Set("fkey", c.fkey)
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://chat.stackexchange.com%s", path),
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, errors.New(res.Status)
	}
	return res, nil
}
