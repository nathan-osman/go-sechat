package sechat

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var conflictRegexp = regexp.MustCompile(`\d+`)

// newRequest wraps http.NewRequest, logging the request and allowing the user
// agent to be customized.
func (c *Conn) newRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	c.log.Debugf("%s: %s", method, urlStr)
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "go-sechat (https://qms.li/gsc)")
	return req, nil
}

// postForm is a utility method for sending a POST request with form data. The
// fkey is automatically added to the form data sent. If a 409 Conflict
// response is received, the request is retried after the specified amount of
// time (to work around any throttle). Consequently, this method is blocking.
func (c *Conn) postForm(path string, data *url.Values) (*http.Response, error) {
	data.Set("fkey", c.fkey)
	for {
		req, err := c.newRequest(
			http.MethodPost,
			fmt.Sprintf("https://chat.stackexchange.com%s", path),
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
		// For HTTP 409, wait for the throttle to cool down
		if res.StatusCode == http.StatusConflict {
			b, err := ioutil.ReadAll(res.Body)
			if err == nil {
				m := conflictRegexp.FindStringSubmatch(string(b))
				if len(m) != 0 {
					i, _ := strconv.Atoi(m[0])
					c.log.Infof("retrying %s in %d second(s)", path, i)
					select {
					case <-c.closeCh:
					case <-time.After(time.Duration(i) * time.Second):
						continue
					}
				}
			}
		}
		if res.StatusCode >= 400 {
			return nil, errors.New(res.Status)
		}
		return res, nil
	}
}
