package sechat

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var conflictRegexp = regexp.MustCompile(`\d+`)

// postForm is a utility method for sending a POST request with form data. The
// fkey is automatically added to the form data sent. If a 409 Conflict
// response is received, the request is retried after the specified amount of
// time (to work around any throttle). Consequently, this method is blocking.
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
	for {
		res, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode >= 400 {
			// For HTTP 409, wait for the throttle to cool down
			if res.StatusCode == http.StatusConflict {
				b, err := ioutil.ReadAll(res.Body)
				if err == nil {
					m := conflictRegexp.FindStringSubmatch(string(b))
					if len(m) != 0 {
						i, _ := strconv.Atoi(m[0])
						time.Sleep(time.Duration(i) * time.Second)
						continue
					}
				}
			}
			return nil, errors.New(res.Status)
		}
		return res, nil
	}
}
