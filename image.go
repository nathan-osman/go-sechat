package sechat

import (
	"errors"
	"io"
)

// Image uploads an image and returns its new URL.
func (c *Conn) Image(r io.Reader) (string, error) {
	res, err := c.upload(
		"https://chat.stackexchange.com/upload/image",
		"filename",
		"untitled",
		r,
	)
	if err != nil {
		return "", err
	}
	program, err := c.parseJavaScript(res)
	if err != nil {
		return "", err
	}
	var (
		upErr string
		upURL string
	)
	for _, stm := range program.Body {
		asns, ok := c.parseAssignments(stm)
		if !ok {
			continue
		}
		for _, a := range asns {
			switch a.Name {
			case "error":
				upErr, _ = a.Value.(string)
			case "result":
				upURL, _ = a.Value.(string)
			}
		}
	}
	if len(upErr) != 0 {
		return "", errors.New(upErr)
	}
	return upURL, nil
}
