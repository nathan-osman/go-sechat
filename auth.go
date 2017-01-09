package sechat

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrNetworkFkey = errors.New("unable to find network fkey")
	ErrChatFkey    = errors.New("unable to find chat fkey")
	ErrAuthURL     = errors.New("unable to find auth URL")
	ErrIncomplete  = errors.New("incomplete login (invalid credentials)")
)

// fetchLoginURL retrieves the URL of the page that contains the login form.
func (c *Conn) fetchLoginURL() (string, error) {
	req, err := http.NewRequest(
		http.MethodGet, "https://stackexchange.com/users/signin", nil,
	)
	if err != nil {
		return "", err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// fetchNetworkFkey retrieves the network fkey from the login form so that it
// can be submitted during the login process.
func (c *Conn) fetchNetworkFkey(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	fkey, ok := doc.Find("#fkey").Attr("value")
	if !ok {
		return "", ErrNetworkFkey
	}
	return fkey, nil
}

// fetchAuthURL exists due to a bug in golang.org/x/net/html that prevents
// noscript elements from being parsed correctly. This method works around it.
func (c *Conn) fetchAuthURL(res *http.Response) (string, error) {
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", err
	}
	noscriptDoc, err := goquery.NewDocumentFromReader(
		strings.NewReader(doc.Find("noscript").Text()),
	)
	if err != nil {
		return "", err
	}
	url, ok := noscriptDoc.Find("a").Attr("href")
	if !ok {
		return "", ErrAuthURL
	}
	return url, nil
}

// submitLoginForm submits the authentication information along with the fkey.
// A URL is returned which is necessary to complete the login process.
func (c *Conn) submitLoginForm(fkey string) (string, error) {
	form := &url.Values{}
	form.Set("email", c.email)
	form.Set("password", c.password)
	form.Set("affId", "11")
	form.Set("fkey", fkey)
	req, err := http.NewRequest(
		http.MethodPost,
		"https://openid.stackexchange.com/affiliate/form/login/submit",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	// fetchAuthURL ought to be part of this method but please see its
	// docstring for an explanation of why it isn't
	return c.fetchAuthURL(res)
}

// completeLogin finishes the login process.
func (c *Conn) completeLogin(authUrl string) error {
	req, err := http.NewRequest(http.MethodGet, authUrl, nil)
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.Request.URL.Path != "/" {
		return ErrIncomplete
	}
	return nil
}

// fetchChatFkey loads the home page for chat in order to retrieve the fkey
// that is required to accompany every authenticated request.
func (c *Conn) fetchChatFkey() (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://chat.stackexchange.com",
		nil,
	)
	if err != nil {
		return "", err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	fkey, ok := doc.Find("#fkey").Attr("value")
	if !ok {
		return "", ErrChatFkey
	}
	return fkey, nil
}

// auth performs the steps necessary to authenticate against the chat server.
func (c *Conn) auth() error {
	loginURL, err := c.fetchLoginURL()
	if err != nil {
		return err
	}
	networkFkey, err := c.fetchNetworkFkey(loginURL)
	if err != nil {
		return err
	}
	authURL, err := c.submitLoginForm(networkFkey)
	if err != nil {
		return err
	}
	if err := c.completeLogin(authURL); err != nil {
		return err
	}
	chatFkey, err := c.fetchChatFkey()
	if err != nil {
		return err
	}
	c.fkey = chatFkey
	return nil
}
